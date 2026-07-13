package delivery

import (
	"context"
	"errors"
	userv2 "restaurant/api/proto/user/v2"
	"restaurant/services/user/internal/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService interface {
	CreateUser(ctx context.Context, name, password string) (*model.User, error)
	LoginUser(ctx context.Context, name, password string) (string, error)
}

type GRPCHandler struct {
	userv2.UnimplementedUserServiceServer
	userService UserService
}

func NewHandler(userService UserService) *GRPCHandler {
	return &GRPCHandler{userService: userService}
}

func (g *GRPCHandler) RegisterUser(ctx context.Context, req *userv2.RegisterUserRequest) (*userv2.RegisterUserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	user, err := g.userService.CreateUser(ctx, req.GetName(), req.GetPassword())
	if err != nil {
		if errors.Is(err, model.ErrInvalidName) {
			return nil, status.Error(codes.InvalidArgument, model.ErrInvalidName.Error())
		}

		if errors.Is(err, model.ErrWeakPassword) {
			return nil, status.Error(codes.InvalidArgument, model.ErrWeakPassword.Error())
		}

		if errors.Is(err, model.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, model.ErrUserAlreadyExists.Error())
		}

		return nil, status.Error(codes.Internal, "any internal error")
	}

	return &userv2.RegisterUserResponse{
		UserId: user.ID,
	}, nil
}

func (g *GRPCHandler) LoginUser(ctx context.Context, req *userv2.LoginUserRequest) (*userv2.LoginUserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	userID, err := g.userService.LoginUser(ctx, req.GetName(), req.GetPassword())
	if err != nil {
		if errors.Is(err, model.ErrInvalidName) {
			return nil, status.Error(codes.InvalidArgument, model.ErrInvalidName.Error())
		}

		if errors.Is(err, model.ErrWeakPassword) {
			return nil, status.Error(codes.InvalidArgument, model.ErrWeakPassword.Error())
		}

		if errors.Is(err, model.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, model.ErrUserNotFound.Error())
		}

		if errors.Is(err, model.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, model.ErrInvalidCredentials.Error())
		}

		return nil, status.Error(codes.Internal, "any internal error")
	}

	return &userv2.LoginUserResponse{
		UserId: userID,
	}, nil
}
