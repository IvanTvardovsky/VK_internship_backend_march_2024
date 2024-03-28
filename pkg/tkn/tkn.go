package tkn

import (
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"vk_march_backend/internal/structures"
	"vk_march_backend/internal/utils"
)

func GenerateToken(user *structures.User, cfg *structures.Config) (string, error) {
	expirationTime, err := utils.TokenExpiresTime(cfg.Token.Expires)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login":   user.Login,
		"user_id": user.ID,
		"exp":     expirationTime.Unix(),
	})

	tokenString, err := token.SignedString([]byte(cfg.Key.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
