package handlers

import (
	"clipboard-server/auth"
	"clipboard-server/config"
	"clipboard-server/database"
	"clipboard-server/models"
	"clipboard-server/utils"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ClipboardHandler for clipboard related handlers
type ClipboardHandler struct{}

// NewClipboardHandler creates clipboard handler instance
func NewClipboardHandler() *ClipboardHandler {
	return &ClipboardHandler{}
}

// CreateItem creates clipboard item
func (h *ClipboardHandler) CreateItem(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "user not authenticated",
		})
		return
	}

	var req models.ClipboardItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	// Validate content size
	cfg := config.GetConfig()
	if utils.GetContentSize(req.Content) > cfg.MaxContentSize {
		c.JSON(http.StatusRequestEntityTooLarge, models.ErrorResponse{
			Error:   "content too large",
			Message: "content size exceeds limit",
		})
		return
	}

	// Validate content type
	if req.Type != "" && !utils.IsValidContentType(string(req.Type)) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid content type",
			Message: "unsupported content type",
		})
		return
	}

	// Default to text if no type specified
	if req.Type == "" {
		req.Type = models.ClipboardTypeText
	}

	// Sanitize sensitive content
	sanitizedContent := utils.SanitizeContent(req.Content)

	// Create clipboard item
	item := models.ClipboardItem{
		UserID:  userID,
		Content: sanitizedContent,
		Type:    req.Type,
	}

	// Use provided timestamp or current time
	if req.Timestamp != nil {
		item.Timestamp = req.Timestamp.Time
	} else {
		item.Timestamp = time.Now()
	}

	db := database.GetDB()
	if err := db.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "creation failed",
			Message: "failed to create clipboard item",
		})
		return
	}

	c.JSON(http.StatusCreated, item.ToResponse())
}

// GetItems gets clipboard items list
func (h *ClipboardHandler) GetItems(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "user not authenticated",
		})
		return
	}

	var query models.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid query parameters",
			Message: err.Error(),
		})
		return
	}

	// Set default values
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	db := database.GetDB()

	// Build query
	dbQuery := db.Model(&models.ClipboardItem{}).Where("user_id = ?", userID)

	// Time filter
	if query.Since != "" {
		if sinceTime, err := time.Parse(time.RFC3339, query.Since); err == nil {
			dbQuery = dbQuery.Where("timestamp >= ?", sinceTime)
		}
	}

	// Type filter
	if query.Type != "" && utils.IsValidContentType(query.Type) {
		dbQuery = dbQuery.Where("type = ?", query.Type)
	}

	// Content search
	if query.Search != "" {
		dbQuery = dbQuery.Where("content LIKE ?", "%"+query.Search+"%")
	}

	// Get total count
	var total int64
	dbQuery.Count(&total)

	// Paginated query
	var items []models.ClipboardItem
	offset := (query.Page - 1) * query.PageSize

	if err := dbQuery.Order("timestamp DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "query failed",
			Message: "failed to query clipboard items",
		})
		return
	}

	// Convert to response format
	responseItems := make([]models.ClipboardItemResponse, len(items))
	for i, item := range items {
		responseItems[i] = item.ToResponse()
	}

	// Calculate pagination info
	totalPages := int(total) / query.PageSize
	if int(total)%query.PageSize > 0 {
		totalPages++
	}

	response := models.PaginationResponse{
		Items:      responseItems,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
		HasNext:    query.Page < totalPages,
		HasPrev:    query.Page > 1,
	}

	c.JSON(http.StatusOK, response)
}

// GetItem gets single clipboard item
func (h *ClipboardHandler) GetItem(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "user not authenticated",
		})
		return
	}

	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid request",
			Message: "item ID is required",
		})
		return
	}

	db := database.GetDB()
	var item models.ClipboardItem

	if err := db.Where("id = ? AND user_id = ?", itemID, userID).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "item not found",
				Message: "clipboard item not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "query failed",
			Message: "failed to get clipboard item",
		})
		return
	}

	c.JSON(http.StatusOK, item.ToResponse())
}

