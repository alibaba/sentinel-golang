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
	ReadSource() []byte
	// Initialize, init datasource and load initial rules
	// start listener
	Initialize()
	// Close the data source.
	io.Closer
}
