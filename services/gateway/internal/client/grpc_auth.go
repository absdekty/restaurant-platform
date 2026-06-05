// services/gateway/internal/client/auth_client.go
package client

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authv1 "restaurant/api/proto/auth/v1"
)

type AuthClient struct {
	client authv1.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(5*time.Second))
	if err != nil {
		return nil, err
	}

	return &AuthClient{
		client: authv1.NewAuthServiceClient(conn),
		conn:   conn,
	}, nil
}

func (a *AuthClient) Close() error {
	return a.conn.Close()
}

func (a *AuthClient) ValidateToken(ctx context.Context, token string) (string, error) {
	resp, err := a.client.ValidateToken(ctx, &authv1.ValidateTokenRequest{
		AccessToken: token})
	if err != nil {
		return "", err
	}

	if resp.Error != nil && *resp.Error != "" {
		return "", err
	}

	return resp.UserId, nil
}
