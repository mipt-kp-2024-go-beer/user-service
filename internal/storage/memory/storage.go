package memory

import (
	"context"
	"sync"
	"time"

	users "github.com/mipt-kp-2024-go-beer/user-service/internal"
	"github.com/mipt-kp-2024-go-beer/user-service/internal/oops"
)

type UserValues struct {
	ID          string
	Password    string
	Permissions uint
}

type UserDb struct {
	mux sync.RWMutex
	// user login is used as a key
	Users map[string]UserValues
}

type Token struct {
	// refresh token is used as a key
	access     string
	expiration time.Time
}

type TokenDb struct {
	mux    sync.RWMutex
	Tokens map[string]Token
}

type Storage struct {
	Users  UserDb
	Tokens TokenDb
}

var curID = 0

func NewStorage() *Storage {
	return &Storage{UserDb{Users: make(map[string]UserValues)}, TokenDb{Tokens: make(map[string]Token)}}
}

func (s *Storage) LoadUsers(ctx context.Context) ([]users.User, error) {
	s.Users.mux.RLock()
	defer s.Users.mux.RUnlock()
	output := make([]users.User, len(s.Users.Users))

	idx := 0
	for i, v := range s.Users.Users {
		output[idx].Login = i
		output[idx].ID = v.ID
		output[idx].Password = v.Password
		output[idx].Permissions = v.Permissions
		idx++
	}

	return output, nil
}

func (s *Storage) CheckUser(ctx context.Context, user users.User) (id string, err error) {
	s.Users.mux.Lock()
	defer s.Users.mux.Unlock()
	val, ok := s.Users.Users[user.Login]
	if !ok {
		return val.ID, oops.ErrNoUser
	}

	return val.ID, nil
}

func (s *Storage) CheckToken(ctx context.Context, token users.Token) (bool, error) {
	s.Tokens.mux.RLock()
	defer s.Tokens.mux.RUnlock()
	val, ok := s.Tokens.Tokens[token.Refresh]

	if !ok {
		return false, oops.ErrDupRefresh
	}

	if val.access != token.Access {
		return false, oops.ErrDupAccess
	}
	return true, nil
}

func (s *Storage) SaveUser(ctx context.Context, user users.User) (id string, err error) {
	s.Users.mux.Lock()
	defer s.Users.mux.Unlock()
	_, ok := s.Users.Users[user.Login]
	if ok {
		return user.ID, oops.ErrDuplicateUser
	}

	s.Users.Users[user.Login] = UserValues{ID: user.Login, Password: user.Password, Permissions: user.Permissions}
	curID++
	return string(curID), nil
}

func (s *Storage) LoadTokens(ctx context.Context) ([]users.Token, error) {
	s.Tokens.mux.RLock()
	defer s.Tokens.mux.RUnlock()
	output := make([]users.Token, len(s.Users.Users))

	idx := 0
	for i, v := range s.Tokens.Tokens {
		output[idx].Refresh = i
		output[idx].Access = v.access
		output[idx].Expiration = v.expiration
		idx++
	}

	return output, nil
}
