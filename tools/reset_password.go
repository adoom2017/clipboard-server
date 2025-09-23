package main

import (
	"clipboard-server/database"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法:")
		fmt.Println("  go run tools/reset_password.go <username> <new_password>")
		fmt.Println("例子:")
		fmt.Println("  go run tools/reset_password.go admin newpassword123")
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Println("错误: 请提供用户名和新密码")
		os.Exit(1)
	}

	username := os.Args[1]
	newPassword := os.Args[2]

	// 初始化数据库
	if err := database.Initialize(); err != nil {
		fmt.Printf("数据库初始化失败: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// 重置密码
	if err := database.ResetUserPasswordWithSalt(username, newPassword); err != nil {
		fmt.Printf("密码重置失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("用户 %s 的密码已成功重置\n", username)
}
