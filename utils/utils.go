package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

// GenerateSalt 生成随机盐值
func GenerateSalt() (string, error) {
	salt := make([]byte, 32) // 32字节盐值
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(salt), nil
}

// HashPasswordWithSalt 使用盐值进行密码哈希
func HashPasswordWithSalt(password, salt string) (string, error) {
	// 将密码和盐值结合
	saltedPassword := password + salt

	// 使用SHA256预哈希来解决BCrypt的72字节长度限制
	preHash := sha256.Sum256([]byte(saltedPassword))
	preHashString := hex.EncodeToString(preHash[:])

	// 使用bcrypt进行最终哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(preHashString), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPasswordWithSalt 验证带盐的密码
func CheckPasswordWithSalt(password, salt, hash string) bool {
	// 将密码和盐值结合
	saltedPassword := password + salt

	// 使用相同的SHA256预哈希
	preHash := sha256.Sum256([]byte(saltedPassword))
	preHashString := hex.EncodeToString(preHash[:])

	// 使用bcrypt验证
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(preHashString))
	return err == nil
}

// HashPassword 为了向后兼容保留的简单哈希函数（已废弃，建议使用HashPasswordWithSalt）
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword 为了向后兼容保留的简单验证函数（已废弃，建议使用CheckPasswordWithSalt）
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	rand.Read(bytes)

	for i := 0; i < length; i++ {
		bytes[i] = charset[int(bytes[i])%len(charset)]
	}
	return string(bytes)
}

func GenerateSecureHash(data string) string {
	hash := sha256.Sum256([]byte(data + time.Now().String()))
	return hex.EncodeToString(hash[:])
}

func ValidateEmail(email string) bool {
	if email == "" {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	local, domain := parts[0], parts[1]
	if local == "" || domain == "" {
		return false
	}

	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}

func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters")
	}

	if len(username) > 50 {
		return fmt.Errorf("username cannot be longer than 50 characters")
	}

	for _, char := range username {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '_' || char == '-') {
			return fmt.Errorf("username can only contain letters, numbers, underscores and hyphens")
		}
	}

	return nil
}

func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}

	if len(password) > 100 {
		return fmt.Errorf("password cannot be longer than 100 characters")
	}

	return nil
}

func TruncateString(s string, maxLength int) string {
	if utf8.RuneCountInString(s) <= maxLength {
		return s
	}

	runes := []rune(s)
	if len(runes) > maxLength {
		return string(runes[:maxLength-3]) + "..."
	}
	return s
}

func SanitizeContent(content string) string {
	patterns := []string{
		//	"password",
		//	"passwd",
		//	"pwd",
		//	"secret",s
		//	"token",S
		//	"key",S
		//	"auth",
	}

	lowerContent := strings.ToLower(content)
	for _, pattern := range patterns {
		if strings.Contains(lowerContent, pattern) {
			if len(content) < 100 && (strings.Contains(lowerContent, "=") || strings.Contains(lowerContent, ":")) {
				return "[SENSITIVE_CONTENT_HIDDEN]"
			}
		}
	}

	return content
}

func GetContentSize(content string) int64 {
	return int64(len([]byte(content)))
}

func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(size)/float64(div), units[exp])
}

func IsValidContentType(contentType string) bool {
	validTypes := []string{"text", "image", "file"}
	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

func GenerateContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func StringPtr(s string) *string {
	return &s
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

func SafeTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%.0fh", d.Hours())
	}
	return fmt.Sprintf("%.0fd", d.Hours()/24)
}

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func RemoveEmpty(slice []string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if strings.TrimSpace(s) != "" {
			result = append(result, s)
		}
	}
	return result
}
