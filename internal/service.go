package users

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/mipt-kp-2024-go-beer/user-service/internal/oops"
)

const TokenLen int = 64
const GenerateRetries int = 5
const ExpirationDuartion int = 10

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

func (s *AppService) CheckUser(ctx context.Context, user User) (bool, error) {
	_, err := s.store.CheckUser(ctx, user)
	return err == nil, err
}

func (s *AppService) NewUser(ctx context.Context, user User) (string, error) {
	_, err := s.store.CheckUser(ctx, user)
	if err == oops.ErrNoUser {
		return s.store.SaveUser(ctx, user)
	}

	return "", err
}

func (s *AppService) GetUniqueToken(ctx context.Context) (Token, error) {
	token := Token{}
	refresh := make([]byte, TokenLen)
	got := false

	for i := 0; i < GenerateRetries; i++ {
		_, err := rand.Read(refresh)
		token.Refresh = string(refresh)
		if err == nil || err == oops.ErrDupAccess {
			flag, _ := s.store.CheckToken(ctx, token)
			if !flag {
				got = true
				break
			}
		}
	}

	if !got {
		return Token{}, oops.ErrNoTokens
	}

	access := make([]byte, TokenLen)
	got = false

	for i := 0; i < GenerateRetries; i++ {
		_, err := rand.Read(access)
		token.Access = string(access)
		if err == nil {
			flag, _ := s.store.CheckToken(ctx, token)
			if !flag {
				got = true
				break
			}
		}
	}

	if !got {
		return Token{}, nil
	}

	return Token{Access: hex.EncodeToString(access), Refresh: hex.EncodeToString(refresh), Expiration: time.Now().Add(time.Minute * 10)}, nil
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
