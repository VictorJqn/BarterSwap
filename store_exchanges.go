package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

func (s *sqlStore) CreateExchange(ctx context.Context, e Exchange) (Exchange, error) {
	var (
		createdAt time.Time
		updatedAt time.Time
	)
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO exchanges (service_id, requester_id, owner_id, status)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at, updated_at`,
		e.ServiceID, e.RequesterID, e.OwnerID, StatusPending,
	).Scan(&e.ID, &createdAt, &updatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return Exchange{}, fmt.Errorf("%w: service déjà réservé", ErrConflict)
		}
		return Exchange{}, err
	}
	e.Status = StatusPending
	e.CreatedAt = createdAt.Format(time.RFC3339)
	e.UpdatedAt = updatedAt.Format(time.RFC3339)
	return e, nil
}

func (s *sqlStore) GetExchangeByID(ctx context.Context, id int) (Exchange, error) {
	var (
		e         Exchange
		createdAt time.Time
		updatedAt time.Time
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT id, service_id, requester_id, owner_id, status, created_at, updated_at
		 FROM exchanges WHERE id = $1`, id,
	).Scan(&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID, &e.Status, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Exchange{}, fmt.Errorf("%w: échange %d", ErrNotFound, id)
	}
	if err != nil {
		return Exchange{}, err
	}
	e.CreatedAt = createdAt.Format(time.RFC3339)
	e.UpdatedAt = updatedAt.Format(time.RFC3339)
	return e, nil
}

func (s *sqlStore) ListExchanges(ctx context.Context, userID int, status string) ([]Exchange, error) {
	query := `SELECT id, service_id, requester_id, owner_id, status, created_at, updated_at
	          FROM exchanges
	          WHERE (requester_id = $1 OR owner_id = $1)`
	args := []any{userID}

	if status != "" {
		query += ` AND status = $2`
		args = append(args, status)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	exchanges := []Exchange{}
	for rows.Next() {
		var (
			e         Exchange
			createdAt time.Time
			updatedAt time.Time
		)
		if err := rows.Scan(&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID, &e.Status, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		e.CreatedAt = createdAt.Format(time.RFC3339)
		e.UpdatedAt = updatedAt.Format(time.RFC3339)
		exchanges = append(exchanges, e)
	}
	return exchanges, rows.Err()
}

func (s *sqlStore) AcceptExchange(ctx context.Context, exchangeID int, credits int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		status      string
		requesterID int
	)
	err = tx.QueryRowContext(ctx,
		`SELECT status, requester_id FROM exchanges WHERE id = $1 FOR UPDATE`, exchangeID,
	).Scan(&status, &requesterID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: échange %d", ErrNotFound, exchangeID)
	}
	if err != nil {
		return err
	}
	if status != StatusPending {
		return fmt.Errorf("%w: seul un échange en attente peut être accepté", ErrValidation)
	}

	balance, err := creditBalanceTx(ctx, tx, requesterID)
	if err != nil {
		return err
	}
	if balance < credits {
		return ErrInsufficientCredits
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO credit_transactions (user_id, exchange_id, montant, type)
		 VALUES ($1, $2, $3, $4)`,
		requesterID, exchangeID, -credits, TxSpend,
	); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx,
		`UPDATE exchanges SET status = $1, updated_at = now() WHERE id = $2 AND status = $3`,
		StatusAccepted, exchangeID, StatusPending,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("%w: transition impossible", ErrValidation)
	}
	return tx.Commit()
}

func (s *sqlStore) RejectExchange(ctx context.Context, exchangeID int) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE exchanges SET status = $1, updated_at = now() WHERE id = $2 AND status = $3`,
		StatusRejected, exchangeID, StatusPending,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		ex, err := s.GetExchangeByID(ctx, exchangeID)
		if err != nil {
			return err
		}
		if ex.Status != StatusPending {
			return fmt.Errorf("%w: seul un échange en attente peut être refusé", ErrValidation)
		}
		return fmt.Errorf("%w: échange %d", ErrNotFound, exchangeID)
	}
	return nil
}

func (s *sqlStore) CompleteExchange(ctx context.Context, exchangeID int, credits int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		status  string
		ownerID int
	)
	err = tx.QueryRowContext(ctx,
		`SELECT status, owner_id FROM exchanges WHERE id = $1 FOR UPDATE`, exchangeID,
	).Scan(&status, &ownerID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: échange %d", ErrNotFound, exchangeID)
	}
	if err != nil {
		return err
	}
	if status != StatusAccepted {
		return fmt.Errorf("%w: seul un échange accepté peut être terminé", ErrValidation)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO credit_transactions (user_id, exchange_id, montant, type)
		 VALUES ($1, $2, $3, $4)`,
		ownerID, exchangeID, credits, TxEarn,
	); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx,
		`UPDATE exchanges SET status = $1, updated_at = now() WHERE id = $2 AND status = $3`,
		StatusCompleted, exchangeID, StatusAccepted,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("%w: transition impossible", ErrValidation)
	}
	return tx.Commit()
}

func (s *sqlStore) CancelExchange(ctx context.Context, exchangeID int, credits int, wasAccepted bool) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		status      string
		requesterID int
	)
	err = tx.QueryRowContext(ctx,
		`SELECT status, requester_id FROM exchanges WHERE id = $1 FOR UPDATE`, exchangeID,
	).Scan(&status, &requesterID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: échange %d", ErrNotFound, exchangeID)
	}
	if err != nil {
		return err
	}
	if status != StatusPending && status != StatusAccepted {
		return fmt.Errorf("%w: cet échange ne peut plus être annulé", ErrValidation)
	}

	if wasAccepted {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO credit_transactions (user_id, exchange_id, montant, type)
			 VALUES ($1, $2, $3, $4)`,
			requesterID, exchangeID, credits, TxRefund,
		); err != nil {
			return err
		}
	}

	expectedStatus := status
	res, err := tx.ExecContext(ctx,
		`UPDATE exchanges SET status = $1, updated_at = now() WHERE id = $2 AND status = $3`,
		StatusCancelled, exchangeID, expectedStatus,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("%w: transition impossible", ErrValidation)
	}
	return tx.Commit()
}

func (s *sqlStore) GetCreditBalance(ctx context.Context, userID int) (int, error) {
	return creditBalanceTx(ctx, s.db, userID)
}

func creditBalanceTx(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, userID int) (int, error) {
	var balance int
	err := q.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(montant), 0) FROM credit_transactions WHERE user_id = $1`, userID,
	).Scan(&balance)
	return balance, err
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
