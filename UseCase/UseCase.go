package UseCase

import (
	"TestProject/Db"
	"TestProject/Token"
	"context"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type UseCase struct {
	DB     Db.DbAccess
	Token  Token.Token
	Ctx    context.Context
	UserIp string
}

func (u *UseCase) CreateSession() (string, string) {
	refreshToken, id := u.Token.CreateRefreshToken(u.UserIp)
	RefreshBcrypt, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	u.DB.Add(u.Ctx, u.Token.Guid, string(RefreshBcrypt), id)
	accessToken := u.Token.CreateAccessToken(u.UserIp, id)
	RefreshBase64 := base64.StdEncoding.EncodeToString([]byte(refreshToken))
	return accessToken, RefreshBase64
}

func (u *UseCase) RefreshSession() (string, string, error) {
	claimsRefresh, err := u.Token.Parse(u.Token.Refresh)
	if err != nil {
		return "", "", err
	}
	claimsAccess, err := u.Token.Parse(u.Token.Access)
	if err != nil {
		return "", "", err
	}

	if u.UserIp != claimsRefresh.IP {
		u.SendEmail()
		err = errors.New("email warning")
		return "", "", err
	}

	OldRefreshBcrypt := u.DB.Check(u.Ctx, claimsRefresh.ID)
	if OldRefreshBcrypt == "" {
		err = errors.New("Refresh token not exict")
		return "", "", err
	}

	if err := u.Token.Check(OldRefreshBcrypt); err != nil {
		return "", "", err
	}

	if claimsAccess.UserID != claimsRefresh.UserID {
		err = errors.New("Token mismatch")
		return "", "", err
	}

	u.Token.Guid = claimsRefresh.UserID
	NewRefreshToken, id := u.Token.CreateRefreshToken(u.UserIp)
	NewRefreshBcrypt, err := bcrypt.GenerateFromPassword([]byte(NewRefreshToken), bcrypt.DefaultCost)
	u.DB.Refresh(u.Ctx, OldRefreshBcrypt, string(NewRefreshBcrypt))
	RefreshBase64 := base64.StdEncoding.EncodeToString([]byte(NewRefreshToken))
	NewAccessToken := u.Token.CreateAccessToken(u.UserIp, id)

	return NewAccessToken, RefreshBase64, err
}

func (u *UseCase) SendEmail() error {
	log.Println("Sending email: ", u.DB.GetEmail(u.Ctx, u.Token.Guid))
	return nil
}
