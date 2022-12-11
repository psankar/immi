package immi

// The contents of this file are exposed to the client code.
// So, changes to this should always be backwards compatible.

const UserHeader = "X-IMMI-USER"

type NewImmi struct {
	Msg string
}
