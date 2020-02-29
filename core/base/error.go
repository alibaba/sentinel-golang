package base

import (
	"errors"
)

var (
	ErrFlowRuleChanBlocked = errors.New("blocked flow rule chan")
	ErrSystemRuleChanBlocked = errors.New("blocked system rule chan")
)
