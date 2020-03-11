package datasource

import (
	"io"
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
	// panic if initialize failed;
	// once initialized, listener should recover all panic and error.
	Initialize()
	// Close the data source.
	io.Closer
}
