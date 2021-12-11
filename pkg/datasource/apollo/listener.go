package apollo

import (
	"github.com/apolloconfig/agollo/v4/storage"
)

type customChangeListener struct {
	ds *apolloDatasource
}

func (c *customChangeListener) OnChange(event *storage.ChangeEvent) {
	for key, value := range event.Changes {
		if c.ds.propertyKey == key {
			c.ds.handle([]byte(value.NewValue.(string)))
		}
	}
}

func (c *customChangeListener) OnNewestChange(event *storage.FullChangeEvent) {
	for key, value := range event.Changes {
		if c.ds.propertyKey == key {
			c.ds.handle([]byte(value.(string)))
		}
	}
}
