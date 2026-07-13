package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ikukhar/refuel-backend/internal/model"
	tc "github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	pg, err := tcpostgres.RunContainer(ctx,
		tcpostgres.WithDatabase("refuel_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tc.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		log.Fatalf("failed to start postgres: %v", err)
	}

	host, err := pg.Host(ctx)
	if err != nil {
		log.Fatalf("failed to get host: %v", err)
	}
	port, err := pg.MappedPort(ctx, "5432")
	if err != nil {
		log.Fatalf("failed to get port: %v", err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=test password=test dbname=refuel_test sslmode=disable", host, port.Port())

	testDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	if err := testDB.AutoMigrate(&model.User{}, &model.Activity{}, &model.DailyNutrition{}, &model.Recipe{}); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	code := m.Run()

	if err := pg.Terminate(ctx); err != nil {
		log.Fatalf("failed to terminate container: %v", err)
	}

	os.Exit(code)
}

func createUser(t *testing.T, db *gorm.DB) *model.User {
	t.Helper()
	user := &model.User{
		Email:    fmt.Sprintf("test-%d@test.com", time.Now().UnixNano()),
		Password: "hashed",
		Name:     "Test User",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return user
}
