package utils

import (
	"strings"
	"testing"
)

func TestGenerateSalt(t *testing.T) {
	salt1, err := GenerateSalt()
	if err != nil {
		t.Fatalf("生成盐值失败: %v", err)
	}

	salt2, err := GenerateSalt()
	if err != nil {
		t.Fatalf("生成盐值失败: %v", err)
	}

	// 盐值应该是64个字符（32字节的十六进制表示）
	if len(salt1) != 64 {
		t.Errorf("盐值长度应该是64，实际是 %d", len(salt1))
	}

	// 两次生成的盐值应该不同
	if salt1 == salt2 {
		t.Error("两次生成的盐值不应该相同")
	}

	// 盐值应该只包含十六进制字符
	for _, char := range salt1 {
		if !strings.ContainsRune("0123456789abcdef", char) {
			t.Errorf("盐值包含非十六进制字符: %c", char)
		}
	}
}

func TestCheckPasswordWithSalt(t *testing.T) {
	password := "testpassword123"
	wrongPassword := "wrongpassword"
	salt := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	// 生成哈希
	hash, err := HashPasswordWithSalt(password, salt)
	if err != nil {
		t.Fatalf("密码哈希失败: %v", err)
	}

	// 正确密码应该验证成功
	if !CheckPasswordWithSalt(password, salt, hash) {
		t.Error("正确密码验证失败")
	}

	// 错误密码应该验证失败
	if CheckPasswordWithSalt(wrongPassword, salt, hash) {
		t.Error("错误密码验证应该失败")
	}

	// 错误盐值应该验证失败
	wrongSalt := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	if CheckPasswordWithSalt(password, wrongSalt, hash) {
		t.Error("错误盐值验证应该失败")
	}
}

func TestBackwardCompatibility(t *testing.T) {
	password := "testpassword123"

	// 使用旧方法生成哈希
	oldHash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("旧方法哈希失败: %v", err)
	}

	// 旧方法应该能验证
	if !CheckPassword(password, oldHash) {
		t.Error("旧方法密码验证失败")
	}

	// 测试新旧方法的哈希应该不同
	salt, _ := GenerateSalt()
	newHash, err := HashPasswordWithSalt(password, salt)
	if err != nil {
		t.Fatalf("新方法哈希失败: %v", err)
	}

	if oldHash == newHash {
		t.Error("新旧方法的哈希不应该相同")
	}
}

func BenchmarkHashPasswordWithSalt(b *testing.B) {
	password := "testpassword123"
	salt := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashPasswordWithSalt(password, salt)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCheckPasswordWithSalt(b *testing.B) {
	password := "testpassword123"
	salt := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	hash, _ := HashPasswordWithSalt(password, salt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckPasswordWithSalt(password, salt, hash)
	}
}

func BenchmarkGenerateSalt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateSalt()
		if err != nil {
			b.Fatal(err)
		}
	}
}
