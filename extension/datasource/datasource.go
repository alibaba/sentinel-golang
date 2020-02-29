package datasource

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/alibaba/sentinel-golang/core/base"
	"gopkg.in/yaml.v2"
)

type Base struct {
	DataFormat string
}

// BuildDecoder, all datasource should
func (b Base) buildDecoder(reader io.Reader) base.PropertyDecoder {
	if len(b.DataFormat) > 1 && b.DataFormat[0] == '.' {
		b.DataFormat = b.DataFormat[1:]
	}
	switch b.DataFormat {
	case "json":
		return json.NewDecoder(reader)
	case "yaml":
		return yaml.NewDecoder(reader)
	case "ini", "toml", "hcl":
		panic("unsupported data format by now")
	}
	panic("invalid data format")
}

func (b Base) ApplyConfig(rawBytes []byte) error {
	return base.UpdateProperty(func() base.PropertyDecoder {
		return b.buildDecoder(bytes.NewBuffer(rawBytes))
	})
}

func (b Base) DeleteConfig() error {
	return base.DeleteProperty()
}
