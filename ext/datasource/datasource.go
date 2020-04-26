package datasource

import (
	"fmt"
	"io"

	"go.uber.org/multierr"
)

// The generic interface to describe the datasource
// Each DataSource instance listen in one property type.
type DataSource interface {
	// Add specified property handler in current datasource
	AddPropertyHandler(h PropertyHandler)
	// Remove specified property handler in current datasource
	RemovePropertyHandler(h PropertyHandler)
	// Read original data from the data source.
	// return source bytes if succeed to read, if not, return error when reading
	ReadSource() ([]byte, error)
	// Initialize the datasource and load initial rules
	// start listener to listen on dynamic source
	// return error if initialize failed;
	// once initialized, listener should recover all panic and error.
	Initialize() error
	// Close the data source.
	io.Closer
}

type Base struct {
	handlers []PropertyHandler
}

func (b *Base) Handle(src []byte) (err error) {
	for _, h := range b.handlers {
		e := h.Handle(src)
		if e != nil {
			err = multierr.Append(err, e)
		}
	}
	if err == nil {
		return nil
	}
	return Error{code: HandleSourceError, desc: fmt.Sprintf("%+v", err)}
}

// return idx if existed, else return -1
func (b *Base) indexOfHandler(h PropertyHandler) int {
	for idx, handler := range b.handlers {
		if handler == h {
			return idx
		}
	}
	return -1
}

func (b *Base) AddPropertyHandler(h PropertyHandler) {
	if h == nil || b.indexOfHandler(h) >= 0 {
		return
	}
	b.handlers = append(b.handlers, h)
}

func (b *Base) RemovePropertyHandler(h PropertyHandler) {
	if h == nil {
		return
	}
	idx := b.indexOfHandler(h)
	if idx < 0 {
		return
	}
	b.handlers = append(b.handlers[:idx], b.handlers[idx+1:]...)
}