// UpdateItem updates clipboard item
func (h *ClipboardHandler) UpdateItem(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "user not authenticated",
		})
		return
	}

	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid request",
			Message: "item ID is required",
		})
		return
	}

	var req models.ClipboardItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	// Validate content size
	cfg := config.GetConfig()
	if utils.GetContentSize(req.Content) > cfg.MaxContentSize {
		c.JSON(http.StatusRequestEntityTooLarge, models.ErrorResponse{
			Error:   "content too large",
			Message: "content size exceeds limit",
		})
		return
	}

	db := database.GetDB()
	var item models.ClipboardItem

	// Find item
	if err := db.Where("id = ? AND user_id = ?", itemID, userID).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "item not found",
				Message: "clipboard item not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "query failed",
			Message: "failed to get clipboard item",
		})
		return
	}

	// Update fields
	item.Content = utils.SanitizeContent(req.Content)
	if req.Type != "" && utils.IsValidContentType(string(req.Type)) {
		item.Type = req.Type
	}
	if req.Timestamp != nil {
		item.Timestamp = req.Timestamp.Time
	}

	// Save update
	if err := db.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update failed",
			Message: "failed to update clipboard item",
		})
		return
	}

	c.JSON(http.StatusOK, item.ToResponse())
}

// DeleteItem deletes clipboard item
func (h *ClipboardHandler) DeleteItem(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "user not authenticated",
		})
		return
	}

	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid request",
			Message: "item ID is required",
		})
		return
	}

	db := database.GetDB()

	// Delete item (ensure only own items can be deleted)
	result := db.Where("id = ? AND user_id = ?", itemID, userID).Delete(&models.ClipboardItem{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "deletion failed",
			Message: "failed to delete clipboard item",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "item not found",
			Message: "clipboard item not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "clipboard item deleted successfully",
	})
}

// BatchSync batch sync clipboard items
func (h *ClipboardHandler) BatchSync(c *gin.Context) {
	log.Printf("[BatchSync] 开始处理批量同步请求")

	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		log.Printf("[BatchSync] 用户未认证")
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "user not authenticated",
		})
		return
	}
	log.Printf("[BatchSync] 用户ID: %s", userID)

	var req models.BatchSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[BatchSync] JSON解析失败: %v", err)
		log.Printf("[BatchSync] 原始请求体长度: %d", c.Request.ContentLength)
		if body, readErr := c.GetRawData(); readErr == nil {
			log.Printf("[BatchSync] 原始请求体前500字符: %s", utils.TruncateString(string(body), 500))
		}
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	log.Printf("[BatchSync] 成功解析请求，设备ID: %s, 项目数量: %d", req.DeviceID, len(req.Items))

	if len(req.Items) == 0 {
		log.Printf("[BatchSync] 请求中没有项目")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "empty request",
			Message: "no items to sync",
		})
		return
	}

	cfg := config.GetConfig()
	db := database.GetDB()
	log.Printf("[BatchSync] 配置最大内容大小: %d 字节", cfg.MaxContentSize)

	var synced []models.ClipboardItemResponse
	var failed []models.FailedItem

	// Process each item in batch
	for i, itemReq := range req.Items {
		log.Printf("[BatchSync] 处理第 %d 个项目，内容长度: %d, 类型: %s",
			i+1, len(itemReq.Content), itemReq.Type)
		log.Printf("[BatchSync] 项目内容前50字符: %s",
			utils.TruncateString(itemReq.Content, 50))

		// Validate content size
		contentSize := utils.GetContentSize(itemReq.Content)
		if contentSize > cfg.MaxContentSize {
			log.Printf("[BatchSync] 项目 %d 内容过大: %d > %d", i+1, contentSize, cfg.MaxContentSize)
			failed = append(failed, models.FailedItem{
				Content: utils.TruncateString(itemReq.Content, 50),
				Error:   "content too large",
			})
			continue
		}

		// Validate content type
		if itemReq.Type != "" && !utils.IsValidContentType(string(itemReq.Type)) {
			log.Printf("[BatchSync] 项目 %d 类型无效: %s", i+1, itemReq.Type)
			failed = append(failed, models.FailedItem{
				Content: utils.TruncateString(itemReq.Content, 50),
				Error:   "invalid content type",
			})
			continue
		}

		// Set default type
		if itemReq.Type == "" {
			itemReq.Type = models.ClipboardTypeText
			log.Printf("[BatchSync] 项目 %d 设置默认类型: text", i+1)
		}

		// Create item
		item := models.ClipboardItem{
			UserID:  userID,
			Content: utils.SanitizeContent(itemReq.Content),
			Type:    itemReq.Type,
		}

		if itemReq.Timestamp != nil {
			item.Timestamp = itemReq.Timestamp.Time
			log.Printf("[BatchSync] 项目 %d 使用提供的时间戳: %v", i+1, itemReq.Timestamp.Time)
		} else {
			item.Timestamp = time.Now()
			log.Printf("[BatchSync] 项目 %d 使用当前时间戳", i+1)
		}

		log.Printf("[BatchSync] 尝试保存项目 %d 到数据库", i+1)
		if err := db.Create(&item).Error; err != nil {
			log.Printf("[BatchSync] 项目 %d 数据库保存失败: %v", i+1, err)
			failed = append(failed, models.FailedItem{
				Content: utils.TruncateString(itemReq.Content, 50),
				Error:   "database error",
			})
			continue
		}

		log.Printf("[BatchSync] 项目 %d 成功保存，ID: %s", i+1, item.ID)
		synced = append(synced, item.ToResponse())
	}

	log.Printf("[BatchSync] 批量同步完成，成功: %d, 失败: %d, 总计: %d",
		len(synced), len(failed), len(req.Items))

	response := models.BatchSyncResponse{
		Synced: synced,
		Failed: failed,
		Total:  len(req.Items),
	}

	c.JSON(http.StatusOK, response)
}

