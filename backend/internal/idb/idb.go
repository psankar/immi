package idb

import (
	"errors"
	"immi/pkg/dao"
)

// By making this an interface, we can potentially
// migrate to a different db if needed
type IDB interface {
	AppendImmis([]dao.Immi) error
}

var (
	ErrInternal = errors.New("INTERNAL_ERROR")
	ErrUser     = errors.New("USER_ERROR")
)
