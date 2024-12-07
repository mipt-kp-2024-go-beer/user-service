package users

import (
	"context"
	"time"
)

type User struct {
	ID          string
	Login       string
	Password    string
	Permissions uint
}

type Token struct {
	Access     string
	Refresh    string
	Expiration time.Time
}

type Service interface {
	//Users(ctx context.Context) ([]User, error)
	GetUniqueToken(ctx context.Context) (Token, error)
	CheckUser(ctx context.Context, user User) (bool, string, error)
	IsExpired(ctx context.Context, access string) (bool, error)
	DeleteToken(ctx context.Context, access string) error
	DeleteUser(ctx context.Context, ID string) error
	NewUser(ctx context.Context, user User) (string, error)
	GetIDByToken(ctx context.Context, access string) (string, error)
	CreateToken(ctx context.Context, login string, password string) (Token, error)
	Bind(ctx context.Context, token Token, ID string) error
	UserInfo(ctx context.Context, ID string) (User, error)
	EditUser(ctx context.Context, token string, user User) (User, error)
	GivePermission(ctx context.Context, token string, ID string, Permissions uint) error
	//Tokens(ctx context.Context) ([]Token, error)
	//Place(ctx context.Context, user User) (id string, err error)
}

type Store interface {
	LoadUsers(ctx context.Context) ([]User, error)
	CheckUser(ctx context.Context, user User) (string, error)
	SaveUser(ctx context.Context, user User) (string, error)
	User(ctx context.Context, ID string) (User, error)
	PopUser(ctx context.Context, ID string) error
	ChangeUser(ctx context.Context, user User) (User, error)
	SetPermission(ctx context.Context, ID string, Permissions uint) error

	SaveToken(ctx context.Context, token Token, ID string) (err error)
	TokenExpired(ctx context.Context, access string) (bool, error)
	PopToken(ctx context.Context, access string) error
	CheckToken(ctx context.Context, access string) (bool, error)
	LoadTokens(ctx context.Context) ([]Token, error)

	GetSessionID(ctx context.Context, access string) (string, error)
}
