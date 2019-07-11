/**
 * @description:
 *
 * @author: helloworld
 * @date:2019-07-11
 */
package base

type ResourceWrapper struct {
	// unique resource name
	ResourceName string
	//
	ResourceType int
}

type SlotResultStatus int8

const (
	ResultStatusPass = iota
	ResultStatusBlocked
	ResultStatusWait
	ResultStatusError
)

type TokenResult struct {
	Status        SlotResultStatus
	BlockedReason string
	WaitMs        uint64
	ErrorMsg      string
}

func NewSlotResultPass() *TokenResult {
	return &TokenResult{Status: ResultStatusPass}
}

func NewSlotResultBlock(blockedReason string) *TokenResult {
	return &TokenResult{Status: ResultStatusBlocked, BlockedReason: blockedReason}
}

func NewSlotResultWait(waitMs uint64) *TokenResult {
	return &TokenResult{Status: ResultStatusWait, WaitMs: waitMs}
}

func NewSlotResultError(errorMsg string) *TokenResult {
	return &TokenResult{Status: ResultStatusError, ErrorMsg: errorMsg}
}
