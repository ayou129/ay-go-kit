package cryptox

import "golang.org/x/crypto/bcrypt"

// DefaultBcryptCost bcrypt 加密强度
const DefaultBcryptCost = 12

// HashPassword 使用 bcrypt 加密密码
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), DefaultBcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword 验证密码是否匹配
func CheckPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
