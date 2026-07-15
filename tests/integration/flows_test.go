package integration

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"barterswap/internal/config"
	"barterswap/internal/database"
	"barterswap/internal/domain"
	"barterswap/internal/repository"
	"barterswap/internal/service"
)

func setupIntegration(t *testing.T) (*repository.Store, func()) {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = config.DefaultDatabaseURL
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Open(ctx, dsn)
	if err != nil {
		t.Skipf("base de données indisponible : %v", err)
	}
	if err := database.Migrate(ctx, db); err != nil {
		db.Close()
		t.Fatalf("migration : %v", err)
	}
	if err := truncateAll(ctx, db); err != nil {
		db.Close()
		t.Fatalf("truncate : %v", err)
	}
	return repository.New(db), func() {
		_ = truncateAll(context.Background(), db)
		db.Close()
	}
}

func truncateAll(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		TRUNCATE reviews, credit_transactions, exchanges, services, skills, users
		RESTART IDENTITY CASCADE`)
	return err
}

func TestIntegrationUserFlow(t *testing.T) {
	store, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()
	svc := service.NewUserService(store)

	u, err := svc.Create(ctx, service.CreateUserInput{Pseudo: "alice", Ville: "Paris"})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if u.CreditBalance != domain.WelcomeCredits {
		t.Fatalf("welcome credits = %d", u.CreditBalance)
	}

	if err := svc.ReplaceSkills(ctx, u.ID, u.ID, []domain.Skill{{Nom: "Jardinage", Niveau: "expert"}}); err != nil {
		t.Fatalf("replace skills: %v", err)
	}
	skills, err := svc.GetSkills(ctx, u.ID)
	if err != nil || len(skills) != 1 {
		t.Fatalf("get skills: %v, len=%d", err, len(skills))
	}

	updated, err := svc.Update(ctx, u.ID, u.ID, service.UpdateUserInput{Pseudo: "alice2", Bio: "bio"})
	if err != nil || updated.Pseudo != "alice2" {
		t.Fatalf("update user: %v, %+v", err, updated)
	}
}

func TestIntegrationServiceAndExchangeFlow(t *testing.T) {
	store, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()

	userSvc := service.NewUserService(store)
	serviceSvc := service.NewServiceService(store, userSvc)
	exchangeSvc := service.NewExchangeService(store)

	alice, _ := userSvc.Create(ctx, service.CreateUserInput{Pseudo: "alice"})
	bob, _ := userSvc.Create(ctx, service.CreateUserInput{Pseudo: "bob"})
	_ = userSvc.ReplaceSkills(ctx, alice.ID, alice.ID, []domain.Skill{{Nom: "Jardinage", Niveau: "expert"}})

	svc, err := serviceSvc.Create(ctx, alice.ID, service.CreateServiceInput{
		Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Ville: "Paris",
	})
	if err != nil {
		t.Fatalf("create service: %v", err)
	}

	services, err := serviceSvc.List(ctx, service.ListServicesInput{Categorie: "Jardinage", Ville: "Paris", Search: "Tont"})
	if err != nil || len(services) != 1 {
		t.Fatalf("list services: %v, count=%d", err, len(services))
	}

	ex, err := exchangeSvc.Create(ctx, bob.ID, service.CreateExchangeInput{ServiceID: svc.ID})
	if err != nil {
		t.Fatalf("create exchange: %v", err)
	}
	if _, err := exchangeSvc.Create(ctx, bob.ID, service.CreateExchangeInput{ServiceID: svc.ID}); err == nil {
		t.Fatal("expected conflict on second exchange")
	}

	ex, err = exchangeSvc.Accept(ctx, alice.ID, ex.ID)
	if err != nil || ex.Status != domain.StatusAccepted {
		t.Fatalf("accept: %v, status=%s", err, ex.Status)
	}

	bobUser, _ := userSvc.GetByID(ctx, bob.ID)
	if bobUser.CreditBalance != 8 {
		t.Fatalf("bob balance after accept = %d", bobUser.CreditBalance)
	}

	ex, err = exchangeSvc.Complete(ctx, alice.ID, ex.ID)
	if err != nil || ex.Status != domain.StatusCompleted {
		t.Fatalf("complete: %v", err)
	}

	aliceUser, _ := userSvc.GetByID(ctx, alice.ID)
	if aliceUser.CreditBalance != 12 {
		t.Fatalf("alice balance after complete = %d", aliceUser.CreditBalance)
	}
}

func TestIntegrationReviewAndStats(t *testing.T) {
	store, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()

	userSvc := service.NewUserService(store)
	serviceSvc := service.NewServiceService(store, userSvc)
	exchangeSvc := service.NewExchangeService(store)
	reviewSvc := service.NewReviewService(store)

	alice, _ := userSvc.Create(ctx, service.CreateUserInput{Pseudo: "alice"})
	bob, _ := userSvc.Create(ctx, service.CreateUserInput{Pseudo: "bob"})
	_ = userSvc.ReplaceSkills(ctx, alice.ID, alice.ID, []domain.Skill{{Nom: "Jardinage", Niveau: "expert"}})

	svc, _ := serviceSvc.Create(ctx, alice.ID, service.CreateServiceInput{
		Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2,
	})
	ex, _ := exchangeSvc.Create(ctx, bob.ID, service.CreateExchangeInput{ServiceID: svc.ID})
	_, _ = exchangeSvc.Accept(ctx, alice.ID, ex.ID)
	_, _ = exchangeSvc.Complete(ctx, alice.ID, ex.ID)

	if _, err := reviewSvc.Create(ctx, bob.ID, ex.ID, service.CreateReviewInput{Note: 5}); err != nil {
		t.Fatalf("create review: %v", err)
	}
	if _, err := reviewSvc.Create(ctx, bob.ID, ex.ID, service.CreateReviewInput{Note: 3}); err == nil {
		t.Fatal("expected duplicate review error")
	}

	reviews, err := reviewSvc.ListByUser(ctx, alice.ID)
	if err != nil || len(reviews) != 1 {
		t.Fatalf("list user reviews: %v, len=%d", err, len(reviews))
	}
	svcReviews, err := reviewSvc.ListByService(ctx, svc.ID)
	if err != nil || len(svcReviews) != 1 {
		t.Fatalf("list service reviews: %v, len=%d", err, len(svcReviews))
	}

	stats, err := userSvc.Stats(ctx, alice.ID)
	if err != nil {
		t.Fatalf("stats: %v", err)
	}
	if stats.CreditBalance != 12 || stats.EchangesCompletes != 1 || stats.NbAvis != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}

func TestIntegrationCancelRefundsCredits(t *testing.T) {
	store, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()

	userSvc := service.NewUserService(store)
	serviceSvc := service.NewServiceService(store, userSvc)
	exchangeSvc := service.NewExchangeService(store)

	alice, _ := userSvc.Create(ctx, service.CreateUserInput{Pseudo: "alice"})
	bob, _ := userSvc.Create(ctx, service.CreateUserInput{Pseudo: "bob"})
	_ = userSvc.ReplaceSkills(ctx, alice.ID, alice.ID, []domain.Skill{{Nom: "Jardinage", Niveau: "expert"}})
	svc, _ := serviceSvc.Create(ctx, alice.ID, service.CreateServiceInput{
		Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2,
	})
	ex, _ := exchangeSvc.Create(ctx, bob.ID, service.CreateExchangeInput{ServiceID: svc.ID})
	_, _ = exchangeSvc.Accept(ctx, alice.ID, ex.ID)
	_, err := exchangeSvc.Cancel(ctx, bob.ID, ex.ID)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	bobUser, _ := userSvc.GetByID(ctx, bob.ID)
	if bobUser.CreditBalance != 10 {
		t.Fatalf("bob balance after cancel = %d, want 10", bobUser.CreditBalance)
	}
}

func TestIntegrationRejectExchange(t *testing.T) {
	store, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()

	userSvc := service.NewUserService(store)
	serviceSvc := service.NewServiceService(store, userSvc)
	exchangeSvc := service.NewExchangeService(store)

	alice, _ := userSvc.Create(ctx, service.CreateUserInput{Pseudo: "alice"})
	bob, _ := userSvc.Create(ctx, service.CreateUserInput{Pseudo: "bob"})
	_ = userSvc.ReplaceSkills(ctx, alice.ID, alice.ID, []domain.Skill{{Nom: "Jardinage", Niveau: "expert"}})
	svc, _ := serviceSvc.Create(ctx, alice.ID, service.CreateServiceInput{
		Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2,
	})
	ex, _ := exchangeSvc.Create(ctx, bob.ID, service.CreateExchangeInput{ServiceID: svc.ID})
	ex, err := exchangeSvc.Reject(ctx, alice.ID, ex.ID)
	if err != nil || ex.Status != domain.StatusRejected {
		t.Fatalf("reject: %v, status=%s", err, ex.Status)
	}
}

func TestIntegrationListExchanges(t *testing.T) {
	store, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()

	userSvc := service.NewUserService(store)
	serviceSvc := service.NewServiceService(store, userSvc)
	exchangeSvc := service.NewExchangeService(store)

	alice, _ := userSvc.Create(ctx, service.CreateUserInput{Pseudo: "alice"})
	bob, _ := userSvc.Create(ctx, service.CreateUserInput{Pseudo: "bob"})
	_ = userSvc.ReplaceSkills(ctx, alice.ID, alice.ID, []domain.Skill{{Nom: "Jardinage", Niveau: "expert"}})
	svc, _ := serviceSvc.Create(ctx, alice.ID, service.CreateServiceInput{
		Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2,
	})
	ex, _ := exchangeSvc.Create(ctx, bob.ID, service.CreateExchangeInput{ServiceID: svc.ID})

	list, err := exchangeSvc.List(ctx, alice.ID, domain.StatusPending)
	if err != nil || len(list) != 1 || list[0].ID != ex.ID {
		t.Fatalf("list exchanges: %v, %+v", err, list)
	}
}

func TestIntegrationUpdateAndDeleteService(t *testing.T) {
	store, cleanup := setupIntegration(t)
	defer cleanup()
	ctx := context.Background()

	userSvc := service.NewUserService(store)
	serviceSvc := service.NewServiceService(store, userSvc)

	alice, _ := userSvc.Create(ctx, service.CreateUserInput{Pseudo: "alice"})
	_ = userSvc.ReplaceSkills(ctx, alice.ID, alice.ID, []domain.Skill{{Nom: "Jardinage", Niveau: "expert"}})
	svc, _ := serviceSvc.Create(ctx, alice.ID, service.CreateServiceInput{
		Titre: "Tonte", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Ville: "Paris",
	})

	updated, err := serviceSvc.Update(ctx, alice.ID, svc.ID, service.UpdateServiceInput{
		Titre: "Tonte+", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2, Ville: "Lyon", Actif: false,
	})
	if err != nil || updated.Titre != "Tonte+" || updated.Actif {
		t.Fatalf("update service: %v, %+v", err, updated)
	}

	if err := serviceSvc.Delete(ctx, alice.ID, svc.ID); err != nil {
		t.Fatalf("delete service: %v", err)
	}
}
