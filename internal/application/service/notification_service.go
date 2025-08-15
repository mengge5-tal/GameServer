package service

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/valueobject"
)

// NotificationService defines the interface for sending real-time notifications
type NotificationService interface {
	SendFriendRequestNotification(toUserID int, notification *dto.FriendRequestNotification) error
}

// notificationService implements NotificationService
type notificationService struct {
	userNotifier UserNotifier
}

// UserNotifier defines the interface for sending messages to specific users
type UserNotifier interface {
	SendToUser(userID int, message []byte) bool
}

// NewNotificationService creates a new notification service
func NewNotificationService(userNotifier UserNotifier) NotificationService {
	return &notificationService{
		userNotifier: userNotifier,
	}
}

// SendFriendRequestNotification sends a friend request notification to a specific user
func (s *notificationService) SendFriendRequestNotification(toUserID int, notification *dto.FriendRequestNotification) error {
	println("Attempting to send notification to user:", toUserID)
	
	// Create response message
	response := valueobject.NewSuccessResponseWithUniqueID(
		valueobject.MessageTypeFriend,
		valueobject.ActionFriendRequestNotification,
		notification,
	)
	
	// Convert to JSON
	messageBytes, err := response.ToJSON()
	if err != nil {
		println("Failed to convert notification to JSON:", err.Error())
		return err
	}
	
	println("Notification JSON:", string(messageBytes))
	
	// Send to user (returns false if user is offline, but we don't treat that as error)
	success := s.userNotifier.SendToUser(toUserID, messageBytes)
	if !success {
		println("Failed to send notification - user might be offline or not found")
		return nil // Don't treat offline user as error
	}
	
	println("Notification sent successfully to user:", toUserID)
	return nil
}