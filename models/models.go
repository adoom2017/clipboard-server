package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CustomTime 自定义时间类型，支持多种格式解析
type CustomTime struct {
	time.Time
}

// UnmarshalJSON 自定义JSON解析，支持各种时间格式
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	if str == "null" || str == "" {
		return nil
	}

	// 尝试多种时间格式，按常见程度排序
	formats := []string{
		// Flutter/Dart 常用格式
		"2006-01-02T15:04:05.000000",     // 微秒格式（Dart默认）
		"2006-01-02T15:04:05.000",        // 毫秒格式
		"2006-01-02T15:04:05.000000Z",    // 带Z的微秒格式
		"2006-01-02T15:04:05.000Z",       // 带Z的毫秒格式
		"2006-01-02T15:04:05.000000000",  // 纳秒格式
		"2006-01-02T15:04:05.000000000Z", // 带Z的纳秒格式

		// 标准ISO 8601格式
		time.RFC3339Nano,       // RFC3339 with nanoseconds
		time.RFC3339,           // 标准RFC3339
		"2006-01-02T15:04:05Z", // 不带毫秒的Z格式
		"2006-01-02T15:04:05",  // 基础ISO格式

		// 带时区的格式
		"2006-01-02T15:04:05.000000Z07:00", // 微秒+时区
		"2006-01-02T15:04:05.000Z07:00",    // 毫秒+时区
		"2006-01-02T15:04:05Z07:00",        // 秒+时区
		"2006-01-02T15:04:05.000000-07:00", // 微秒+负时区
		"2006-01-02T15:04:05.000-07:00",    // 毫秒+负时区
		"2006-01-02T15:04:05-07:00",        // 秒+负时区

		// 其他常见格式
		"2006-01-02 15:04:05.000000", // 空格分隔的微秒
		"2006-01-02 15:04:05.000",    // 空格分隔的毫秒
		"2006-01-02 15:04:05",        // 空格分隔的秒
		"2006/01/02 15:04:05.000000", // 斜杠分隔的微秒
		"2006/01/02 15:04:05.000",    // 斜杠分隔的毫秒
		"2006/01/02 15:04:05",        // 斜杠分隔的秒

		// Unix时间戳
		"1136239445.000000", // Unix微秒时间戳
		"1136239445.000",    // Unix毫秒时间戳
		"1136239445",        // Unix秒时间戳
	}

	var lastErr error
	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			ct.Time = t
			return nil
		} else {
			lastErr = err
		}
	}

	// 如果都解析失败，尝试使用标准库的JSON解析
	if err := json.Unmarshal(data, &ct.Time); err == nil {
		return nil
	}

	// 所有方法都失败了，返回一个有意义的错误信息
	return fmt.Errorf("无法解析时间格式 '%s'，尝试了 %d 种格式，最后一个错误: %v", str, len(formats), lastErr)
}

// ClipboardType enum for clipboard types
type ClipboardType string

const (
	ClipboardTypeText  ClipboardType = "text"
	ClipboardTypeImage ClipboardType = "image"
	ClipboardTypeFile  ClipboardType = "file"
)

// ClipboardItem model
type ClipboardItem struct {
	ID        string        `json:"id" gorm:"primaryKey"`
	UserID    string        `json:"user_id" gorm:"index"`
	ClientID  string        `json:"client_id" gorm:"index"` // 客户端唯一ID
	Content   string        `json:"content" gorm:"type:text"`
	Type      ClipboardType `json:"type" gorm:"type:varchar(20);default:'text'"`
	Timestamp time.Time     `json:"timestamp" gorm:"index"`
	CreatedAt time.Time     `json:"created_at" gorm:"autoCreateTime:nano"`
	UpdatedAt time.Time     `json:"updated_at" gorm:"autoUpdateTime:nano"`
}

// User model
type User struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;size:100"`
	Email     string    `json:"email" gorm:"uniqueIndex;size:255"`
	Password  string    `json:"-" gorm:"size:255"` // Hidden in JSON response
	Salt      string    `json:"-" gorm:"size:32"`  // Salt for password hashing, hidden in JSON
	Token     string    `json:"token,omitempty" gorm:"size:500"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Associated clipboard items
	ClipboardItems []ClipboardItem `json:"clipboard_items,omitempty" gorm:"foreignKey:UserID"`
}

// BeforeCreate hook to set ID and timestamp
func (c *ClipboardItem) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	if c.Timestamp.IsZero() {
		c.Timestamp = time.Now()
	}
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	c.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate hook to update timestamp
func (c *ClipboardItem) BeforeUpdate(tx *gorm.DB) error {
	c.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate hook to set ID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// TableName custom table names
func (ClipboardItem) TableName() string {
	return "clipboard_items"
}

func (User) TableName() string {
	return "users"
}

// ClipboardItemRequest for creating clipboard items
type ClipboardItemRequest struct {
	Content   string        `json:"content" binding:"required"`
	Type      ClipboardType `json:"type" binding:"omitempty"`
	Timestamp *CustomTime   `json:"timestamp"`
}

// ClipboardItemResponse response structure
type ClipboardItemResponse struct {
	ID        string        `json:"id"`
	Content   string        `json:"content"`
	Type      ClipboardType `json:"type"`
	Timestamp time.Time     `json:"timestamp"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// ToResponse converts to response structure
func (c *ClipboardItem) ToResponse() ClipboardItemResponse {
	return ClipboardItemResponse{
		ID:        c.ID,
		Content:   c.Content,
		Type:      c.Type,
		Timestamp: c.Timestamp,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// BatchSyncRequest for batch sync
type BatchSyncRequest struct {
	DeviceID string                 `json:"device_id"`
	Items    []ClipboardItemRequest `json:"items" binding:"required"`
}

// BatchSyncResponse for batch sync response
type BatchSyncResponse struct {
	Synced []ClipboardItemResponse `json:"synced"`
	Failed []FailedItem            `json:"failed"`
	Total  int                     `json:"total"`
}

// FailedItem for sync failures
type FailedItem struct {
	Content string `json:"content"`
	Error   string `json:"error"`
}

// LoginRequest for login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest for registration
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// ChangePasswordRequest for changing password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// LoginResponse for login response
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// ErrorResponse for errors
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

// SuccessResponse for success
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationQuery for pagination
type PaginationQuery struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Since    string `form:"since"`  // ISO 8601 time format
	Type     string `form:"type"`   // Filter by type
	Search   string `form:"search"` // Search content
}

// PaginationResponse for pagination response
type PaginationResponse struct {
	Items      []ClipboardItemResponse `json:"items"`
	Total      int64                   `json:"total"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
	TotalPages int                     `json:"total_pages"`
	HasNext    bool                    `json:"has_next"`
	HasPrev    bool                    `json:"has_prev"`
}

// StatisticsResponse for statistics
type StatisticsResponse struct {
	TotalItems       int64            `json:"total_items"`
	SyncedItems      int64            `json:"synced_items"`
	UnsyncedItems    int64            `json:"unsynced_items"`
	TotalContentSize int64            `json:"total_content_size"`
	TypeDistribution map[string]int64 `json:"type_distribution"`
	RecentActivity   []DailyActivity  `json:"recent_activity"`
}

// DailyActivity for daily activity stats
type DailyActivity struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// RecentSyncResponse for recent sync clipboard items
type RecentSyncResponse struct {
	Items []ClipboardItemResponse `json:"items"`
	Total int64                   `json:"total"`
}
