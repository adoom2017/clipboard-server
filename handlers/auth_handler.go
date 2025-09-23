package handlers

import (
	"clipboard-server/auth"
	"clipboard-server/database"
	"clipboard-server/models"
	"clipboard-server/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthHandler for authentication related handlers
type AuthHandler struct{}

// NewAuthHandler creates auth handler instance
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// Register user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	// Validate username
	if err := utils.ValidateUsername(req.Username); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid username",
			Message: err.Error(),
		})
		return
	}

	// Validate password
	if err := utils.ValidatePassword(req.Password); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid password",
			Message: err.Error(),
		})
		return
	}

	// Validate email
	if !utils.ValidateEmail(req.Email) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid email",
			Message: "invalid email format",
		})
		return
	}

	db := database.GetDB()

	// Check if username exists
	var existingUser models.User
	if err := db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error:   "username already exists",
			Message: "the username is already taken",
		})
		return
	}

	// Check if email exists
	if err := db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error:   "email already exists",
			Message: "the email is already registered",
		})
		return
	}

	// Generate salt
	salt, err := utils.GenerateSalt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "salt generation failed",
			Message: "failed to generate password salt",
		})
		return
	}

	// Hash password with salt
	hashedPassword, err := utils.HashPasswordWithSalt(req.Password, salt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "password encryption failed",
			Message: "failed to encrypt password",
		})
		return
	}

	// Create user
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Salt:     salt,
		IsActive: true,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "user creation failed",
			Message: "failed to create user",
		})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "token generation failed",
			Message: "failed to generate authentication token",
		})
		return
	}

	// Save token to user record
	user.Token = token
	db.Save(&user)

	// Return login info (without password)
	user.Password = ""
	c.JSON(http.StatusCreated, models.LoginResponse{
		Token: token,
		User:  user,
	})
}

// Login user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	db := database.GetDB()
	var user models.User

	// Find user (support username or email login)
	query := db.Where("username = ? OR email = ?", req.Username, req.Username)
	if err := query.First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "invalid credentials",
				Message: "username or password is incorrect",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database error",
			Message: "failed to query user",
		})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusForbidden, models.ErrorResponse{
			Error:   "account disabled",
			Message: "your account has been disabled",
		})
		return
	}

	// Verify password with salt (支持向后兼容)
	var passwordValid bool
	if user.Salt != "" {
		// 新用户：使用盐值验证
		passwordValid = utils.CheckPasswordWithSalt(req.Password, user.Salt, user.Password)
	} else {
		// 旧用户：使用旧方法验证，然后升级他们的密码
		passwordValid = utils.CheckPassword(req.Password, user.Password)

		if passwordValid {
			// 密码验证通过，升级用户的密码存储方式
			salt, err := utils.GenerateSalt()
			if err == nil {
				hashedPassword, err := utils.HashPasswordWithSalt(req.Password, salt)
				if err == nil {
					user.Salt = salt
					user.Password = hashedPassword
					database.GetDB().Save(&user) // 保存升级后的密码
					fmt.Printf("用户 %s 的密码存储方式已升级\n", user.Username)
				}
			}
		}
	}

	if !passwordValid {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "invalid credentials",
			Message: "username or password is incorrect",
		})
		return
	}

	// Generate new JWT token
	token, err := auth.GenerateToken(user.ID, user.Username, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "token generation failed",
			Message: "failed to generate authentication token",
		})
		return
	}

	// Update user token
	user.Token = token
	db.Save(&user)

	// Return login info (without password)
	user.Password = ""
	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		User:  user,
	})
}

// RefreshToken refresh token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "missing token",
			Message: "authorization token is required",
		})
		return
	}

	// Remove "Bearer " prefix
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Refresh token
	newToken, err := auth.RefreshToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "token refresh failed",
			Message: err.Error(),
		})
		return
	}

	// Get user info
	claims, err := auth.ParseToken(newToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "token parsing failed",
			Message: "failed to parse new token",
		})
		return
	}

	// Update token in database
	db := database.GetDB()
	db.Model(&models.User{}).Where("id = ?", claims.UserID).Update("token", newToken)

	c.JSON(http.StatusOK, gin.H{
		"token":      newToken,
		"expires_at": time.Unix(claims.ExpiresAt, 0).Format(time.RFC3339),
	})
}

// Logout user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "user not authenticated",
		})
		return
	}

	// Clear token in database
	db := database.GetDB()
	db.Model(&models.User{}).Where("id = ?", userID).Update("token", "")

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "logout successful",
	})
}

// GetProfile get user profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "user not authenticated",
		})
		return
	}

	db := database.GetDB()
	var user models.User

	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "user not found",
				Message: "user profile not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database error",
			Message: "failed to get user profile",
		})
		return
	}

	// Don't return password and token
	user.Password = ""
	user.Token = ""

	c.JSON(http.StatusOK, user)
}

// ChangePassword change user password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "user not authenticated",
		})
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	// Validate new password
	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid new password",
			Message: err.Error(),
		})
		return
	}

	db := database.GetDB()
	var user models.User

	// Get current user
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "user not found",
				Message: "user profile not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database error",
			Message: "failed to get user profile",
		})
		return
	}

	// Verify current password
	if !utils.CheckPasswordWithSalt(req.CurrentPassword, user.Salt, user.Password) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "invalid current password",
			Message: "current password is incorrect",
		})
		return
	}

	// Generate new salt
	newSalt, err := utils.GenerateSalt()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "salt generation failed",
			Message: "failed to generate new salt",
		})
		return
	}

	// Hash new password with new salt
	hashedNewPassword, err := utils.HashPasswordWithSalt(req.NewPassword, newSalt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "password encryption failed",
			Message: "failed to encrypt new password",
		})
		return
	}

	// Update user password and salt
	if err := db.Model(&user).Updates(models.User{
		Password: hashedNewPassword,
		Salt:     newSalt,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update failed",
			Message: "failed to update password",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "password changed successfully",
	})
}
