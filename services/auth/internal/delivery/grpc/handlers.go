package delivery

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	authv3 "restaurant/api/proto/auth/v3"
	"restaurant/services/auth/internal/model"
)

type HandlerToken interface {
	ValidateAccessToken(tokenStr string) (string, error)
	GenerateTokens(ctx context.Context, userID string) (string, string, int32, error)
	RefreshTokens(ctx context.Context, token string) (string, string, int32, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}

type GRPCHandler struct {
	authv3.UnimplementedAuthServiceServer
	tokenService HandlerToken
}

func NewHandler(tokenService HandlerToken) *GRPCHandler {
	return &GRPCHandler{
		tokenService: tokenService,
	}
}

func (g *GRPCHandler) ValidateToken(ctx context.Context, req *authv3.ValidateTokenRequest) (*authv3.ValidateTokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	userID, err := g.tokenService.ValidateAccessToken(req.GetAccessToken())
	if err != nil {
		if errors.Is(err, model.ErrTokenNotFound) {
			return nil, status.Error(codes.Unauthenticated, "token not found")
		}
		if errors.Is(err, model.ErrTokenRevoked) {
			return nil, status.Error(codes.PermissionDenied, "token revoked")
		}
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &authv3.ValidateTokenResponse{
		UserId: userID,
	}, nil
}

func (g *GRPCHandler) GenerateTokens(ctx context.Context, req *authv3.GenerateTokensRequest) (*authv3.GenerateTokensResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	accessToken, refreshToken, refreshTTL, err := g.tokenService.GenerateTokens(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "any internal error")
	}

	return &authv3.GenerateTokensResponse{
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		RefreshTokenTtl: refreshTTL,
	}, nil
}

func (g *GRPCHandler) RefreshTokens(ctx context.Context, req *authv3.RefreshTokensRequest) (*authv3.RefreshTokensResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	accessToken, refreshToken, refreshTTL, err := g.tokenService.RefreshTokens(ctx, req.GetRefreshToken())
	if err != nil {
		if errors.Is(err, model.ErrTokenRevoked) {
			return nil, status.Error(codes.PermissionDenied, "token revoked")
		}

		if errors.Is(err, model.ErrTokenNotFound) {
			return nil, status.Error(codes.Unauthenticated, "token not found")
		}

		return nil, status.Error(codes.Internal, "any internal error")
	}

	return &authv3.RefreshTokensResponse{
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		RefreshTokenTtl: refreshTTL,
	}, nil
}

func (g *GRPCHandler) RevokeRefreshToken(ctx context.Context, req *authv3.RevokeRefreshTokenRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	if err := g.tokenService.RevokeRefreshToken(ctx, req.GetRefreshToken()); err != nil {
		if errors.Is(err, model.ErrTokenNotFound) {
			return nil, status.Error(codes.Unauthenticated, "token not found")
		}

		return nil, status.Error(codes.Internal, "any internal error")
	}

	return &emptypb.Empty{}, nil
}
