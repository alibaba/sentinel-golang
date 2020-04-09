package datasource

type Code uint32

const (
	OK                  = iota
	ConvertSourceError  = 1
	UpdatePropertyError = 2
	HandleSourceError   = 3
)

func NewError(code Code, desc string) Error {
	return Error{
		code: code,
		desc: desc,
	}
}

type Error struct {
	code Code
	desc string
}

func (e Error) Code() Code {
	return e.code
}

func (e Error) Error() string {
	return e.desc
}
