package handlers

import (
	"bytes"
	"clipboard-server/auth"
	"clipboard-server/database"
	"clipboard-server/models"
	"clipboard-server/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}

	// 自动迁移
	db.AutoMigrate(&models.User{}, &models.ClipboardItem{})
	return db
}

func TestChangePassword(t *testing.T) {
	// 设置测试数据库
	database.DB = setupTestDB()
	defer func() {
		sqlDB, _ := database.DB.DB()
		sqlDB.Close()
	}()

	// 创建测试用户
	salt, _ := utils.GenerateSalt()
	hashedPassword, _ := utils.HashPasswordWithSalt("oldpassword123", salt)
	user := models.User{
		ID:       "test-user-id",
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		Salt:     salt,
		IsActive: true,
	}
	database.DB.Create(&user)

	// 生成JWT token
	token, _ := auth.GenerateToken(user.ID, user.Username, user.Email)

	// 设置Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authHandler := NewAuthHandler()

	// 设置路由
	authenticated := router.Group("/")
	authenticated.Use(auth.JWTAuthMiddleware())
	authenticated.PUT("/password", authHandler.ChangePassword)

	tests := []struct {
		name           string
		requestBody    models.ChangePasswordRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "成功修改密码",
			requestBody: models.ChangePasswordRequest{
				CurrentPassword: "oldpassword123",
				NewPassword:     "newpassword456",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "当前密码错误",
			requestBody: models.ChangePasswordRequest{
				CurrentPassword: "wrongpassword",
				NewPassword:     "newpassword456",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid current password",
		},
		{
			name: "新密码太短",
			requestBody: models.ChangePasswordRequest{
				CurrentPassword: "oldpassword123",
				NewPassword:     "123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid new password",
		},
		{
			name: "缺少当前密码",
			requestBody: models.ChangePasswordRequest{
				NewPassword: "newpassword456",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 准备请求
			reqBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PUT", "/password", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			// 执行请求
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// 验证响应
			if w.Code != tt.expectedStatus {
				t.Errorf("期望状态码 %d，实际得到 %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response models.SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("解析响应失败: %v", err)
				}
				if response.Message != "password changed successfully" {
					t.Errorf("期望消息 'password changed successfully'，实际得到 '%s'", response.Message)
				}

				// 验证密码确实被更改了
				var updatedUser models.User
				database.DB.Where("id = ?", user.ID).First(&updatedUser)

				// 旧密码不应该再有效
				if utils.CheckPasswordWithSalt("oldpassword123", updatedUser.Salt, updatedUser.Password) {
					t.Error("旧密码仍然有效，但应该无效")
				}

				// 新密码应该有效
				if !utils.CheckPasswordWithSalt("newpassword456", updatedUser.Salt, updatedUser.Password) {
					t.Error("新密码无效，但应该有效")
				}
			} else {
				var errorResponse models.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				if err != nil {
					t.Errorf("解析错误响应失败: %v", err)
				}
				if tt.expectedError != "" && errorResponse.Error != tt.expectedError {
					t.Errorf("期望错误 '%s'，实际得到 '%s'", tt.expectedError, errorResponse.Error)
				}
			}
		})
	}
}

func TestChangePasswordWithoutAuth(t *testing.T) {
	// 设置Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authHandler := NewAuthHandler()

	// 设置路由（需要认证）
	authenticated := router.Group("/")
	authenticated.Use(auth.JWTAuthMiddleware())
	authenticated.PUT("/password", authHandler.ChangePassword)

	// 准备请求（没有Authorization头）
	reqBody, _ := json.Marshal(models.ChangePasswordRequest{
		CurrentPassword: "oldpassword123",
		NewPassword:     "newpassword456",
	})
	req, _ := http.NewRequest("PUT", "/password", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证响应
	if w.Code != http.StatusUnauthorized {
		t.Errorf("期望状态码 %d，实际得到 %d", http.StatusUnauthorized, w.Code)
	}
}
