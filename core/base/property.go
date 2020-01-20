package base

// Each property Consumer implement this interface to consume the newest properties
type PropertyConsumer interface {
	// property Consumer need to implement this method to update properties dynamically
	// Return nil if the consumer succeed to consume otherwise return error
	ConsumeConfig(value interface{}) error
}

// Each datasource implement this interface to publish the newest properties
type PropertyPublisher interface {
	// Each datasource need to implement this method to publish the newest properties
	// Return nil if the Publisher succeed to publish otherwise return error
	PublishConfig(value interface{}) error
}
