package memory

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	users "github.com/mipt-kp-2024-go-beer/user-service/internal"
	"github.com/mipt-kp-2024-go-beer/user-service/internal/oops"
)

// UserValues holds the information for a user, including their ID, password, and permissions.
type UserValues struct {
	ID          string
	Password    string
	Permissions uint
}

// UserDb is a thread-safe structure that stores user information indexed by their login.
type UserDb struct {
	mux sync.RWMutex
	// user login is used as a key
	// as we need to generate ID for every person
	Users map[string]UserValues
}

// Token represents an authentication token associated with a user.
type Token struct {
	// acess token is used as a key
	refresh    string
	expiration time.Time
	user       string
}

// TokenDb is a thread-safe structure that stores tokens indexed by their access tokens.
type TokenDb struct {
	mux    sync.RWMutex
	Tokens map[string]Token
}

// Storage combines UserDb and TokenDb to provide a unified storage solution for users and tokens.
type Storage struct {
	Users  UserDb
	Tokens TokenDb
}

// curID is a global variable for generating unique IDs.
var curID = 0

// Storage constructor
func NewStorage() *Storage {
	return &Storage{UserDb{Users: make(map[string]UserValues)}, TokenDb{Tokens: make(map[string]Token)}}
}

// Load users table to user by user with corresponding permissions
// @param ctx context.Context for managing the scope of the operation.
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

// Check if user is present in current project
// @param ctx context.Context for managing the scope of the operation.
// @param user users.User user to be checked
func (s *Storage) CheckUser(ctx context.Context, user users.User) (id string, err error) {
	s.Users.mux.Lock()
	defer s.Users.mux.Unlock()
	val, ok := s.Users.Users[user.Login]
	if !ok {
		return val.ID, oops.ErrNoUser
	}

	if val.Password != user.Password {
		return val.ID, oops.ErrNoUser
	}

	return val.ID, nil
}

// Check if token is present in current project
// @param ctx context.Context for managing the scope of the operation.
// @param access users.User user to be checked
func (s *Storage) CheckToken(ctx context.Context, access string) (users.Token, error) {
	s.Tokens.mux.RLock()
	defer s.Tokens.mux.RUnlock()
	val, ok := s.Tokens.Tokens[access]

	if ok {
		return users.Token{Access: access, Refresh: val.refresh, Expiration: val.expiration}, oops.ErrDupAccess
	}

	delete(s.Tokens.Tokens, access)

	return users.Token{}, nil
}

// Save user if it is not in the d
// @param ctx context.Context for managing the scope of the operation.
// @param user users.User user to be added
func (s *Storage) SaveUser(ctx context.Context, user users.User) (id string, err error) {
	s.Users.mux.Lock()
	defer s.Users.mux.Unlock()
	fmt.Printf("we got here")
	_, ok := s.Users.Users[user.Login]
	if ok {
		return user.ID, oops.ErrDuplicateUser
	}

	curID++
	ID := strconv.Itoa(curID)
	s.Users.Users[user.Login] = UserValues{ID: ID, Password: user.Password, Permissions: user.Permissions}
	return ID, nil
}

// Save user if it is not in the d
// @param ctx context.Context for managing the scope of the operation.
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

// Save user if it is not in the d
// @param ctx context.Context for managing the scope of the operation.
// @param access string access token
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

// Save token if it is not in the storage
// @param ctx context.Context for managing the scope of the operation.
// @param token users.Token token to be saved
// @param ID string related user ID
func (s *Storage) SaveToken(ctx context.Context, token users.Token, ID string) (err error) {

	s.Tokens.Tokens[token.Access] = Token{refresh: token.Refresh, expiration: token.Expiration, user: ID}
	return nil
}

// check if token is expired
// @param ctx context.Context for managing the scope of the operation.
// @param token users.Token token to be saved
// @param ID string related user ID
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

// delete token
// @param ctx context.Context for managing the scope of the operation.
// @param acess string token to be deleted
func (s *Storage) PopToken(ctx context.Context, access string) error {
	delete(s.Users.Users, access)
	return nil
}

// get User from storage
// @param ctx context.Context for managing the scope of the operation.
// @param ID string user ID
func (s *Storage) User(ctx context.Context, ID string) (users.User, error) {
	for i, v := range s.Users.Users {
		if v.ID == ID {
			return users.User{ID: v.ID, Login: i, Password: v.Password, Permissions: v.Permissions}, nil
		}
	}

	return users.User{}, oops.ErrNoUser
}

// delete User from storage
// @param ctx context.Context for managing the scope of the operation.
// @param ID string user ID
func (s *Storage) PopUser(ctx context.Context, ID string) error {
	s.Users.mux.Lock()
	defer s.Users.mux.Unlock()
	for i, v := range s.Users.Users {
		if v.ID == ID {
			delete(s.Users.Users, i)
			return nil
		}
	}

	return oops.ErrNoUser
}

// delete User from storage
// @param ctx context.Context for managing the scope of the operation.
// @param user users.User user to be changed
func (s *Storage) ChangeUser(ctx context.Context, user users.User) (users.User, error) {
	s.Users.mux.Lock()
	defer s.Users.mux.Unlock()
	for i, v := range s.Users.Users {
		if v.ID == user.ID {
			Permissions := v.Permissions
			delete(s.Users.Users, i)
			s.Users.Users[user.Login] = UserValues{ID: user.ID, Password: user.Password, Permissions: Permissions}
			return user, nil
		}
	}

	return users.User{}, oops.ErrNoUser
}

// delete User from storage
// @param ctx context.Context for managing the scope of the operation.
// @param ID string user ID
// @param Permission uint user Permission
func (s *Storage) SetPermission(ctx context.Context, ID string, Permissions uint) error {
	for i, v := range s.Users.Users {
		if v.ID == ID {
			s.Users.Users[i] = UserValues{ID: v.ID, Password: v.Password, Permissions: Permissions}
			return nil
		}
	}

	return oops.ErrNoUser
}
