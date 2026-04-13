package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateAccessToken(userID, role, branchID, secret string) (string, int64, error) {
	expiresIn := int64(3600) // 1 hour
	claims := jwt.MapClaims{
		"sub":       userID,
		"role":      role,
		"branch_id": branchID,
		"exp":       time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
		"iat":       time.Now().Unix(),
		"type":      "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}

	return tokenStr, expiresIn, nil
}

func GenerateRefreshToken(userID, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days
		"iat":  time.Now().Unix(),
		"type": "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ValidateToken(tokenStr, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}
