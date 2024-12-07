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

// Permission is the flag enum for the permissions a user might have
type Permission uint

const (
	// PermManageBooks allows the user to add, edit and delete books from the library
	PermManageBooks uint = 1 << 0
	// PermQueryTotalStock allows the user to get the total stored book count
	PermQueryTotalStock uint = 1 << 1
	// PermChangeTotalStock allows the user to register updates to the total stored book count.
	// Requires PermGetTotalStock as a prerequisite.
	PermChangeTotalStock uint = 1 << 2
	// PermQueryUsers allows the user to get information about other users, including their permissions.
	// Not required to get information about oneself, other rules apply.
	PermQueryUsers uint = 1 << 3
	// PermManageUsers allows the user to add, edit and delete other users.
	// Not required to manage oneself, other rules apply.
	// Requires PermQueryUsers as a prerequisite.
	PermManageUsers uint = 1 << 4
	// PermGrantPermissions allows the user to grant permissions to other users.
	// Only a subset of own permissions may be granted.
	// Requires PermQueryUsers as a prerequisite.
	PermGrantPermissions uint = 1 << 5
	// PermLoanBooks allows the user to register book takeouts and returns.
	PermLoanBooks uint = 1 << 6
	// PermQueryAvailableStock allows the user to get the number of available (not lent out) copies of a book.
	PermQueryAvailableStock uint = 1 << 7
	// PermQueryReservations allows the user to get information related to book reservations.
	PermQueryReservations uint = 1 << 8
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

	// generate access
	for i := 0; i < GenerateRetries; i++ {
		_, err := rand.Read(access)
		token.Access = string(access)
		if err == nil || err == oops.ErrDupAccess {
			_, err := s.store.CheckToken(ctx, token.Access)
			if err == nil {
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

	// generate refresh
	for i := 0; i < GenerateRetries; i++ {
		_, err := rand.Read(refresh)
		token.Refresh = string(refresh)
		if err == nil {
			_, err := s.store.CheckToken(ctx, token.Refresh)
			if err == nil {
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

	_, err := s.store.CheckToken(ctx, access)

	if err == nil {
		return "", oops.ErrTokenExistance
	}

	flag, err := s.IsExpired(ctx, access)
	if err != nil {
		return "", oops.ErrTokenExpired
	}
	if flag {
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

func (s *AppService) DeleteUser(ctx context.Context, ID string) error {
	return s.store.PopUser(ctx, ID)
}

func (s *AppService) EditUser(ctx context.Context, token string, user User) (User, error) {
	ID, err := s.GetIDByToken(ctx, token)
	if err != nil {
		return User{}, oops.ErrTokenExistance
	}

	info, errinfo := s.UserInfo(ctx, ID)
	if errinfo != nil {
		return User{}, oops.ErrNoUser
	}

	if (info.Permissions & PermManageUsers) == 0 {
		return User{}, oops.ErrWrongPermissions
	}

	return s.store.ChangeUser(ctx, user)
}

func (s *AppService) GivePermission(ctx context.Context, token string, ID string, Permissions uint) error {
	adminID, err := s.GetIDByToken(ctx, token)
	if err != nil {
		return oops.ErrTokenExistance
	}
	println("user got")

	info, errinfo := s.UserInfo(ctx, adminID)
	if errinfo != nil {
		return oops.ErrNoUser
	}

	if (info.Permissions & PermManageUsers) == 0 {
		return oops.ErrWrongPermissions
	}

	return s.store.SetPermission(ctx, ID, Permissions)
}

func (s *AppService) RefreshToken(ctx context.Context, access string, refresh string) (Token, error) {
	token, err := s.store.CheckToken(ctx, access)
	if err == nil {
		return Token{}, err
	}

	if token.Refresh != refresh {
		return Token{}, oops.ErrNoRefresh
	}

	ID, UserError := s.GetIDByToken(ctx, access)

	if UserError != nil {
		return Token{}, oops.ErrNoUser
	}

	s.store.PopToken(ctx, access)

	newToken, newErr := s.GetUniqueToken(ctx)

	if newErr != nil {
		return Token{}, oops.ErrNoTokens
	}

	s.store.SaveToken(ctx, newToken, ID)

	return newToken, nil
}