// GetStatistics gets clipboard statistics
func (h *ClipboardHandler) GetStatistics(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "user not authenticated",
		})
		return
	}

	db := database.GetDB()

	// Total items
	var totalItems int64
	db.Model(&models.ClipboardItem{}).Where("user_id = ?", userID).Count(&totalItems)

	// All items stored on server are considered synced
	syncedItems := totalItems

	// Unsynced items (always 0 for server items)
	unsyncedItems := int64(0)

	// Total content size
	var totalContentSize int64
	db.Model(&models.ClipboardItem{}).
		Where("user_id = ?", userID).
		Select("SUM(LENGTH(content))").
		Scan(&totalContentSize)

	// Type distribution
	typeDistribution := make(map[string]int64)
	rows, err := db.Model(&models.ClipboardItem{}).
		Select("type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("type").Rows()

	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var contentType string
			var count int64
			if err := rows.Scan(&contentType, &count); err == nil {
				typeDistribution[contentType] = count
			}
		}
	}

	// Recent 7 days activity
	var recentActivity []models.DailyActivity
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	activityRows, err := db.Model(&models.ClipboardItem{}).
		Select("DATE(timestamp) as date, COUNT(*) as count").
		Where("user_id = ? AND timestamp >= ?", userID, sevenDaysAgo).
		Group("DATE(timestamp)").
		Order("date DESC").Rows()

	if err == nil {
		defer activityRows.Close()
		for activityRows.Next() {
			var activity models.DailyActivity
			if err := activityRows.Scan(&activity.Date, &activity.Count); err == nil {
				recentActivity = append(recentActivity, activity)
			}
		}
	}

	stats := models.StatisticsResponse{
		TotalItems:       totalItems,
		SyncedItems:      syncedItems,
		UnsyncedItems:    unsyncedItems,
		TotalContentSize: totalContentSize,
		TypeDistribution: typeDistribution,
		RecentActivity:   recentActivity,
	}

	c.JSON(http.StatusOK, stats)
}

