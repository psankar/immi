package idb

import (
	"context"
	"immi/internal/common"
	"immi/pkg/dao"
)

// By making this an interface, we can potentially
// migrate to a different db if needed
type IDB interface {
	AppendImmis(context.Context, []dao.Immi) *common.Error
	CreateUser(context.Context, dao.User) *common.Error
	GetUser(ctx context.Context, username string) (dao.User, *common.Error)
}
