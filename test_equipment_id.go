package main

import (
	"fmt"
	"log"
	"os"

	"GameServer/internal/infrastructure/config"
	"GameServer/internal/infrastructure/database"
	"GameServer/internal/infrastructure/repository"
)

func main() {
	// Set test environment variables
	os.Setenv("DB_HOST", "localhost")
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

	// Test equipment ID generation logic
	fmt.Println("Testing Equipment ID Generation Logic")
	fmt.Println("====================================")

	testCases := []struct {
		equipType int
		quality   int
		expected  string
	}{
		{4, 1, "41000001"}, // First type=4, quality=1 equipment
		{4, 1, "41000002"}, // Second type=4, quality=1 equipment  
		{2, 3, "23000001"}, // First type=2, quality=3 equipment
		{4, 2, "42000001"}, // First type=4, quality=2 equipment
	}

	for i, tc := range testCases {
		fmt.Printf("Test case %d: type=%d, quality=%d\n", i+1, tc.equipType, tc.quality)
		
		// Get current max sequence
		maxSeq, err := equipRepo.GetMaxSequenceByTypeAndQuality(tc.equipType, tc.quality)
		if err != nil {
			log.Printf("Error getting max sequence: %v", err)
			continue
		}
		
		// Calculate expected equipment ID
		newSequence := maxSeq + 1
		expectedID := tc.equipType*10000000 + tc.quality*1000000 + newSequence
		
		fmt.Printf("  Current max sequence: %d\n", maxSeq)
		fmt.Printf("  New sequence: %d\n", newSequence)
		fmt.Printf("  Generated Equipment ID: %d\n", expectedID)
		fmt.Printf("  Expected format: %s\n", tc.expected)
		fmt.Println()
	}

	fmt.Println("Equipment ID Generation Test Completed!")
}