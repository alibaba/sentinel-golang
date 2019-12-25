package core

import "fmt"

type SentinelRule interface {
	fmt.Stringer

	ResourceName() string
}
