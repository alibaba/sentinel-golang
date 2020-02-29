package base

import (
	"bytes"
	"encoding/json"
	"testing"

	"gotest.tools/assert"
)

func TestPropertyConsumer(t *testing.T) {
	RegisterPropertyConsumer(func(decoder PropertyDecoder) error {
		type Data struct {
			A int `json:"a" toml:"a" yaml:"a"`
			B string `json:"b" toml:"b" yaml:"b"`
		}

		var data Data
		err :=  decoder.Decode(&data)
		return err
	}, func() error {
		return nil
	})

	assert.NilError(t, UpdateProperty(func() PropertyDecoder {
		buff := bytes.NewBuffer([]byte(`{"a":1,"b":"2"}`))
		return json.NewDecoder(buff)
	}))

	// todo(gorexlv): add concurrency testing
}
