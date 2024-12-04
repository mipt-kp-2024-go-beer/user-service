package memory

import (
	"context"
	"strconv"
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
	// as we need to generate ID for every person
	Users map[string]UserValues
}

type Token struct {
	// acess token is used as a key
	refresh    string
	expiration time.Time
	user       string
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

func (s *Storage) CheckToken(ctx context.Context, access string) (bool, error) {
	s.Tokens.mux.RLock()
	defer s.Tokens.mux.RUnlock()
	_, ok := s.Tokens.Tokens[access]

	if !ok {
		return false, nil
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

	curID++
	ID := strconv.Itoa(curID)
	s.Users.Users[user.Login] = UserValues{ID: ID, Password: user.Password, Permissions: user.Permissions}
	return ID, nil
}

func (s *Storage) LoadTokens(ctx context.Context) ([]users.Token, error) {
	s.Tokens.mux.RLock()
	defer s.Tokens.mux.RUnlock()
	output := make([]users.Token, len(s.Users.Users))

	idx := 0
	for i, v := range s.Tokens.Tokens {
		output[idx].Refresh = v.refresh
		output[idx].Access = i
		output[idx].Expiration = v.expiration
		idx++
	}

	return output, nil
}

func (s *Storage) GetSessionID(ctx context.Context, access string) (string, error) {
	s.Tokens.mux.RLock()
	defer s.Tokens.mux.RUnlock()

	s.Users.mux.RLock()
	defer s.Users.mux.RUnlock()

	val, ok := s.Tokens.Tokens[access]
	if !ok {
		return "", oops.ErrTokenExistance
	}

	return val.user, nil
}

func (s *Storage) SaveToken(ctx context.Context, token users.Token, ID string) (err error) {

	s.Tokens.Tokens[token.Access] = Token{refresh: token.Refresh, expiration: token.Expiration, user: ID}
	return nil
}

func (s *Storage) TokenExpired(ctx context.Context, access string) (bool, error) {
	val, ok := s.Tokens.Tokens[access]
	if !ok {
		return false, oops.ErrTokenExistance
	}

	if time.Now().After(val.expiration) {
		return true, nil
	}

	return false, nil
}

func (s *Storage) PopToken(ctx context.Context, access string) error {
	delete(s.Users.Users, access)
	return nil
}
