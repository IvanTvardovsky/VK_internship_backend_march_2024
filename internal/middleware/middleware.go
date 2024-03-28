package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"strconv"
	"strings"
	"vk_march_backend/internal/structures"
)

func VerifyToken(tokenString string, secretKey []byte) (*jwt.Token, *jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, nil, err
	}

	if !token.Valid {
		return nil, nil, errors.New("token signature is invalid")
	}

	return token, &claims, nil
}

func VerifyAndGetInfoFromToken(c *gin.Context, cfg *structures.Config) (*structures.TokenInfo, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return &structures.TokenInfo{}, errors.New("authorization header is missing")
	}

	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || authHeaderParts[0] != "Bearer" {
		return &structures.TokenInfo{}, errors.New("invalid authorization header format")
	}

	tokenString := authHeaderParts[1]

	_, claims, err := VerifyToken(tokenString, []byte(cfg.Key.SecretKey))
	if err != nil {
		if err.Error() == "token signature is invalid" {
			return &structures.TokenInfo{}, errors.New("token signature is invalid")
		}
		return &structures.TokenInfo{}, errors.New("error verifying token")
	}

	idTemp, err := strconv.Atoi(fmt.Sprint((*claims)["user_id"]))
	if err != nil {
		return &structures.TokenInfo{}, errors.New("error converting user_id to integer")
	}

	expFloat, err := strconv.ParseFloat(fmt.Sprint((*claims)["exp"]), 64)
	if err != nil {
		return &structures.TokenInfo{}, errors.New("error converting exp to float")
	}

	expTemp := int(expFloat)

	tokenInfo := structures.TokenInfo{
		Login:   fmt.Sprintln((*claims)["login"]),
		ID:      idTemp,
		Expires: expTemp,
	}
	return &tokenInfo, nil
}
