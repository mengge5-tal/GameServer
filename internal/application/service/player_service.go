package service

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/entity"
	"GameServer/internal/domain/repository"
	"GameServer/internal/infrastructure/cache"
	"fmt"
)

// PlayerService handles player-related business logic
type PlayerService struct {
	playerRepo     repository.PlayerRepository
	equipmentRepo  repository.EquipmentRepository
	sourceStoneRepo repository.SourceStoneRepository
	cacheService   cache.CacheService
}

// NewPlayerService creates a new player service
func NewPlayerService(
	playerRepo repository.PlayerRepository,
	equipmentRepo repository.EquipmentRepository,
	sourceStoneRepo repository.SourceStoneRepository,
	cacheService cache.CacheService,
) *PlayerService {
	return &PlayerService{
		playerRepo:      playerRepo,
		equipmentRepo:   equipmentRepo,
		sourceStoneRepo: sourceStoneRepo,
		cacheService:    cacheService,
	}
}

// GetPlayerInfo retrieves player information
func (s *PlayerService) GetPlayerInfo(userID int) (*dto.PlayerInfoResponse, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("player_info:%d", userID)
	if cachedPlayer, err := s.cacheService.GetPlayerInfo(cacheKey); err == nil && cachedPlayer != nil {
		return &dto.PlayerInfoResponse{
			UserID:      cachedPlayer.UserID,
			Level:       cachedPlayer.Level,
			Experience:  cachedPlayer.Experience,
			GameLevel:   cachedPlayer.GameLevel,
			BloodEnergy: cachedPlayer.BloodEnergy,
		}, nil
	}

	// Get from database
	playerInfo, err := s.playerRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}
	if playerInfo == nil {
		return nil, entity.NewDomainError("player info not found")
	}

	// Cache the result
	s.cacheService.SetPlayerInfo(cacheKey, playerInfo)

	return &dto.PlayerInfoResponse{
		UserID:      playerInfo.UserID,
		Level:       playerInfo.Level,
		Experience:  playerInfo.Experience,
		GameLevel:   playerInfo.GameLevel,
		BloodEnergy: playerInfo.BloodEnergy,
	}, nil
}

// UpdatePlayer updates player information
func (s *PlayerService) UpdatePlayer(req *dto.UpdatePlayerRequest) error {
	// Get current player info
	playerInfo, err := s.playerRepo.GetByUserID(req.UserID)
	if err != nil {
		return err
	}
	if playerInfo == nil {
		return entity.NewDomainError("player info not found")
	}

	// Update fields if provided
	if req.Level != nil {
		playerInfo.Level = *req.Level
	}
	if req.Experience != nil {
		playerInfo.Experience = *req.Experience
	}
	if req.GameLevel != nil {
		playerInfo.GameLevel = *req.GameLevel
	}
	if req.BloodEnergy != nil {
		playerInfo.BloodEnergy = *req.BloodEnergy
	}

	// Update in database
	if err := s.playerRepo.Update(playerInfo); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("player_info:%d", req.UserID)
	s.cacheService.Delete(cacheKey)

	return nil
}

// GetUserEquipment retrieves all equipment for a user
func (s *PlayerService) GetUserEquipment(userID int) ([]*dto.EquipmentResponse, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("equipment:%d", userID)
	if cachedEquipment, err := s.cacheService.GetEquipment(cacheKey); err == nil && cachedEquipment != nil {
		return s.convertEquipmentToDTO(cachedEquipment), nil
	}

	// Get from database
	equipment, err := s.equipmentRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	s.cacheService.SetEquipment(cacheKey, equipment)

	return s.convertEquipmentToDTO(equipment), nil
}

// SaveEquipment saves or updates equipment
func (s *PlayerService) SaveEquipment(req *dto.SaveEquipmentRequest) error {
	// Convert DTO to entity
	equipment := &entity.Equipment{
		EquipID:     req.EquipID,
		Quality:     req.Quality,
		Damage:      req.Damage,
		Crit:        req.Crit,
		CritDamage:  req.CritDamage,
		DamageSpeed: req.DamageSpeed,
		BloodSuck:   req.BloodSuck,
		HP:          req.HP,
		MoveSpeed:   req.MoveSpeed,
		EquipName:   req.EquipName,
		UserID:      req.UserID,
		Defense:     req.Defense,
		GoodFortune: req.GoodFortune,
		Type:        req.Type,
	}

	// Check if equipment exists
	existing, err := s.equipmentRepo.GetByEquipID(req.EquipID)
	if err != nil {
		return err
	}

	if existing != nil {
		// Update existing equipment
		if err := s.equipmentRepo.Update(equipment); err != nil {
			return err
		}
	} else {
		// Create new equipment
		if err := s.equipmentRepo.Create(equipment); err != nil {
			return err
		}
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("equipment:%d", req.UserID)
	s.cacheService.Delete(cacheKey)

	return nil
}

// DeleteEquipment deletes equipment
func (s *PlayerService) DeleteEquipment(equipID, userID int) error {
	// Delete from database
	if err := s.equipmentRepo.Delete(equipID); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("equipment:%d", userID)
	s.cacheService.Delete(cacheKey)

	return nil
}

// GetUserSourceStones retrieves all source stones for a user
func (s *PlayerService) GetUserSourceStones(userID int) ([]*dto.SourceStoneResponse, error) {
	sourceStones, err := s.sourceStoneRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	var response []*dto.SourceStoneResponse
	for _, stone := range sourceStones {
		response = append(response, &dto.SourceStoneResponse{
			EquipID:    stone.EquipID,
			SourceType: stone.SourceType,
			Count:      stone.Count,
			Quality:    stone.Quality,
			UserID:     stone.UserID,
		})
	}

	return response, nil
}

// convertEquipmentToDTO converts equipment entities to DTOs
func (s *PlayerService) convertEquipmentToDTO(equipment []*entity.Equipment) []*dto.EquipmentResponse {
	var response []*dto.EquipmentResponse
	for _, equip := range equipment {
		response = append(response, &dto.EquipmentResponse{
			EquipID:     equip.EquipID,
			Quality:     equip.Quality,
			Damage:      equip.Damage,
			Crit:        equip.Crit,
			CritDamage:  equip.CritDamage,
			DamageSpeed: equip.DamageSpeed,
			BloodSuck:   equip.BloodSuck,
			HP:          equip.HP,
			MoveSpeed:   equip.MoveSpeed,
			EquipName:   equip.EquipName,
			UserID:      equip.UserID,
			Defense:     equip.Defense,
			GoodFortune: equip.GoodFortune,
			Type:        equip.Type,
		})
	}
	return response
}