package core

type BlockError struct {
	blockType BlockType
	blockMsg  string
	rule      SentinelRule
}

func (e *BlockError) BlockMsg() string {
	return e.blockMsg
}

func (e *BlockError) BlockType() BlockType {
	return e.blockType
}

func NewBlockError(blockType BlockType, blockMsg string) *BlockError {
	return &BlockError{blockType: blockType, blockMsg: blockMsg}
}

func (e *BlockError) Error() string {
	return "SentinelBlockException: " + e.blockMsg
}
