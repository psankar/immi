package common

type Error struct {
	Err      string
	HTTPCode int
}

func (e *Error) Error() string {
	return e.Err
}

func Err(err error, code int) *Error {
	return &Error{
		Err:      err.Error(),
		HTTPCode: code,
	}
}
