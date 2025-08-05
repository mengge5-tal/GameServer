package auth

import (
	"database/sql"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"unicode"
)

// LoginRequest 登录请求数据
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest 注册请求数据
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// User 用户信息
type User struct {
	UserID   int    `json:"userid"`
	Username string `json:"username"`
}

// AuthService 认证服务
type AuthService struct {
	db *sql.DB
}

// NewAuthService 创建认证服务
func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{db: db}
}

// HashPassword 密码哈希函数 - 使用bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// HashPasswordWithCost 使用指定成本的密码哈希函数
func HashPasswordWithCost(password string, cost int) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

// CheckPasswordHash 验证密码
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePassword 密码强度验证
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return NewCustomError("Password must be at least 8 characters long")
	}
	if len(password) > 128 {
		return NewCustomError("Password must be less than 128 characters")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return NewCustomError("Password must contain uppercase, lowercase, number and special character")
	}

	return nil
}

// ValidateUsername 用户名验证
func ValidateUsername(username string) error {
	if len(username) < 3 || len(username) > 20 {
		return NewCustomError("Username must be 3-20 characters")
	}

	// 只允许字母、数字和下划线
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]+$", username)
	if !matched {
		return NewCustomError("Username can only contain letters, numbers and underscores")
	}

	return nil
}

// CustomError 自定义错误类型
type CustomError struct {
	Message string
}

func (e *CustomError) Error() string {
	return e.Message
}

func NewCustomError(message string) *CustomError {
	return &CustomError{Message: message}
}
