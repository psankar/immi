package idb

import (
	"context"
	"immi/pkg/dao"
)

// By making this an interface, we can potentially
// migrate to a different db if needed
type IDB interface {
	AppendImmis(context.Context, []dao.Immi) error
	CreateUser(context.Context, dao.User) error
}
