package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/taraslis453/territory-service-bot/config"
	"github.com/taraslis453/territory-service-bot/internal/entity"
	"github.com/taraslis453/territory-service-bot/pkg/database"
	"github.com/taraslis453/territory-service-bot/pkg/logging"
)

func main() {
	logger := logging.NewZap("seed")
	logger.Info("Starting database seeding...")

	cfg := config.Get()

	sql, err := database.NewPostgreSQL(&database.PostgreSQLConfig{
		User:     cfg.PostgreSQL.User,
		Password: cfg.PostgreSQL.Password,
		Host:     cfg.PostgreSQL.Host,
		Database: cfg.PostgreSQL.Database,
	})
	if err != nil {
		logger.Fatal("failed to init postgresql", "err", err)
	}

	// Enable UUID extension
	err = sql.DB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error
	if err != nil {
		logger.Fatal("failed to create extension uuid-ossp", "err", err)
	}

	// Run migrations
	err = sql.DB.AutoMigrate(
		&entity.User{},
		&entity.Congregation{},
		&entity.CongregationTerritory{},
		&entity.CongregationTerritoryNote{},
		&entity.CongregationTerritoryGroup{},
		&entity.RequestActionState{},
	)
	if err != nil {
		logger.Fatal("automigration failed", "err", err)
	}

	logger.Info("Migrations completed successfully")

	// Clear existing data (optional - comment out if you want to keep existing data)
	logger.Info("Clearing existing data...")
	sql.DB.Exec("DELETE FROM congregation_territory_notes")
	sql.DB.Exec("DELETE FROM congregation_territories")
	sql.DB.Exec("DELETE FROM congregation_territory_groups")
	sql.DB.Exec("DELETE FROM users")
	sql.DB.Exec("DELETE FROM congregations")
	sql.DB.Exec("DELETE FROM request_action_states")

	// Seed Congregations
	logger.Info("Seeding congregations...")
	congregations := []entity.Congregation{
		{
			ID:   uuid.New().String(),
			Name: "Lorem Central Congregation",
		},
		{
			ID:   uuid.New().String(),
			Name: "Ipsum North Congregation",
		},
		{
			ID:   uuid.New().String(),
			Name: "Dolor South Congregation",
		},
	}

	for i := range congregations {
		if err := sql.DB.Create(&congregations[i]).Error; err != nil {
			logger.Error("failed to create congregation", "name", congregations[i].Name, "err", err)
		} else {
			logger.Info("Created congregation", "name", congregations[i].Name)
		}
	}

	// Seed Territory Groups
	logger.Info("Seeding territory groups...")
	territoryGroups := []entity.CongregationTerritoryGroup{
		{
			ID:             uuid.New().String(),
			CongregationID: congregations[0].ID,
			Title:          "Lorem District",
		},
		{
			ID:             uuid.New().String(),
			CongregationID: congregations[0].ID,
			Title:          "Ipsum Heights",
		},
		{
			ID:             uuid.New().String(),
			CongregationID: congregations[0].ID,
			Title:          "Dolor Hills",
		},
		{
			ID:             uuid.New().String(),
			CongregationID: congregations[1].ID,
			Title:          "Amet Center",
		},
		{
			ID:             uuid.New().String(),
			CongregationID: congregations[1].ID,
			Title:          "Sit Quarter",
		},
	}

	for i := range territoryGroups {
		if err := sql.DB.Create(&territoryGroups[i]).Error; err != nil {
			logger.Error("failed to create territory group", "title", territoryGroups[i].Title, "err", err)
		} else {
			logger.Info("Created territory group", "title", territoryGroups[i].Title)
		}
	}

	// Seed Users
	logger.Info("Seeding users...")
	users := []entity.User{
		{
			ID:              uuid.New().String(),
			CongregationID:  congregations[0].ID,
			MessengerUserID: "123456789",
			MessengerChatID: "123456789",
			FullName:        "John Doe",
			Role:            entity.UserRoleAdmin,
			Stage:           entity.UserStageSelectActionFromMenu,
		},
		{
			ID:              uuid.New().String(),
			CongregationID:  congregations[0].ID,
			MessengerUserID: "987654321",
			MessengerChatID: "987654321",
			FullName:        "Jane Smith",
			Role:            entity.UserRolePublisher,
			Stage:           entity.UserStageSelectActionFromMenu,
		},
		{
			ID:              uuid.New().String(),
			CongregationID:  congregations[1].ID,
			MessengerUserID: "555666777",
			MessengerChatID: "555666777",
			FullName:        "Bob Johnson",
			Role:            entity.UserRoleAdmin,
			Stage:           entity.UserStageSelectActionFromMenu,
		},
		{
			ID:                 uuid.New().String(),
			JoinCongregationID: congregations[0].ID,
			MessengerUserID:    "111222333",
			MessengerChatID:    "111222333",
			FullName:           "Alice Williams",
			Role:               entity.UserRolePublisher,
			Stage:              entity.UserPublisherStageWaitingForAdminApproval,
		},
	}

	for i := range users {
		if err := sql.DB.Create(&users[i]).Error; err != nil {
			logger.Error("failed to create user", "name", users[i].FullName, "err", err)
		} else {
			logger.Info("Created user", "name", users[i].FullName, "role", users[i].Role)
		}
	}

	// Seed Territories
	logger.Info("Seeding territories...")

	territories := []entity.CongregationTerritory{
		{
			ID:             uuid.New().String(),
			CongregationID: congregations[0].ID,
			Title:          "Terr-1 Lorem Street",
			GroupID:        territoryGroups[0].ID,
			FileID:         "file_001",
			FileType:       entity.CongregationTerritoryFileTypePhoto,
			LastTakenAt:    time.Now().AddDate(0, -2, 0),
		},
		{
			ID:             uuid.New().String(),
			CongregationID: congregations[0].ID,
			Title:          "Terr-2 Ipsum Square",
			GroupID:        territoryGroups[0].ID,
			FileID:         "file_002",
			FileType:       entity.CongregationTerritoryFileTypeDocument,
			LastTakenAt:    time.Now().AddDate(0, -1, 0),
		},
		{
			ID:             uuid.New().String(),
			CongregationID: congregations[0].ID,
			Title:          "Terr-3 Dolor Avenue",
			GroupID:        territoryGroups[1].ID,
			FileID:         "file_003",
			FileType:       entity.CongregationTerritoryFileTypePhoto,
			InUseByUserID:  &users[1].ID,
			LastTakenAt:    time.Now(),
		},
		{
			ID:             uuid.New().String(),
			CongregationID: congregations[0].ID,
			Title:          "Terr-4 Amet Boulevard",
			GroupID:        territoryGroups[2].ID,
			FileID:         "file_004",
			FileType:       entity.CongregationTerritoryFileTypePhoto,
			LastTakenAt:    time.Now().AddDate(0, -3, 0),
		},
		{
			ID:             uuid.New().String(),
			CongregationID: congregations[1].ID,
			Title:          "Terr-1 Sit Center",
			GroupID:        territoryGroups[3].ID,
			FileID:         "file_005",
			FileType:       entity.CongregationTerritoryFileTypeDocument,
			LastTakenAt:    time.Now().AddDate(0, 0, -15),
		},
		{
			ID:             uuid.New().String(),
			CongregationID: congregations[1].ID,
			Title:          "Terr-2 Consectetur Park",
			GroupID:        territoryGroups[4].ID,
			FileID:         "file_006",
			FileType:       entity.CongregationTerritoryFileTypePhoto,
			LastTakenAt:    time.Now().AddDate(0, -1, -10),
		},
	}

	for i := range territories {
		if err := sql.DB.Create(&territories[i]).Error; err != nil {
			logger.Error("failed to create territory", "title", territories[i].Title, "err", err)
		} else {
			status := "available"
			if territories[i].InUseByUserID != nil {
				status = "in use"
			}
			logger.Info("Created territory", "title", territories[i].Title, "status", status)
		}
	}

	// Seed Territory Notes
	logger.Info("Seeding territory notes...")
	territoryNotes := []entity.CongregationTerritoryNote{
		{
			ID:          uuid.New().String(),
			TerritoryID: territories[2].ID,
			UserID:      users[1].ID,
			Text:        "Lorem ipsum dolor sit amet, visited 3 buildings. Left literature.",
			CreatedAt:   time.Now().AddDate(0, 0, -1),
		},
		{
			ID:          uuid.New().String(),
			TerritoryID: territories[2].ID,
			UserID:      users[1].ID,
			Text:        "Consectetur adipiscing elit. Very well received by residents.",
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			TerritoryID: territories[0].ID,
			UserID:      users[0].ID,
			Text:        "Territory was fully covered last month. Sed do eiusmod tempor incididunt.",
			CreatedAt:   time.Now().AddDate(0, -1, 0),
		},
	}

	for i := range territoryNotes {
		if err := sql.DB.Create(&territoryNotes[i]).Error; err != nil {
			logger.Error("failed to create territory note", "err", err)
		} else {
			logger.Info("Created territory note")
		}
	}

	logger.Info("Database seeding completed successfully!")
	logger.Info("Summary:")
	logger.Info("  - Congregations: ", "count", len(congregations))
	logger.Info("  - Territory Groups: ", "count", len(territoryGroups))
	logger.Info("  - Users: ", "count", len(users))
	logger.Info("  - Territories: ", "count", len(territories))
	logger.Info("  - Territory Notes: ", "count", len(territoryNotes))
}
