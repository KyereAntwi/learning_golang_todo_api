package main

import (
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type JWTManager struct {
	config *config
}

type CustomClaims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

var issuer string
var key []byte

func NewJWTManager(config *config) *JWTManager {
	issuer = "http://" + config.appHost + ":" + strconv.Itoa(config.port)
	key = []byte(config.jwtSecret)
	return &JWTManager{config: config}
}

func (jm *JWTManager) GenerateAccessToken(userID string) (string, error) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
		},
	})

	tokenString, err := jwtToken.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (jm *JWTManager) GenerateRefreshToken(userID string) (string, time.Time, error) {

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    issuer,
		},
	})

	tokenString, err := jwtToken.SignedString(key)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (jm *JWTManager) IsAccessToken(tokenString string) (bool, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return key, nil
	})

	if err != nil {
		return false, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid || claims.TokenType != "access" {
		return false, nil
	}

	return true, nil
}

func (jm *JWTManager) HashRefreshToken(tokenString string) (string, error) {
	// SHA-256 the token first to get a fixed 32-byte input, since bcrypt has a 72-byte limit
	// make sure to use the raw token string for hashing, not the JWT claims or any other representation
	// use hte SHA-256 the raw token before comparing for verification in the future, since bcrypt has a 72-byte limit and the JWT token can be longer than that

	sha := sha256.Sum256([]byte(tokenString))
	hashedToken, err := bcrypt.GenerateFromPassword(sha[:], bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hashedToken), nil
}

func (jm *JWTManager) GetUserIDFromToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return key, nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return "", jwt.ErrInvalidKey
	}

	return claims.Subject, nil
}
