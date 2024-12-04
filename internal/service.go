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

func (s *AppService) CheckUser(ctx context.Context, user User) (bool, string, error) {
	ID, err := s.store.CheckUser(ctx, user)
	return err == nil, ID, err
}

func (s *AppService) NewUser(ctx context.Context, user User) (string, error) {
	_, err := s.store.CheckUser(ctx, user)
	if err == oops.ErrNoUser {
		return s.store.SaveUser(ctx, user)
	}

	return "", err
}

func (s *AppService) CreateToken(ctx context.Context, login string, password string) (Token, error) {
	// Check credentials
	checked, ID, err := s.CheckUser(ctx, User{"", login, password, 0})
	if err != nil || !checked {
		return Token{}, oops.ErrNoUser
	}

	token, err := s.GetUniqueToken(ctx)

	if err != nil {
		return Token{}, err
	}

	return token, s.Bind(ctx, token, ID)
}

func (s *AppService) Bind(ctx context.Context, token Token, ID string) error {
	return s.store.SaveToken(ctx, token, ID)
}

func (s *AppService) GetUniqueToken(ctx context.Context) (Token, error) {
	token := Token{}
	access := make([]byte, TokenLen)
	got := false

	// generate acc
	for i := 0; i < GenerateRetries; i++ {
		_, err := rand.Read(access)
		token.Access = string(access)
		if err == nil || err == oops.ErrDupAccess {
			flag, _ := s.store.CheckToken(ctx, token.Access)
			if !flag {
				got = true
				break
			}
		}
	}

	if !got {
		return Token{}, oops.ErrNoTokens
	}

	refresh := make([]byte, TokenLen)
	got = false

	for i := 0; i < GenerateRetries; i++ {
		_, err := rand.Read(refresh)
		token.Refresh = string(refresh)
		if err == nil {
			flag, _ := s.store.CheckToken(ctx, token.Refresh)
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

func (s *AppService) GetIDByToken(ctx context.Context, access string) (string, error) {

	flag, err := s.store.CheckToken(ctx, access)

	if !flag || err != nil {
		return "", oops.ErrTokenExistance
	}

	flag, err = s.IsExpired(ctx, access)
	if err != nil {
		return "", oops.ErrTokenExpired
	}
	if flag {
		s.DeleteToken(ctx, access)
		return "", oops.ErrTokenExpired
	}

	return s.store.GetSessionID(ctx, access)
}

func (s *AppService) IsExpired(ctx context.Context, access string) (bool, error) {
	return s.store.TokenExpired(ctx, access)
}

func (s *AppService) DeleteToken(ctx context.Context, access string) error {
	return s.store.PopToken(ctx, access)
}

func (s *AppService) Place(ctx context.Context, product User) (id string, err error) {
	id, err = s.store.SaveUser(ctx, product)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *AppService) UserInfo(ctx context.Context, ID string) (User, error) {
	return s.store.User(ctx, ID)
}
