package main

import (
	"fmt"
	"log"
	"os"

	"GameServer/internal/domain/entity"
	"GameServer/internal/infrastructure/config"
	"GameServer/internal/infrastructure/database"
	"GameServer/internal/infrastructure/repository"
)

func main() {
	// Set test environment variables
	os.Setenv("DB_HOST", "101.201.51.135")
	os.Setenv("DB_PORT", "3306")
	os.Setenv("DB_NAME", "gameserver")
	os.Setenv("DB_USER", "root")
	os.Setenv("DB_PASSWORD", "")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	dbConnection, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	// Create equipment repository
	equipRepo := repository.NewMySQLEquipmentRepository(dbConnection.GetDB())

	// Test auto-increment equipment ID
	fmt.Println("Testing Auto-Increment Equipment ID")
	fmt.Println("===================================")

	// Create test equipment
	testEquipment := &entity.Equipment{
		Quality:       1,
		Damage:        100,
		Crit:          10,
		CritDamage:    150,
		DamageSpeed:   120,
		BloodSuck:     5,
		HP:            200,
		MoveSpeed:     100,
		SuitID:        1,
		SuitName:      "Test Suit",
		EquipTypeID:   1,
		EquipTypeName: "Weapon",
		UserID:        1,
		Defense:       50,
		GoodFortune:   10,
		Type:          1,
	}

	// Test creating equipment
	fmt.Println("Creating test equipment...")
	err = equipRepo.Create(testEquipment)
	if err != nil {
		log.Printf("Error creating equipment: %v", err)
	} else {
		fmt.Printf("Equipment created successfully with auto-generated ID: %d\n", testEquipment.EquipID)
	}

	// Test retrieving equipment
	fmt.Println("Retrieving created equipment...")
	retrieved, err := equipRepo.GetByEquipID(testEquipment.EquipID)
	if err != nil {
		log.Printf("Error retrieving equipment: %v", err)
	} else if retrieved != nil {
		fmt.Printf("Retrieved equipment: ID=%d, SuitName=%s, EquipTypeName=%s\n",
			retrieved.EquipID, retrieved.SuitName, retrieved.EquipTypeName)
	}

	fmt.Println("Auto-Increment Equipment ID Test Completed!")
}
