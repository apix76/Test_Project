package usecase

import (
	"TestProject/service/db"
	"TestProject/service/token"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type ErrorsBody struct {
	Header string // "Contents_type/json"
	Body   string // "Errors: ..."
}
type UseCase struct {
	DB     db.DbAccess
	Token  token.Token
	UserIp string
}

func (u *UseCase) CreateSession(guid string) (string, string, error) {
	refreshToken, id := u.Token.CreateRefreshToken(u.UserIp, guid)
	RefreshBcrypt, err := bcrypt.GenerateFromPassword((u.Token.HashSHA256(refreshToken)), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	err = u.DB.Add(guid, string(RefreshBcrypt), id)
	if err != nil {
		return "", "", err
	}
	accessToken := u.Token.CreateAccessToken(u.UserIp, guid, id)
	RefreshBase64 := base64.StdEncoding.EncodeToString([]byte(refreshToken))
	return accessToken, RefreshBase64, nil
}

func (u *UseCase) RefreshSession(Access, Refresh string) (string, string, error) {
	RefreshByte, err := base64.StdEncoding.DecodeString(Refresh)
	if err != nil {
		return "", "", err
	}

	Refresh = string(RefreshByte)

	claimsRefresh, err := u.Token.Parse(Refresh)
	if err != nil {
		return "", "", err
	}
	claimsAccess, err := u.Token.Parse(Access)
	if err != nil {
		return "", "", err
	}
	guid := claimsRefresh.UserID
	if u.UserIp != claimsRefresh.IP {
		u.SendEmail(guid)
		err = errors.New("email warning")
		return "", "", err
	}

	OldRefreshBcrypt, err := u.DB.Check(claimsRefresh.ID)
	if OldRefreshBcrypt == "" {
		err = errors.New("Refresh token not exict")
		return "", "", err
	}

	if err := u.Token.Check(OldRefreshBcrypt, Refresh); err != nil {
		return "", "", err
	}

	if claimsAccess.UserID != claimsRefresh.UserID {
		err = errors.New("Token mismatch")
		return "", "", err
	}

	NewRefreshToken, id := u.Token.CreateRefreshToken(u.UserIp, guid)
	NewRefreshBcrypt, err := bcrypt.GenerateFromPassword(u.Token.HashSHA256(NewRefreshToken), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	err = u.DB.Refresh(OldRefreshBcrypt, string(NewRefreshBcrypt))
	if err != nil {
		return "", "", err
	}
	RefreshBase64 := base64.StdEncoding.EncodeToString([]byte(NewRefreshToken))
	NewAccessToken := u.Token.CreateAccessToken(u.UserIp, guid, id)

	return NewAccessToken, RefreshBase64, err
}

func (u *UseCase) SendEmail(guid string) error {
	log.Println("Sending email: ", u.DB.GetEmail(guid))
	return nil
}
