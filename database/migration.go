package database

import (
	"clipboard-server/models"
	"clipboard-server/utils"
	"fmt"
)

// MigrateExistingUsers 为现有用户添加盐值并重新哈希密码
func MigrateExistingUsers() error {
	fmt.Println("开始迁移现有用户的密码...")

	var users []models.User
	if err := DB.Find(&users).Error; err != nil {
		return fmt.Errorf("failed to fetch users: %v", err)
	}

	for _, user := range users {
		// 跳过已经有盐值的用户
		if user.Salt != "" {
			fmt.Printf("用户 %s 已经有盐值，跳过\n", user.Username)
			continue
		}

		// 生成新的盐值
		salt, err := utils.GenerateSalt()
		if err != nil {
			fmt.Printf("为用户 %s 生成盐值失败: %v\n", user.Username, err)
			continue
		}

		// 对于现有用户，我们需要假设他们的密码是用旧方法(bcrypt without custom salt)哈希的
		// 这种情况下，我们不能恢复原始密码，所以需要用户重新设置密码
		// 或者，如果你知道有一些测试用户，可以为他们设置默认密码

		// 更新用户记录
		user.Salt = salt
		if err := DB.Save(&user).Error; err != nil {
			fmt.Printf("更新用户 %s 失败: %v\n", user.Username, err)
			continue
		}

		fmt.Printf("用户 %s 迁移成功，添加了盐值\n", user.Username)
	}

	fmt.Println("用户迁移完成")
	return nil
}

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
