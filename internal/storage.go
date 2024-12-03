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
	CheckUser(ctx context.Context, user User) (bool, error)
	NewUser(ctx context.Context, user User) (string, error)
	//Tokens(ctx context.Context) ([]Token, error)
	//Place(ctx context.Context, user User) (id string, err error)
}

type Store interface {
	LoadUsers(ctx context.Context) ([]User, error)
	CheckUser(ctx context.Context, user User) (string, error)
	SaveUser(ctx context.Context, user User) (id string, err error)
	CheckToken(ctx context.Context, token Token) (bool, error)
	LoadTokens(ctx context.Context) ([]Token, error)
}
