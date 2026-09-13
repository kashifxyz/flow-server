package services

import (
	"context"

	"github.com/kashifxyz/flow-server/internal/auth"
	"github.com/kashifxyz/flow-server/internal/modules/users/models"
)

type Service struct {
	Auth *auth.Service
}

func New(a *auth.Service) *Service {
	return &Service{Auth: a}
}

func (s *Service) Register(ctx context.Context, in models.RegisterRequest, meta auth.RequestMeta) error {
	return s.Auth.Register(ctx, auth.RegisterInput{
		Email:       in.Email,
		Password:    in.Password,
		DisplayName: in.DisplayName,
	}, meta)
}

func (s *Service) Login(ctx context.Context, in models.LoginRequest, meta auth.RequestMeta) (auth.IssuedSession, error) {
	return s.Auth.Login(ctx, auth.LoginInput{
		Email:    in.Email,
		Password: in.Password,
	}, meta)
}

func (s *Service) Logout(ctx context.Context, sess auth.Session, meta auth.RequestMeta) error {
	return s.Auth.Logout(ctx, sess, meta)
}

func (s *Service) VerifyEmail(ctx context.Context, token string, meta auth.RequestMeta) error {
	return s.Auth.VerifyEmail(ctx, token, meta)
}

func (s *Service) ResendVerification(ctx context.Context, email string, meta auth.RequestMeta) error {
	return s.Auth.ResendVerification(ctx, email, meta)
}

func (s *Service) ForgotPassword(ctx context.Context, email string, meta auth.RequestMeta) error {
	return s.Auth.ForgotPassword(ctx, email, meta)
}

func (s *Service) ResetPassword(ctx context.Context, token, password string, meta auth.RequestMeta) error {
	return s.Auth.ResetPassword(ctx, token, password, meta)
}