// GetRecentSyncItems gets recently synced clipboard items
func (h *ClipboardHandler) GetRecentSyncItems(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "user not authenticated",
		})
		return
	}

	db := database.GetDB()

	// Get limit parameter (default 10, max 50)
	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			if parsedLimit > 50 {
				limit = 50
			} else {
				limit = parsedLimit
			}
		}
	}

	// Get recent items ordered by created_at desc
	var items []models.ClipboardItem
	result := db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&items)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database error",
			Message: "failed to fetch recent items",
		})
		return
	}

	// Get total count
	var totalCount int64
	db.Model(&models.ClipboardItem{}).Where("user_id = ?", userID).Count(&totalCount)

	// Convert to response format
	responseItems := make([]models.ClipboardItemResponse, len(items))
	for i, item := range items {
		responseItems[i] = item.ToResponse()
	}

	response := models.RecentSyncResponse{
		Items: responseItems,
		Total: totalCount,
	}

	c.JSON(http.StatusOK, response)
}

// GetLatestSyncItem gets the latest synced clipboard item
func (h *ClipboardHandler) GetLatestSyncItem(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "user not authenticated",
		})
		return
	}

	db := database.GetDB()

	// Get the latest item ordered by updated_at desc
	var item models.ClipboardItem
	result := db.Where("user_id = ?", userID).
		Order("updated_at DESC").
		First(&item)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "no data found",
				Message: "no clipboard items found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database error",
			Message: "failed to fetch latest item",
		})
		return
	}

	response := item.ToResponse()
	c.JSON(http.StatusOK, response)
}

// SyncSingleItem syncs a single clipboard item by client ID
func (h *ClipboardHandler) SyncSingleItem(c *gin.Context) {
	userID, exists := auth.GetCurrentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "user not authenticated",
		})
		return
	}

	type SyncSingleItemRequest struct {
		ClientID  string               `json:"client_id" binding:"required"`
		Content   string               `json:"content" binding:"required"`
		Type      models.ClipboardType `json:"type"`
		Timestamp *models.CustomTime   `json:"timestamp"`
	}

	var req SyncSingleItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid request",
			Message: err.Error(),
		})
		return
	}

	// Validate content size
	cfg := config.GetConfig()
	if utils.GetContentSize(req.Content) > cfg.MaxContentSize {
		c.JSON(http.StatusRequestEntityTooLarge, models.ErrorResponse{
			Error:   "content too large",
			Message: "content size exceeds limit",
		})
		return
	}

	// Default to text if no type specified
	if req.Type == "" {
		req.Type = models.ClipboardTypeText
	}

	// Validate content type
	if !utils.IsValidContentType(string(req.Type)) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid content type",
			Message: "unsupported content type",
		})
		return
	}

	// Sanitize sensitive content
	sanitizedContent := utils.SanitizeContent(req.Content)

	db := database.GetDB()

	// Check if item already exists with this client_id for this user
	var existingItem models.ClipboardItem
	err := db.Where("user_id = ? AND client_id = ?", userID, req.ClientID).First(&existingItem).Error

	timestamp := time.Now()
	if req.Timestamp != nil {
		timestamp = req.Timestamp.Time
	}

	if err == gorm.ErrRecordNotFound {
		// Create new item
		item := models.ClipboardItem{
			UserID:    userID,
			ClientID:  req.ClientID,
			Content:   sanitizedContent,
			Type:      req.Type,
			Timestamp: timestamp,
		}

		if err := db.Create(&item).Error; err != nil {
			log.Printf("[SyncSingleItem] 创建失败: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "creation failed",
				Message: "failed to create clipboard item",
			})
			return
		}

		log.Printf("[SyncSingleItem] 创建新记录: client_id=%s, user_id=%s", req.ClientID, userID)
		c.JSON(http.StatusCreated, item.ToResponse())
	} else if err != nil {
		// Database error
		log.Printf("[SyncSingleItem] 数据库错误: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "database error",
			Message: "failed to query clipboard item",
		})
		return
	} else {
		// Update existing item (only update timestamp and content)
		existingItem.Timestamp = timestamp
		existingItem.Content = sanitizedContent
		existingItem.Type = req.Type

		if err := db.Save(&existingItem).Error; err != nil {
			log.Printf("[SyncSingleItem] 更新失败: %v", err)
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Error:   "update failed",
				Message: "failed to update clipboard item",
			})
			return
		}

		log.Printf("[SyncSingleItem] 更新现有记录: client_id=%s, user_id=%s", req.ClientID, userID)
		c.JSON(http.StatusOK, existingItem.ToResponse())
	}
}
