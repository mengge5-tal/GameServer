package server

import (
	"GameServer/pkg/logger"
	"time"
)

// AuthMiddleware 认证中间件
func AuthMiddleware() Middleware {
	return MiddlewareFunc(func(c *Client, message *Message, next Handler) *Response {
		// 对于认证和心跳消息，不需要验证
		if message.Type == MessageTypeAuth || message.Type == MessageTypeHeartbeat {
			return next.Handle(c, message)
		}

		// 检查客户端是否已认证
		if !c.IsAuth {
			logger.Warn("Unauthorized access attempt", map[string]interface{}{
				"client_id":    c.ID,
				"message_type": message.Type,
				"action":       message.Action,
			})
			return NewErrorResponse(message.RequestID, CodeUnauthorized, "Authentication required")
		}

		return next.Handle(c, message)
	})
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware() Middleware {
	return MiddlewareFunc(func(c *Client, message *Message, next Handler) *Response {
		start := time.Now()
		
		logger.Info("Processing message", map[string]interface{}{
			"client_id":    c.ID,
			"user_id":      c.UserID,
			"message_type": message.Type,
			"action":       message.Action,
			"request_id":   message.RequestID,
		})

		response := next.Handle(c, message)
		
		duration := time.Since(start)
		logLevel := "info"
		if !response.Success {
			logLevel = "warn"
		}

		logData := map[string]interface{}{
			"client_id":    c.ID,
			"user_id":      c.UserID,
			"message_type": message.Type,
			"action":       message.Action,
			"request_id":   message.RequestID,
			"success":      response.Success,
			"code":         response.Code,
			"duration_ms":  duration.Milliseconds(),
		}

		if logLevel == "info" {
			logger.Info("Message processed successfully", logData)
		} else {
			logger.Warn("Message processing failed", logData)
		}

		return response
	})
}

// RateLimitMiddleware 限流中间件（简单实现）
func RateLimitMiddleware() Middleware {
	clientRequestCounts := make(map[string][]time.Time)
	maxRequestsPerMinute := 60

	return MiddlewareFunc(func(c *Client, message *Message, next Handler) *Response {
		now := time.Now()
		clientID := c.ID

		// 清理过期的请求记录
		if requests, exists := clientRequestCounts[clientID]; exists {
			validRequests := make([]time.Time, 0)
			for _, requestTime := range requests {
				if now.Sub(requestTime) < time.Minute {
					validRequests = append(validRequests, requestTime)
				}
			}
			clientRequestCounts[clientID] = validRequests
		}

		// 检查请求频率
		if len(clientRequestCounts[clientID]) >= maxRequestsPerMinute {
			logger.Warn("Rate limit exceeded", map[string]interface{}{
				"client_id":      c.ID,
				"user_id":        c.UserID,
				"request_count":  len(clientRequestCounts[clientID]),
				"limit":          maxRequestsPerMinute,
			})
			return NewErrorResponse(message.RequestID, CodeServerError, "Rate limit exceeded")
		}

		// 记录当前请求
		clientRequestCounts[clientID] = append(clientRequestCounts[clientID], now)

		return next.Handle(c, message)
	})
}

// ValidationMiddleware 消息验证中间件
func ValidationMiddleware() Middleware {
	return MiddlewareFunc(func(c *Client, message *Message, next Handler) *Response {
		// 验证必要字段
		if message.Type == "" {
			return NewErrorResponse(message.RequestID, CodeInvalidRequest, "Message type is required")
		}

		if message.Action == "" {
			return NewErrorResponse(message.RequestID, CodeInvalidRequest, "Action is required")
		}

		if message.RequestID == "" {
			return NewErrorResponse(message.RequestID, CodeInvalidRequest, "Request ID is required")
		}

		return next.Handle(c, message)
	})
}