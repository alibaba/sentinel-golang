package base

import (
	"context"
)

type Context struct {
	name         string
	entranceNode DefaultNode
	curEntry     Entry
	origin       string
	context      context.Context
}
