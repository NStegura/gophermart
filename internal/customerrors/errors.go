package customerrors

import (
	"errors"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrAlreadyExists       = errors.New("already exists")
	ErrCurrUserUploaded    = errors.New("order already uploaded by current user")
	ErrAnotherUserUploaded = errors.New("order already uploaded by another user")
)
