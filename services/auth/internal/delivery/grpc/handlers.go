package delivery

import (
	"context"
	"errors"
	authv1 "restaurant/api/proto/auth/v1"
	"restaurant/services/auth/internal/model"
)

type HandlerToken interface {
	ValidateAccessToken(tokenStr string) (string, error)
}

type GRPCHandler struct {
	authv1.UnimplementedAuthServiceServer
	tokenService HandlerToken
}

func NewHandler(tokenService HandlerToken) *GRPCHandler {
	return &GRPCHandler{
		tokenService: tokenService,
	}
}

func (g *GRPCHandler) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	userID, err := g.tokenService.ValidateAccessToken(req.AccessToken)
	if err != nil {
		if errors.Is(err, model.ErrTokenNotFound) {
			return &authv1.ValidateTokenResponse{
				Error: ptrString("Token not found"),
			}, nil
		}

		if errors.Is(err, model.ErrTokenRevoked) {
			return &authv1.ValidateTokenResponse{
				Error: ptrString("Token revoked"),
			}, nil
		}

		return &authv1.ValidateTokenResponse{
			Error: ptrString(err.Error()),
		}, nil
	}

	return &authv1.ValidateTokenResponse{
		UserId: userID,
		Error:  nil,
	}, nil
}

func ptrString(s string) *string {
	return &s
}
