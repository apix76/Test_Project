package main

import (
	"TestProject/service/db"
	"TestProject/service/token"
	"TestProject/usecase"
	"context"
	"encoding/json"
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
	ExpTimeAccess   int
	ExpTimeRefresh  int
}

type HTTPHandler struct {
	usecase.UseCase
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	con := NewConf()
	useCase := usecase.UseCase{
		Token: token.Token{
			Key:            con.KeyVerification,
			ExpTimeAccess:  con.ExpTimeAccess,
			ExpTimeRefresh: con.ExpTimeRefresh,
		},
	}

	var err error
	useCase.DB, err = db.New(con.PgsqlNameServe)
	if err != nil {
		log.Fatal(err)
	}

	handler := HTTPHandler{useCase}

	mux := http.NewServeMux()
	mux.HandleFunc(con.GetPath, handler.ServeGet)
	mux.HandleFunc(con.RefreshPath, handler.ServeRefresh)
	if con.CertFile != "" || con.Keyfile != "" {
		go http.ListenAndServeTLS(con.Port, con.CertFile, con.Keyfile, mux)
	}
	http.ListenAndServe(con.Port, mux)
}

func (g HTTPHandler) ServeGet(w http.ResponseWriter, req *http.Request) {
	type Request struct {
		Guid string
	}

	type Response struct {
		RefreshToken string
		AccessToken  string
	}

	set := Request{}
	res := Response{}

	err := json.NewDecoder(req.Body).Decode(&set)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer req.Body.Close()

	if set.Guid == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		g.UseCase.UserIp = strings.Split(ip, ",")[0]
	}

	if ip := req.Header.Get("X-Real-IP"); ip != "" && g.UseCase.UserIp == "" {
		g.UseCase.UserIp = strings.Split(ip, ",")[0]
	}

	if g.UseCase.UserIp == "" {
		g.UseCase.UserIp = strings.Split(req.RemoteAddr, ":")[0]
	}

	res.AccessToken, res.RefreshToken, err = g.CreateSession(set.Guid)
	if err != nil {

	}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Fatal(err)
	}
}

func (r HTTPHandler) ServeRefresh(w http.ResponseWriter, req *http.Request) {
	type Request struct {
		RefreshToken string `json:"RefreshToken"`
		AccessToken  string `json:"AccessToken"`
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

	set.Ctx = context.Background()

	if ip := req.Header.Get("X-Forwarded-For"); ip != "" {
		r.UseCase.UserIp = strings.Split(ip, ",")[0]
	}

	if ip := req.Header.Get("X-Real-IP"); ip != "" && r.UseCase.UserIp == "" {
		r.UseCase.UserIp = strings.Split(ip, ",")[0]
	}

	if r.UseCase.UserIp == "" {
		r.UseCase.UserIp = strings.Split(req.RemoteAddr, ":")[0]
	}

	res.AccessToken, res.RefreshToken, err = r.UseCase.RefreshSession(set.AccessToken, set.RefreshToken)
	if err != nil {
		if _, err := w.Write([]byte(err.Error())); err != nil {
			log.Fatal(err)
		}
		return
	}
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
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
