package immi

import (
	"immi/internal/common"
	"net/http"
)

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

type NewImmi struct {
	Msg string
}

type NewListy struct {
	Name string
}

// Errors exposed to the clients from the backend
var (
	ErrDuplicateUsername = &common.Error{
		Err:      "ERROR_DUPLICATE_USERNAME",
		HTTPCode: http.StatusConflict,
	}
	ErrInvalidUsername = &common.Error{
		Err:      "ERROR_INVALID_USERNAME",
		HTTPCode: http.StatusBadRequest,
	}
	ErrInvalidPassword = &common.Error{
		Err:      "ERROR_INVALID_PASSWORD",
		HTTPCode: http.StatusBadRequest,
	}
	ErrImmiInternal = &common.Error{
		Err:      "ERROR_IMMI_INTERNAL",
		HTTPCode: http.StatusInternalServerError,
	}
	ErrAuthenticationFailed = &common.Error{
		Err:      "ERROR_AUTHENTICATION_FAILED",
		HTTPCode: http.StatusUnauthorized,
	}

	ErrDuplicateListyName = &common.Error{
		Err:      "ERROR_DUPLICATE_LISTNAME",
		HTTPCode: http.StatusConflict,
	}
)
