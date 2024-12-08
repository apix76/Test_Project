package Token

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log"
	"math/rand"
	"time"
)

type Token struct {
	Guid           string
	Refresh        string
	Access         string
	Key            string
	ExpTimeAccess  int
	ExpTimeRefresh int
}

type Claims struct {
	ID     string
	UserID string
	IP     string
}

func (t *Token) CreateRefreshToken(ip string) (string, int64) {
	idSession := rand.Int63()
	RefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"exp":      time.Now().Add(time.Duration(t.ExpTimeRefresh) * time.Minute),
		"jti":      idSession,
		"guid":     t.Guid,
		"ipClient": ip,
	})
	token, err := RefreshToken.SignedString(t.Key)
	if err != nil {
		log.Fatal(err)
	}
	return token, idSession
}

func (t *Token) CreateAccessToken(ip string, idSession int64) string {
	AccessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"exp":      time.Now().Add(time.Duration(t.ExpTimeAccess) * time.Minute),
		"jti":      idSession,
		"guid":     t.Guid,
		"ipClient": ip,
	})
	token, err := AccessToken.SignedString(t.Key)
	if err != nil {
		log.Fatal(err)
	}
	return token
}

func (t *Token) Check(token string) error {
	err := bcrypt.CompareHashAndPassword([]byte(token), []byte(t.Refresh))
	return err
}

func (t *Token) Parse(token string) (Claims, error) {
	Token, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.Key), nil
	})
	if err != nil {
		log.Fatal(err)
	}

	claims, ok := Token.Claims.(jwt.MapClaims)
	if !ok {
		return Claims{}, errors.New("Empty claims")
	}
	claim := Claims{
		ID:     claims["jti"].(string),
		UserID: claims["guid"].(string),
		IP:     claims["ipClient"].(string),
	}
	return claim, err
}
