package users

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type AppService struct {
	store Store
}

func NewAppService(s Store) *AppService {
	return &AppService{
		store: s,
	}
}

// generating random session token
func (s *AppService) GenerateSecureToken(ctx context.Context, length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func (s *AppService) Products(ctx context.Context) ([]User, error) {
	products, err := s.store.LoadUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("firdge.Products() error: %w", err)
	}

	return products, nil
}

func (s *AppService) Place(ctx context.Context, product User) (id string, err error) {
	id, err = s.store.SaveUser(ctx, product)
	if err != nil {
		return "", err
	}

	return id, nil
}
