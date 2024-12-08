package main

import (
	"TestProject/Db"
	"TestProject/Token"
	"TestProject/UseCase"
	"context"
	"encoding/json"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"net/http"
	"os"
	"strings"
)

type Conf struct {
	CertFile        string
	Keyfile         string
	Port            string
	GetPath         string
	RefreshPath     string
	KeyVerification string
	PgsqlNameServe  string
}

type GetToken struct {
	UseCase.UseCase
}

type RefreshToken struct {
	UseCase.UseCase
}

func main() {
	con := NewConf()
	useCase := UseCase.UseCase{
		DB: Db.DbAccess{
			PgsqlNameServe: con.PgsqlNameServe,
		},
		Token: Token.Token{
			Key: con.KeyVerification,
		},
	}
	err := useCase.DB.Connect()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle(con.GetPath, &GetToken{useCase})
	mux.Handle(con.RefreshPath, &RefreshToken{useCase})
	if con.CertFile != "" || con.Keyfile != "" {
		go http.ListenAndServeTLS(con.Port, con.CertFile, con.Keyfile, mux)
	}
	http.ListenAndServe(con.Port, mux)
}

func (g *GetToken) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	type Request struct {
		guid string
		Ctx  context.Context
	}

	type Response struct {
		RefreshToken string
		AccessToken  string
	}

	set := Request{}
	res := Response{}

	err := json.NewDecoder(req.Body).Decode(&set.guid)
	if err != nil {
		log.Fatal(err)
	}
	defer req.Body.Close()

	g.Ctx = set.Ctx
	g.Token.Guid = set.guid
	g.UserIp = strings.Split(req.RemoteAddr, ":")[0]

	res.AccessToken, res.RefreshToken = g.CreateSession()
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Fatal(err)
	}
}

func (r *RefreshToken) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	type Request struct {
		RefreshToken string
		AccessToken  string
		Ctx          context.Context
	}
	type Response struct {
		RefreshToken string
		AccessToken  string
	}

	set := Request{}
	res := Response{}

	err := json.NewDecoder(req.Body).Decode(&set)
	if err != nil {
		log.Fatal(err)
	}
	defer req.Body.Close()

	res.AccessToken, res.RefreshToken, err = r.UseCase.RefreshSession()
	if err != nil {
		log.Fatal(err)
	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Fatal(err)
	}
}

func NewConf() Conf {
	con := Conf{}

	FileCon, err := os.Open("config.cfg")
	if err != nil {
		log.Fatal(err)
	}
	defer FileCon.Close()

	err = json.NewDecoder(FileCon).Decode(&con)
	if err != nil {
		log.Fatal(err)
	}
	return con
}
