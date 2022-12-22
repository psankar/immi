package dao

import "time"

// The structs in this file should always be consistent
// with the database schema

type Immi struct {
	ID     string
	UserID int64
	Msg    string
	CTime  time.Time
}

var (
	ActiveUser   = "ACTIVE_USER"
	DisabledUser = "DISABLED_USER"
)

type User struct {
	Username     string
	EmailAddress string
	PasswordHash string
	UserState    string
}
