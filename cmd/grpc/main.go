package main

import (
	"TestProject/ProtoDirectory"
	"TestProject/service/db"
	"TestProject/service/token"
	"TestProject/usecase"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

type Conf struct {
	//	CertFile        string
	//	Keyfile         string
	Port string
	//	GetPath         string
	//	RefreshPath     string
	KeyVerification string
	PgsqlNameServe  string
	ExpTimeAccess   int
	ExpTimeRefresh  int
}

func main() {
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

	l, err := net.Listen("tcp", con.Port)
	if err != nil {
		log.Fatal(err)
	}

	serv := ProtoDirectory.Server{UseCase: useCase}
	grpcServer := grpc.NewServer()
	ProtoDirectory.RegisterTokenServer(grpcServer, serv)

	fmt.Println("Start")
	err = grpcServer.Serve(l)
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
