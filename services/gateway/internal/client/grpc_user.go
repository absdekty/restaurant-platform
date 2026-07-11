package client

import (
	"context"
	"fmt"
	"time"

	"restaurant/pkg/interceptors"
	"restaurant/services/gateway/internal/model"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"

	userv1 "restaurant/api/proto/user/v1"
)

type UserClient struct {
	client userv1.UserServiceClient
	conn   *grpc.ClientConn
}

type UserConfig struct {
	RetryMaxAttempts       int
	RetryInitialBackoff    string
	RetryMaxBackoff        string
	RetryBackoffMultiplier float64

	KeepaliveTime          time.Duration
	KeepaliveTimeout       time.Duration
	KeepalivePermitWithout bool
}

func NewUserClient(creds credentials.TransportCredentials, addr string, config UserConfig) (*UserClient, error) {
	serviceConfig := fmt.Sprintf(`{
		"methodConfig": [{
			"name": [{"service": "auth.v3.AuthService"}],
			"retryPolicy": {
				"maxAttempts": %d,
				"initialBackoff": "%s",
				"maxBackoff": "%s",
				"backoffMultiplier": %.1f,
				"retryableStatusCodes": ["UNAVAILABLE"]
			}
		}]
	}`,
		config.RetryMaxAttempts,
		config.RetryInitialBackoff,
		config.RetryMaxBackoff,
		config.RetryBackoffMultiplier,
	)

	keepaliveParams := keepalive.ClientParameters{
		Time:                config.KeepaliveTime,
		Timeout:             config.KeepaliveTimeout,
		PermitWithoutStream: config.KeepalivePermitWithout,
	}

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultServiceConfig(serviceConfig),
		grpc.WithKeepaliveParams(keepaliveParams),
		grpc.WithUnaryInterceptor(interceptors.TraceClient()))
	if err != nil {
		return nil, err
	}

	return &UserClient{
		client: userv1.NewUserServiceClient(conn),
		conn:   conn,
	}, nil
}

func (a *UserClient) Close() error {
	return a.conn.Close()
}

func (a *UserClient) RegisterUser(ctx context.Context, name, password string) (string, error) {
	resp, err := a.client.RegisterUser(ctx, &userv1.RegisterUserRequest{
		Name:     name,
		Password: password,
	})
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			return "", model.ErrUserInvalidRegisterDetails
		}

		if status.Code(err) == codes.AlreadyExists {
			return "", model.ErrUserAlreadyExists
		}

		return "", err
	}

	return resp.UserId, nil
}

func (a *UserClient) LoginUser(ctx context.Context, name, password string) (string, string, int32, error) {
	resp, err := a.client.LoginUser(ctx, &userv1.LoginUserRequest{
		Name:     name,
		Password: password,
	})
	if err != nil {
		if status.Code(err) == codes.InvalidArgument {
			return "", "", 0, model.ErrUserInvalidCredentials
		}

		if status.Code(err) == codes.NotFound {
			return "", "", 0, model.ErrUserNotFound
		}

		return "", "", 0, err
	}

	return resp.AccessToken, resp.RefreshToken, resp.RefreshTokenTtl, nil
}
