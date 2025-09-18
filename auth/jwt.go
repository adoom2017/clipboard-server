package auth

import (
	"clipboard-server/config"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.StandardClaims
}

// GenerateToken 生成JWT令牌
func GenerateToken(userID, username, email string) (string, error) {
	cfg := config.GetConfig()

	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(cfg.JWTExpireHour)).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "clipboard-sync-server",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// ParseToken 解析JWT令牌
func ParseToken(tokenString string) (*JWTClaims, error) {
	cfg := config.GetConfig()

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// JWTAuthMiddleware JWT认证中间�?
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{
				"error":   "unauthorized",
				"message": "missing authorization header",
			})
			c.Abort()
			return
		}

		// 移除 "Bearer " 前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		claims, err := ParseToken(token)
		if err != nil {
			c.JSON(401, gin.H{
				"error":   "unauthorized",
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// 将用户信息设置到上下文中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// GetCurrentUser 从上下文中获取当前用户信�?
func GetCurrentUser(c *gin.Context) (userID, username, email string, exists bool) {
	userIDInterface, exists1 := c.Get("user_id")
	usernameInterface, exists2 := c.Get("username")
	emailInterface, exists3 := c.Get("email")

	if !exists1 || !exists2 || !exists3 {
		return "", "", "", false
	}

	userID, ok1 := userIDInterface.(string)
	username, ok2 := usernameInterface.(string)
	email, ok3 := emailInterface.(string)

	if !ok1 || !ok2 || !ok3 {
		return "", "", "", false
	}

	return userID, username, email, true
}

// GetCurrentUserID 从上下文中获取当前用户ID
func GetCurrentUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}

	userIDStr, ok := userID.(string)
	return userIDStr, ok
}

// RefreshToken 刷新JWT令牌
func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查令牌是否即将过期（�?小时内过期才允许刷新�?
	if time.Until(time.Unix(claims.ExpiresAt, 0)) > time.Hour {
		return "", errors.New("token is not eligible for refresh yet")
	}

	// 生成新令�?
	return GenerateToken(claims.UserID, claims.Username, claims.Email)
}
