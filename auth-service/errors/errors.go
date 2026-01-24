package errors

import (
	"github.com/bpurdy1/auth-service/account"
	"github.com/bpurdy1/auth-service/metadata"
)

var (
	ErrUserNotFound       = account.ErrUserNotFound
	ErrUserAlreadyExists  = account.ErrUserAlreadyExists
	ErrInvalidCredentials = account.ErrInvalidCredentials
	ErrMetadataNotFound   = metadata.ErrMetadataNotFound
)
