package delivery

import (
	"context"
	"errors"
	userv1 "restaurant/api/proto/user/v1"
	"restaurant/services/user/internal/model"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService interface {
	CreateUser(ctx context.Context, name, password string) (*model.User, error)
	LoginUser(ctx context.Context, name, password string) (string, error)
}

type AuthService interface {
	GenerateTokens(ctx context.Context, userID string) (string, string, int32, error)
}

type GRPCHandler struct {
	userv1.UnimplementedUserServiceServer
	userService UserService
	authService AuthService
}

func NewHandler(userService UserService, authService AuthService) *GRPCHandler {
	return &GRPCHandler{userService: userService, authService: authService}
}

func (g *GRPCHandler) RegisterUser(ctx context.Context, req *userv1.RegisterUserRequest) (*userv1.RegisterUserResponse, error) {
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

	return &userv1.RegisterUserResponse{
		UserId: user.ID,
	}, nil
}

func (g *GRPCHandler) LoginUser(ctx context.Context, req *userv1.LoginUserRequest) (*userv1.LoginUserResponse, error) {
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

	accessToken, refreshToken, refreshTokenTtl, err := g.authService.GenerateTokens(ctx, userID)
	// FixMe: проверять ошибку :(

	return &userv1.LoginUserResponse{
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		RefreshTokenTtl: refreshTokenTtl,
	}, nil
}
