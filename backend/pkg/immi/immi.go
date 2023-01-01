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
	RouteName   string
	DisplayName string
}

type Graf struct {
	ListRouteName string
	Username      string
}

type SubscribeListyTimeline struct {
	// For now, we will keep all Listys private
	ListyRouteName string
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
	ErrDuplicateDisplayName = &common.Error{
		Err:      "ERROR_DUPLICATE_DISPLAYNAME",
		HTTPCode: http.StatusConflict,
	}
	ErrDuplicateRouteName = &common.Error{
		Err:      "ERROR_DUPLICATE_ROUTENAME",
		HTTPCode: http.StatusConflict,
	}
	ErrListAddFailed = &common.Error{
		Err:      "ERROR_LIST_ADD_FAILED",
		HTTPCode: http.StatusBadRequest,
	}
)
