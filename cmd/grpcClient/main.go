package main

import (
	"TestProject/ProtoDirectory"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"log"
	"os"
)

type Conf struct {
	Port string
}

type Execute struct {
	client ProtoDirectory.TokenClient
}

func main() {
	conf := NewConf()
	con, err := grpc.NewClient(fmt.Sprintf("127.0.0.1%s", conf.Port), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer con.Close()

	e := Execute{client: ProtoDirectory.NewTokenClient(con)}
	wg := errgroup.Group{}
	for i := 0; i < 10; i++ {
		wg.Go(e.Testing)
	}
	if err := wg.Wait(); err != nil {
		fmt.Println(err)
	}
}

func (e *Execute) Testing() error {
	for i := 0; i < 10; i++ {
		token, err := e.Get()
		if err != nil {
			return err
		}
		err = e.Refresh(token)
		if err != nil {
			return err
		}
	}
	return nil
}
func (e *Execute) Get() (*ProtoDirectory.Tokens, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		log.Fatal(err)
	}

	guid := ProtoDirectory.Guid{Guid: id.String()}
	tokens, err := e.client.CreateToken(context.Background(), &guid)
	if err != nil {
		return nil, err
	}
	return tokens, err
}

func (e *Execute) Refresh(Tokens *ProtoDirectory.Tokens) error {
	_, err := e.client.RefreshToken(context.Background(), Tokens)
	return err
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
