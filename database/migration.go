package database

import (
	"clipboard-server/models"
	"clipboard-server/utils"
	"fmt"
)

// ResetUserPasswordWithSalt 为用户重置密码（使用新的盐值哈希方法）
func ResetUserPasswordWithSalt(username, newPassword string) error {
	var user models.User
	if err := DB.Where("username = ?", username).First(&user).Error; err != nil {
		return fmt.Errorf("user not found: %v", err)
	}

	// 生成新盐值
	salt, err := utils.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %v", err)
	}

	// 使用盐值哈希密码
	hashedPassword, err := utils.HashPasswordWithSalt(newPassword, salt)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	// 更新用户
	user.Salt = salt
	user.Password = hashedPassword
	if err := DB.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	fmt.Printf("用户 %s 的密码已重置\n", username)
	return nil
}
