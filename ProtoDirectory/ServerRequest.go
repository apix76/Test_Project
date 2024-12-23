package ProtoDirectory

import (
	"TestProject/usecase"
	"context"
	"errors"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"net"
)

type Server struct {
	UnimplementedTokenServer
	UseCase usecase.UseCase
}

func (s Server) CreateToken(ctx context.Context, guid *Guid) (*Tokens, error) {
	if ctx == nil || guid == nil {
		return nil, errors.New("ctx and guid cannot be nil")
	}
	tokens := Tokens{}

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if ip := md.Get("X-Real-Ip"); ip != nil {
			s.UseCase.UserIp = ip[0]
		}
		if ip := md.Get("X-Forwarded-For"); ip != nil && s.UseCase.UserIp == "" {
			s.UseCase.UserIp = ip[0]
		}
	}

	if s.UseCase.UserIp == "" {
		p, ok := peer.FromContext(ctx)
		if !ok {
			return nil, errors.New("failed to get client info")
		}
		var err error
		s.UseCase.UserIp, _, err = net.SplitHostPort(p.Addr.String())
		if err != nil {
			return nil, err
		}
	}

	var err error
	tokens.AccessToken, tokens.RefreshToken, err = s.UseCase.CreateSession(guid.Guid)
	if err != nil {
		return nil, err
	}
	return &tokens, nil
}

func (s Server) RefreshToken(ctx context.Context, tokens *Tokens) (*Tokens, error) {
	if ctx == nil || tokens == nil {
		return nil, errors.New("ctx and tokens cannot be nil")
	}
	NewTokens := Tokens{}
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("failed to get client info")
	}
	host, _, err := net.SplitHostPort(p.Addr.String())
	if err != nil {
		return nil, err
	}
	s.UseCase.UserIp = host

	NewTokens.AccessToken, NewTokens.RefreshToken, err = s.UseCase.RefreshSession(tokens.AccessToken, tokens.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &NewTokens, nil
}
