package delivery

import (
	"context"
)

type UserService interface {
	RegisterUser(ctx context.Context, name, password string) (string, error)
	LoginUser(ctx context.Context, name, password string) (string, string, int32, error)
}
