package base

import "fmt"

// BlockError indicates the request was blocked by Sentinel.
type BlockError struct {
	blockType BlockType
	// blockMsg provides additional message for the block error.
	blockMsg string

	rule SentinelRule
	// snapshotValue represents the triggered "snapshot" value
	snapshotValue interface{}
}

func (e *BlockError) BlockMsg() string {
	return e.blockMsg
}

func (e *BlockError) BlockType() BlockType {
	return e.blockType
}

func (e *BlockError) TriggeredRule() SentinelRule {
	return e.rule
}

func (e *BlockError) TriggeredValue() interface{} {
	return e.snapshotValue
}

func NewBlockErrorFromDeepCopy(from *BlockError) *BlockError {
	return &BlockError{
		blockType:     from.blockType,
		blockMsg:      from.blockMsg,
		rule:          from.rule,
		snapshotValue: from.snapshotValue,
	}
}

func NewBlockError(blockType BlockType) *BlockError {
	return &BlockError{blockType: blockType}
}

func NewBlockErrorWithMessage(blockType BlockType, message string) *BlockError {
	return &BlockError{blockType: blockType, blockMsg: message}
}

func NewBlockErrorWithCause(blockType BlockType, blockMsg string, rule SentinelRule, snapshot interface{}) *BlockError {
	return &BlockError{blockType: blockType, blockMsg: blockMsg, rule: rule, snapshotValue: snapshot}
}

func (e *BlockError) Error() string {
	if len(e.blockMsg) == 0 {
		return fmt.Sprintf("SentinelBlockError: %s", e.blockType.String())
	}
	return fmt.Sprintf("SentinelBlockError: %s, message: %s", e.blockType.String(), e.blockMsg)
}
