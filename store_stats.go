package main

import "context"

func (s *sqlStore) GetUserStats(ctx context.Context, userID int) (UserStats, error) {
	stats := UserStats{UserID: userID}
	err := s.db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM services WHERE provider_id = $1 AND actif),
			(SELECT COUNT(*) FROM exchanges
			 WHERE (requester_id = $1 OR owner_id = $1) AND status = 'completed'),
			(SELECT COALESCE(SUM(montant), 0) FROM credit_transactions WHERE user_id = $1),
			(SELECT COALESCE(ROUND(AVG(note)::numeric, 2), 0) FROM reviews WHERE target_id = $1),
			(SELECT COUNT(*) FROM reviews WHERE target_id = $1),
			(SELECT COALESCE(SUM(montant), 0) FROM credit_transactions
			 WHERE user_id = $1 AND type = 'earn' AND exchange_id IS NOT NULL),
			(SELECT COALESCE(-SUM(montant), 0) FROM credit_transactions
			 WHERE user_id = $1 AND type = 'spend')`,
		userID,
	).Scan(
		&stats.ServicesActifs,
		&stats.EchangesCompletes,
		&stats.CreditBalance,
		&stats.NoteMoyenne,
		&stats.NbAvis,
		&stats.TotalGagne,
		&stats.TotalDepense,
	)
	if err != nil {
		return UserStats{}, err
	}
	return stats, nil
}
