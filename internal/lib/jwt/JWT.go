package jwt

import (
	"complaint_server/internal/domain"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

func GenerateJWT(jwtSecret string, student *domain.Student) (string, error) {
	claims := jwt.MapClaims{
		"barcode": student.Barcode,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func EncodeJWT(jwtSecret string, tokenString string) (domain.Student, error) {
	var student domain.Student

	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return student, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return student, errors.New("invalid token claims")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return student, errors.New("invalid exp format")
	}

	if int64(exp) < time.Now().Unix() {
		return student, errors.New("token expired")
	}
	switch val := claims["barcode"].(type) {
	case float64:
		student.Barcode = int(val)
	case string:
		parsed, err := strconv.Atoi(val)
		if err != nil {
			return student, errors.New("invalid barcode string")
		}
		student.Barcode = parsed
	default:
		return student, errors.New("barcode not found or invalid type")
	}
	return student, nil
}
