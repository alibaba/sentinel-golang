package core

import (
	"fmt"
)

type RuleBasedCheckBlockedEvent int8

const (
	UnknownEvent RuleBasedCheckBlockedEvent = iota
)

type SlotResultStatus int8

const (
	ResultStatusPass SlotResultStatus = iota
	ResultStatusBlocked
)

type RuleCheckResult struct {
	Status       SlotResultStatus
	BlockedEvent RuleBasedCheckBlockedEvent
	BlockedMsg   string
}

func (r *RuleCheckResult) status() string {
	if r.Status == ResultStatusPass {
		return "ResultStatusPass"
	} else if r.Status == ResultStatusBlocked {
		return "ResultStatusBlocked"
	} else {
		return "Unknown"
	}
}

func (r *RuleCheckResult) toString() string {
	return fmt.Sprintf("check result:%s; BlockedEvent is:%v; BlockedMsg is:%s;", r.status(), r.BlockedEvent, r.BlockedMsg)
}

func NewSlotResultPass() *RuleCheckResult {
	return &RuleCheckResult{Status: ResultStatusPass}
}

func NewSlotResultBlocked(blockEvent RuleBasedCheckBlockedEvent, blockReason string) *RuleCheckResult {
	return &RuleCheckResult{
		Status:       ResultStatusBlocked,
		BlockedEvent: blockEvent,
		BlockedMsg:   blockReason,
	}
}
