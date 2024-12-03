package oops

import "errors"

var ErrDuplicateUser = errors.New("login duplication")
var ErrNoUser = errors.New("no user")
var ErrDupAccess = errors.New("not unique acess token")
var ErrDupRefresh = errors.New("not unique refresh token")
var ErrNoTokens = errors.New("impossible to generate token")
