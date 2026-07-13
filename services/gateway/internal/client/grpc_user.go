package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"restaurant/pkg/interceptors"
	"restaurant/pkg/models"
	"restaurant/services/gateway/internal/model"

	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"

	userv2 "restaurant/api/proto/user/v2"
)

type UserClient struct {
	client userv2.UserServiceClient
	conn   *grpc.ClientConn
	cb     CircuitBreaker
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

func NewUserClient(creds credentials.TransportCredentials, addr string, config UserConfig, cb CircuitBreaker) (*UserClient, error) {
	serviceConfig := fmt.Sprintf(`{
		"methodConfig": [{
			"name": [{"service": "user.v2.UserService"}],
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
		client: userv2.NewUserServiceClient(conn),
		conn:   conn,
		cb:     cb,
	}, nil
}

func (a *UserClient) Close() error {
	return a.conn.Close()
}

func (a *UserClient) RegisterUser(ctx context.Context, name, password string) (string, error) {
	result, err := a.cb.Execute(func() (interface{}, error) {
		return a.client.RegisterUser(ctx, &userv2.RegisterUserRequest{
			Name:     name,
			Password: password,
		})
	})

	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			return "", models.ErrServiceUnavailable
		}

		if status.Code(err) == codes.Unavailable {
			return "", models.ErrServiceUnavailable
		}

		if status.Code(err) == codes.InvalidArgument {
			return "", model.ErrUserInvalidRegisterDetails
		}

		if status.Code(err) == codes.AlreadyExists {
			return "", model.ErrUserAlreadyExists
		}

		return "", err
	}

	resp := result.(*userv2.RegisterUserResponse)
	return resp.UserId, nil
}

func (a *UserClient) LoginUser(ctx context.Context, name, password string) (string, error) {
	result, err := a.cb.Execute(func() (interface{}, error) {
		return a.client.LoginUser(ctx, &userv2.LoginUserRequest{
			Name:     name,
			Password: password,
		})
	})

	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			return "", models.ErrServiceUnavailable
		}

		if status.Code(err) == codes.Unavailable {
			return "", models.ErrServiceUnavailable
		}

		if status.Code(err) == codes.InvalidArgument {
			return "", model.ErrUserInvalidRegisterDetails
		}

		if status.Code(err) == codes.Unauthenticated {
			return "", model.ErrUserInvalidCredentials
		}

		if status.Code(err) == codes.NotFound {
			return "", model.ErrUserNotFound
		}

		return "", err
	}

	resp := result.(*userv2.LoginUserResponse)
	return resp.UserId, nil
}
