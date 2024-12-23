package token

import (
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log"
	"math/rand"
	"os"
	"time"
)

type Token struct {
	ExpTimeAccess  int
	ExpTimeRefresh int
	PublicKey      *rsa.PublicKey
	PrivateKey     *rsa.PrivateKey
}

type Claims struct {
	ID     string
	UserID string
	IP     string
}

func TokenKey(private, public string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privatekey, err := ParsePrivateKey(private)
	if err != nil {
		return nil, nil, err
	}
	publickey, err := ParsePublicKey(public)
	if err != nil {
		return nil, nil, err
	}
	return privatekey, publickey, nil
}

func (t *Token) CreateRefreshToken(ip, guid string) (string, string) {
	idSessionInt := rand.Int63()
	idSession := fmt.Sprintf("%v", idSessionInt)

	RefreshToken := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"exp":      jwt.NewNumericDate(time.Now().Add(time.Duration(t.ExpTimeRefresh) * time.Minute)),
		"jti":      idSession,
		"guid":     guid,
		"ipClient": ip,
	})
	token, err := RefreshToken.SignedString(t.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	return token, idSession
}

func (t *Token) CreateAccessToken(ip, guid, idSession string) string {
	AccessToken := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"exp":      jwt.NewNumericDate(time.Now().Add(time.Duration(t.ExpTimeAccess) * time.Minute)),
		"jti":      idSession,
		"guid":     guid,
		"ipClient": ip,
	})
	token, err := AccessToken.SignedString(t.PrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	return token
}

func (t *Token) Check(hashtoken, Refresh string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashtoken), t.HashSHA256(Refresh))
	return err
}

func (t *Token) Parse(token string) (Claims, error) {
	tokenParse, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			err := errors.New("Invalid signing method. Expected RSA")
			return nil, err
		}
		return t.PublicKey, nil
	})
	if err != nil {
		return Claims{}, err
	}

	claims, ok := tokenParse.Claims.(jwt.MapClaims)
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

func (t *Token) HashSHA256(str string) []byte {
	hash := sha256.Sum256([]byte(str))
	return hash[:]
}

func ParsePrivateKey(private string) (*rsa.PrivateKey, error) {
	filecon, err := os.ReadFile(private)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	privatekey, err := jwt.ParseRSAPrivateKeyFromPEM(filecon)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return privatekey, err
}

func ParsePublicKey(public string) (*rsa.PublicKey, error) {
	filecon, err := os.ReadFile(public)
	if err != nil {
		return nil, err
	}

	publickey, err := jwt.ParseRSAPublicKeyFromPEM(filecon)
	if err != nil {
		return nil, err
	}

	return publickey, err
}
