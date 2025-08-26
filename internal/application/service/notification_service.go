package service

import (
	"GameServer/internal/application/dto"
	"GameServer/internal/domain/valueobject"
)

// NotificationService defines the interface for sending real-time notifications
type NotificationService interface {
	SendFriendRequestNotification(toUserID int, notification *dto.FriendRequestNotification) error
	SendUnionJoinRequestNotification(chairpersonID int, notification *dto.UnionJoinRequestNotification) error
	SendUnionInviteNotification(toUserID int, notification *dto.UnionInviteNotification) error
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

// SendUnionJoinRequestNotification sends a union join request notification to chairperson
func (s *notificationService) SendUnionJoinRequestNotification(chairpersonID int, notification *dto.UnionJoinRequestNotification) error {
	println("Attempting to send union join request notification to chairperson:", chairpersonID)
	
	// Create response message
	response := valueobject.NewSuccessResponseWithUniqueID(
		valueobject.MessageTypeUnion,
		valueobject.ActionJoinUnion, // Using same action but this will be a notification push
		notification,
	)
	
	// Convert to JSON
	messageBytes, err := response.ToJSON()
	if err != nil {
		println("Failed to convert union notification to JSON:", err.Error())
		return err
	}
	
	println("Union notification JSON:", string(messageBytes))
	
	// Send to chairperson (returns false if user is offline, but we don't treat that as error)
	success := s.userNotifier.SendToUser(chairpersonID, messageBytes)
	if !success {
		println("Failed to send union notification - chairperson might be offline or not found")
		return nil // Don't treat offline user as error
	}
	
	println("Union join request notification sent successfully to chairperson:", chairpersonID)
	return nil
}

// SendUnionInviteNotification sends a union invite notification to a specific user
func (s *notificationService) SendUnionInviteNotification(toUserID int, notification *dto.UnionInviteNotification) error {
	println("Attempting to send union invite notification to user:", toUserID)
	
	// Create response message
	response := valueobject.NewSuccessResponseWithUniqueID(
		valueobject.MessageTypeUnion,
		valueobject.ActionInviteToUnion, // Using invite action for notification push
		notification,
	)
	
	// Convert to JSON
	messageBytes, err := response.ToJSON()
	if err != nil {
		println("Failed to convert union invite notification to JSON:", err.Error())
		return err
	}
	
	println("Union invite notification JSON:", string(messageBytes))
	
	// Send to user (returns false if user is offline, but we don't treat that as error)
	success := s.userNotifier.SendToUser(toUserID, messageBytes)
	if !success {
		println("Failed to send union invite notification - user might be offline or not found")
		return nil // Don't treat offline user as error
	}
	
	println("Union invite notification sent successfully to user:", toUserID)
	return nil
}