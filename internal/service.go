package users

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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

// Service constructor
func NewAppService(s Store) *AppService {
	return &AppService{
		store: s,
	}
}

// CheckUser checks if the provided user exists in the store.
// @param ctx context.Context for managing the scope of the operation.
// @param user User representing the user credentials to check.
// @return bool indicating whether the user exists, string containing user ID (if found), and error (if any).
func (s *AppService) CheckUser(ctx context.Context, user User) (bool, string, error) {
	ID, err := s.store.CheckUser(ctx, user)
	return err == nil, ID, err
}

// NewUser creates a new user in the store.
// @param ctx context.Context for managing the scope of the operation.
// @param user User containing the new user credentials.
// @return string containing the new user ID or an error if user already exists.
func (s *AppService) NewUser(ctx context.Context, user User) (string, error) {
	_, err := s.store.CheckUser(ctx, user)
	if err == oops.ErrNoUser {
		return s.store.SaveUser(ctx, user)
	}

	return "", oops.ErrNoUser
}

// CreateToken creates a new authentication token for the user based on their credentials.
// @param ctx context.Context for managing the scope of the operation.
// @param login string for the user's login name.
// @param password string for the user's password.
// @return Token representing the created token and an error (if any).
func (s *AppService) CreateToken(ctx context.Context, login string, password string) (Token, error) {
	// Check credentials of user, exit if there is no user with such credentials
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

// Bind associates a token with a user ID by saving it in the store.
// @param ctx context.Context for managing the scope of the operation.
// @param token Token to be bound to the user ID.
// @param ID string representing the user ID to which the token will be associated.
// @return error indicating if the operation was successful or if an error occurred.
func (s *AppService) Bind(ctx context.Context, token Token, ID string) error {
	return s.store.SaveToken(ctx, token, ID)
}

// GetUniqueToken generates a unique access and refresh token.
// @param ctx context.Context for managing the scope of the operation.
// @return Token containing generated access and refresh tokens, and an error if the generation fails.
func (s *AppService) GetUniqueToken(ctx context.Context) (Token, error) {
	token := Token{}
	access := make([]byte, TokenLen)
	got := false

	// Generate unique access token
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

	println("err1")

	// error of getting unique access token
	if !got {
		return Token{}, oops.ErrNoTokens
	}

	refresh := make([]byte, TokenLen)
	got = false

	// Generate unique refresh token
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

	println("err2")

	// error of getting unique refresh token
	if !got {
		return Token{}, nil
	}

	println("err3")

	return Token{
		Access:     hex.EncodeToString(access),
		Refresh:    hex.EncodeToString(refresh),
		Expiration: time.Now().Add(time.Minute * 10),
	}, nil
}

// GetIDByToken retrieves the user ID associated with the access token.
// @param ctx context.Context for managing the scope of the operation.
// @param access string representing the user's access token.
// @return string representing the user ID and an error if retrieval fails.
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

// IsExpired checks if the provided access token has expired.
// @param ctx context.Context for managing the scope of the operation.
// @param access string representing the access token to check.
// @return bool indicating whether the token is expired and an error if the check fails.
func (s *AppService) IsExpired(ctx context.Context, access string) (bool, error) {
	return s.store.TokenExpired(ctx, access)
}

// DeleteToken removes the specified access token from the store.
// @param ctx context.Context for managing the scope of the operation.
// @param access string representing the access token to delete.
// @return error indicating if the operation was successful or if an error occurred.
func (s *AppService) DeleteToken(ctx context.Context, access string) error {
	return s.store.PopToken(ctx, access)
}

// UserInfo retrieves user information based on the user ID provided.
// @param ctx context.Context for managing the scope of the operation.
// @param ID string representing the user ID to retrieve information for.
// @return User containing the details of the requested user and an error if retrieval fails.
func (s *AppService) UserInfo(ctx context.Context, ID string) (User, error) {
	return s.store.User(ctx, ID)
}

// DeleteUser removes a user from the store based on their ID.
// @param ctx context.Context for managing the scope of the operation.
// @param ID string representing the user ID to delete.
// @return error indicating if the operation was successful or if an error occurred.
func (s *AppService) DeleteUser(ctx context.Context, ID string) error {
	return s.store.PopUser(ctx, ID)
}

// EditUser modifies an existing user's information.
// @param ctx context.Context for managing the scope of the operation.
// @param token string containing the access token of the user making the request.
// @param user User containing the updated user details.
// @return User with the updated information and an error if the operation fails.
func (s *AppService) EditUser(ctx context.Context, token string, user User) (User, error) {
	// get permission of user with token
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

// GivePermission grants permissions to a specified user.
// @param ctx context.Context for managing the scope of the operation.
// @param token string containing the access token of the admin user.
// @param ID string representing the user ID to which permissions will be given.
// @param Permissions uint representing the permission bits to set.
// @return error indicating if the operation was successful or if an error occurred.
func (s *AppService) GivePermission(ctx context.Context, token string, ID string, Permissions uint) error {
	adminID, err := s.GetIDByToken(ctx, token)
	if err != nil {
		return oops.ErrTokenExistance
	}

	info, errinfo := s.UserInfo(ctx, adminID)
	if errinfo != nil {
		return oops.ErrNoUser
	}

	if (info.Permissions & PermManageUsers) == 0 {
		return oops.ErrWrongPermissions
	}

	return s.store.SetPermission(ctx, ID, Permissions)
}

// RefreshToken handles the token refresh operation by validating the provided access and refresh tokens,
// and generating a new token if valid.
// @param ctx context.Context for managing the scope of the operation.
// @param access string representing the user's existing access token.
// @param refresh string representing the user's refresh token.
// @return Token containing the newly generated tokens and an error, if any occurs during the process.
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
