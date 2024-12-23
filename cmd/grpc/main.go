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
	PublicKey      string
	PrivateKey     string
	PgsqlNameServe string
	ExpTimeAccess  int
	ExpTimeRefresh int
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	con := NewConf()
	useCase := usecase.UseCase{
		Token: token.Token{
			ExpTimeAccess:  con.ExpTimeAccess,
			ExpTimeRefresh: con.ExpTimeRefresh,
		},
	}

	var err error
	useCase.Token.PrivateKey, useCase.Token.PublicKey, err = token.TokenKey(con.PrivateKey, con.PublicKey)
	if err != nil {
		log.Fatal(err)
	}
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

	err = grpcServer.Serve(l)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Start")
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
