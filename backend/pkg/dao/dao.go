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
