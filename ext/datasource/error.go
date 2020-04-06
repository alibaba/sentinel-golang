package datasource

type Code uint32

const (
	OK                  = iota
	ConvertSourceError  = 1
	UpdatePropertyError = 2
	HandleSourceError   = 3

	EtcdKeyNotExistedError = 4
	EtcdGetValueError      = 5
)

func NewDSError(code Code, desc string) DSError {
	return DSError{
		code: code,
		desc: desc,
	}
}

type DSError struct {
	code Code
	desc string
}

func (e DSError) Code() Code {
	return e.code
}

func (e DSError) Error() string {
	return e.desc
}
