package utils

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

var requestCounter int64

// RequestIDGenerator generates unique request IDs for different message types and actions
type RequestIDGenerator struct{}

// NewRequestIDGenerator creates a new request ID generator
func NewRequestIDGenerator() *RequestIDGenerator {
	return &RequestIDGenerator{}
}

// Generate creates a unique request ID with format: {type}_{action}_{uuid}_{counter}
func (g *RequestIDGenerator) Generate(msgType, action string) string {
	counter := atomic.AddInt64(&requestCounter, 1)
	shortUUID := uuid.New().String()[:8] // Use first 8 characters of UUID
	
	return fmt.Sprintf("%s_%s_%s_%d", msgType, action, shortUUID, counter)
}

// GenerateWithTimestamp creates a unique request ID with timestamp: {type}_{action}_{timestamp}_{counter}
func (g *RequestIDGenerator) GenerateWithTimestamp(msgType, action string) string {
	counter := atomic.AddInt64(&requestCounter, 1)
	timestamp := time.Now().Unix()
	
	return fmt.Sprintf("%s_%s_%d_%d", msgType, action, timestamp, counter)
}

// GenerateSimple creates a simple unique request ID: {type}_{action}_{counter}
func (g *RequestIDGenerator) GenerateSimple(msgType, action string) string {
	counter := atomic.AddInt64(&requestCounter, 1)
	
	return fmt.Sprintf("%s_%s_%d", msgType, action, counter)
}