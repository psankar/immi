package immi

import "errors"

// The contents of this file are exposed to the client code.
// So, changes to this should always be backwards compatible.

type SignUp struct {
	Username     string
	EmailAddress string
	Password     string
}

type Login struct {
	Username string
	Password string
}

const UserHeader = "X-IMMI-USER"

var (
	ErrImmiInternal = errors.New("ERROR_IMMI_INTERNAL")
)

type NewImmi struct {
	Msg string
}

type UserError error

// Immi User Errors
var (
	ErrDuplicateUsername UserError = errors.New("ERROR_DUPLICATE_USERNAME")
	ErrInvalidUsername   UserError = errors.New("ERROR_INVALID_USERNAME")
	ErrInvalidPassword   UserError = errors.New("ERROR_INVALID_PASSWORD")
)
