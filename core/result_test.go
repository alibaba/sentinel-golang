package core

import "testing"

func TestNewSlotResultBlock_normal(t *testing.T) {
	r := NewSlotResultBlocked(UnknownEvent, "UnknownEvent")
	t.Log(r.toString())
}
