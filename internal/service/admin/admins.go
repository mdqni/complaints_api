package service

import (
	"complaint_server/internal/domain"
	"complaint_server/internal/storage/pg"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	adminStorage *pg.Storage
}

func NewAdminService(adminStorage *pg.Storage) *AdminService {
	return &AdminService{adminStorage: adminStorage}
}

func (s *AdminService) Login(ctx context.Context, username, password string) (*domain.Admin, error) {
	admin, err := s.adminStorage.GetAdminByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	return admin, nil
}
