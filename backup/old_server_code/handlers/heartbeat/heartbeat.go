package heartbeat

// HeartbeatHandler 心跳处理器
type HeartbeatHandler struct{}

// NewHeartbeatHandler 创建心跳处理器
func NewHeartbeatHandler() *HeartbeatHandler {
	return &HeartbeatHandler{}
}

// Response 响应结构
type Response struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"requestId"`
	Timestamp int64       `json:"timestamp"`
}

// Client 客户端接口
type Client interface {
	GetID() string
}

const (
	CodeSuccess        = 0
	CodeInvalidRequest = 1001
)

// HandlePing 处理ping消息
func (h *HeartbeatHandler) HandlePing(c Client, data interface{}, requestID string) *Response {
	return NewSuccessResponse(requestID, map[string]string{
		"action": "pong",
		"time":   "ok",
	})
}

func NewSuccessResponse(requestID string, data interface{}) *Response {
	return &Response{
		Success:   true,
		Code:      CodeSuccess,
		Message:   "Success",
		Data:      data,
		RequestID: requestID,
		Timestamp: 0,
	}
}

func NewErrorResponse(requestID string, code int, message string) *Response {
	return &Response{
		Success:   false,
		Code:      code,
		Message:   message,
		Data:      nil,
		RequestID: requestID,
		Timestamp: 0,
	}
}