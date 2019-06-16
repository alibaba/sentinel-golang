package data

type UnaryLeapArray struct {
	LeapArray
}

func (uls *UnaryLeapArray) newEmptyBucket(startTime uint64) interface{} {
	return uint64(0)
}

func (uls *UnaryLeapArray) resetWindowTo(ww *WindowWrap, startTime uint64) (*WindowWrap, error) {
	ww.WindowStart = startTime
	ww.Value = uint64(0)
	return ww, nil
}
