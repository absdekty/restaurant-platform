package delivery

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	authv2 "restaurant/api/proto/auth/v2"
	"restaurant/services/auth/internal/model"
)

type HandlerToken interface {
	ValidateAccessToken(tokenStr string) (string, error)
}

type GRPCHandler struct {
	authv2.UnimplementedAuthServiceServer
	tokenService HandlerToken
}

func NewHandler(tokenService HandlerToken) *GRPCHandler {
	return &GRPCHandler{
		tokenService: tokenService,
	}
}

func (g *GRPCHandler) ValidateToken(ctx context.Context, req *authv2.ValidateTokenRequest) (*authv2.ValidateTokenResponse, error) {
	userID, err := g.tokenService.ValidateAccessToken(req.AccessToken)
	if err != nil {
		if errors.Is(err, model.ErrTokenNotFound) {
			return nil, status.Error(codes.NotFound, "token not found")
		}
		if errors.Is(err, model.ErrTokenRevoked) {
			return nil, status.Error(codes.PermissionDenied, "token revoked")
		}
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &authv2.ValidateTokenResponse{
		UserId: userID,
	}, nil
}

func ptrString(s string) *string {
	return &s
}
