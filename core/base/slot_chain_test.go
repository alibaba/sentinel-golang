package base

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type StatPrepareSlotMock1 struct {
	Name string
}

func (spl *StatPrepareSlotMock1) Prepare(ctx *EntryContext) {
	fmt.Println(spl.Name)
	return
}

func TestSlotChain_addStatPrepareSlotFirstAndLast(t *testing.T) {
	sc := NewSlotChain()
	for i := 9; i >= 0; i-- {
		sc.AddStatPrepareSlotFirst(&StatPrepareSlotMock1{
			Name: "mock2" + strconv.Itoa(i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.AddStatPrepareSlotFirst(&StatPrepareSlotMock1{
			Name: "mock1" + strconv.Itoa(i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.AddStatPrepareSlotLast(&StatPrepareSlotMock1{
			Name: "mock3" + strconv.Itoa(i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.AddStatPrepareSlotFirst(&StatPrepareSlotMock1{
			Name: "mock" + strconv.Itoa(i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.AddStatPrepareSlotLast(&StatPrepareSlotMock1{
			Name: "mock4" + strconv.Itoa(i),
		})
	}

	spSlice := sc.statPres
	if len(spSlice) != 50 {
		t.Error("len error")
	}

	for idx, slot := range spSlice {
		n := "mock" + strconv.Itoa(idx)
		spsm, ok := slot.(*StatPrepareSlotMock1)
		if !ok {
			t.Error("type error")
		}
		reflect.DeepEqual(n, spsm.Name)
	}
}

type RuleCheckSlotMock1 struct {
	Name string
}

func (rcs *RuleCheckSlotMock1) Check(ctx *EntryContext) *TokenResult {
	fmt.Println(rcs.Name)
	return nil
}
func TestSlotChain_addRuleCheckSlotFirstAndLast(t *testing.T) {
	sc := NewSlotChain()
	for i := 9; i >= 0; i-- {
		sc.AddRuleCheckSlotFirst(&RuleCheckSlotMock1{
			Name: "mock2" + strconv.Itoa(i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.AddRuleCheckSlotFirst(&RuleCheckSlotMock1{
			Name: "mock1" + strconv.Itoa(i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.AddRuleCheckSlotLast(&RuleCheckSlotMock1{
			Name: "mock3" + strconv.Itoa(i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.AddRuleCheckSlotFirst(&RuleCheckSlotMock1{
			Name: "mock" + strconv.Itoa(i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.AddRuleCheckSlotLast(&RuleCheckSlotMock1{
			Name: "mock4" + strconv.Itoa(i),
		})
	}

	spSlice := sc.ruleChecks
	if len(spSlice) != 50 {
		t.Error("len error")
	}

	for idx, slot := range spSlice {
		n := "mock" + strconv.Itoa(idx)
		spsm, ok := slot.(*RuleCheckSlotMock1)
		if !ok {
			t.Error("type error")
		}
		reflect.DeepEqual(n, spsm.Name)
	}
}

type StatSlotMock1 struct {
	Name string
}

func (ss *StatSlotMock1) OnEntryPassed(ctx *EntryContext) {
	fmt.Println(ss.Name)
}
func (ss *StatSlotMock1) OnEntryBlocked(ctx *EntryContext, blockError *BlockError) {
	fmt.Printf("%s blocked: %v\n", ss.Name, blockError)
}
func (ss *StatSlotMock1) OnCompleted(ctx *EntryContext) {
	fmt.Println(ss.Name)
}
func TestSlotChain_addStatSlotFirstAndLast(t *testing.T) {
	sc := NewSlotChain()
	for i := 9; i >= 0; i-- {
		sc.AddStatSlotFirst(&StatSlotMock1{
			Name: "mock2" + strconv.Itoa(i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.AddStatSlotFirst(&StatSlotMock1{
			Name: "mock1" + strconv.Itoa(i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.AddStatSlotLast(&StatSlotMock1{
			Name: "mock3" + strconv.Itoa(i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.AddStatSlotFirst(&StatSlotMock1{
			Name: "mock" + strconv.Itoa(i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.AddStatSlotLast(&StatSlotMock1{
			Name: "mock4" + strconv.Itoa(i),
		})
	}

	spSlice := sc.stats
	if len(spSlice) != 50 {
		t.Error("len error")
	}

	for idx, slot := range spSlice {
		n := "mock" + strconv.Itoa(idx)
		spsm, ok := slot.(*StatSlotMock1)
		assert.True(t, ok, "slot type must be StatSlotMock1")
		if !ok {
			t.Error("type error")
		}
		reflect.DeepEqual(n, spsm.Name)
	}
}

type prepareSlotMock struct {
	mock.Mock
}

func (m *prepareSlotMock) Prepare(ctx *EntryContext) {
	m.Called(ctx)
	return
}

type mockRuleCheckSlot1 struct {
	mock.Mock
}

func (m *mockRuleCheckSlot1) Check(ctx *EntryContext) *TokenResult {
	arg := m.Called(ctx)
	return arg.Get(0).(*TokenResult)
}

type mockRuleCheckSlot2 struct {
	mock.Mock
}

func (m *mockRuleCheckSlot2) Check(ctx *EntryContext) *TokenResult {
	arg := m.Called(ctx)
	return arg.Get(0).(*TokenResult)
}

type statisticSlotMock struct {
	mock.Mock
}

func (m *statisticSlotMock) OnEntryPassed(ctx *EntryContext) {
	m.Called(ctx)
	return
}
func (m *statisticSlotMock) OnEntryBlocked(ctx *EntryContext, blockError *BlockError) {
	m.Called(ctx, blockError)
	return
}
func (m *statisticSlotMock) OnCompleted(ctx *EntryContext) {
	m.Called(ctx)
	return
}

func TestSlotChain_Entry_Pass_And_Exit(t *testing.T) {
	sc := NewSlotChain()
	ctx := sc.GetPooledContext()
	rw := NewResourceWrapper("abc", ResTypeCommon, Inbound)
	ctx.Resource = rw
	ctx.SetEntry(NewSentinelEntry(ctx, rw, sc))
	ctx.StatNode = &StatNodeMock{}
	ctx.Input = &SentinelInput{
		AcquireCount: 1,
		Flag:         0,
		Args:         nil,
		Attachments:  nil,
	}

	ps1 := &prepareSlotMock{}
	rcs1 := &mockRuleCheckSlot1{}
	rcs2 := &mockRuleCheckSlot2{}
	ssm := &statisticSlotMock{}
	sc.AddStatPrepareSlotFirst(ps1)
	sc.AddRuleCheckSlotFirst(rcs1)
	sc.AddRuleCheckSlotFirst(rcs2)
	sc.AddStatSlotFirst(ssm)

	ps1.On("Prepare", mock.Anything).Return()
	rcs1.On("Check", mock.Anything).Return(NewTokenResultPass())
	rcs2.On("Check", mock.Anything).Return(NewTokenResultPass())
	ssm.On("OnEntryPassed", mock.Anything).Return()
	ssm.On("OnCompleted", mock.Anything).Return()

	r := sc.Entry(ctx)
	assert.Equal(t, ResultStatusPass, r.status, "expected to pass but actually blocked")
	time.Sleep(time.Millisecond * 100)

	sc.exit(ctx)

	ps1.AssertNumberOfCalls(t, "Prepare", 1)
	rcs1.AssertNumberOfCalls(t, "Check", 1)
	rcs2.AssertNumberOfCalls(t, "Check", 1)
	ssm.AssertNumberOfCalls(t, "OnEntryPassed", 1)
	ssm.AssertNumberOfCalls(t, "OnEntryBlocked", 0)
	ssm.AssertNumberOfCalls(t, "OnCompleted", 1)
}

func TestSlotChain_Entry_Block(t *testing.T) {
	sc := NewSlotChain()
	ctx := sc.GetPooledContext()
	rw := NewResourceWrapper("abc", ResTypeCommon, Inbound)
	ctx.SetEntry(NewSentinelEntry(ctx, rw, sc))
	ctx.Resource = rw
	ctx.StatNode = &StatNodeMock{}
	ctx.Input = &SentinelInput{
		AcquireCount: 1,
		Flag:         0,
		Args:         nil,
		Attachments:  nil,
	}

	rbs := &prepareSlotMock{}
	fsm := &mockRuleCheckSlot1{}
	dsm := &mockRuleCheckSlot2{}
	ssm := &statisticSlotMock{}
	sc.AddStatPrepareSlotFirst(rbs)
	sc.AddRuleCheckSlotFirst(fsm)
	sc.AddRuleCheckSlotLast(dsm)
	sc.AddStatSlotFirst(ssm)

	blockType := BlockTypeFlow

	rbs.On("Prepare", mock.Anything).Return()
	fsm.On("Check", mock.Anything).Return(NewTokenResultPass())
	dsm.On("Check", mock.Anything).Return(NewTokenResultBlocked(blockType))
	ssm.On("OnEntryPassed", mock.Anything).Return()
	ssm.On("OnEntryBlocked", mock.Anything, mock.Anything).Return()
	ssm.On("OnCompleted", mock.Anything).Return()

	r := sc.Entry(ctx)
	assert.True(t, r.IsBlocked(), "expected to be blocked but actually passed")
	if r.blockErr == nil || r.blockErr.blockType != blockType {
		t.Fatalf("invalid block error: expected blockType is %v", blockType)
		return
	}
	sc.exit(ctx)

	rbs.AssertNumberOfCalls(t, "Prepare", 1)
	fsm.AssertNumberOfCalls(t, "Check", 1)
	dsm.AssertNumberOfCalls(t, "Check", 1)
	ssm.AssertNumberOfCalls(t, "OnEntryPassed", 0)
	ssm.AssertNumberOfCalls(t, "OnEntryBlocked", 1)
	ssm.AssertNumberOfCalls(t, "OnCompleted", 0)
}

type badPrepareSlotMock struct {
	mock.Mock
}

func (m *badPrepareSlotMock) Prepare(ctx *EntryContext) {
	m.Called(ctx)
	panic("sentinel internal panic")
}

func TestSlotChain_Entry_With_Panic(t *testing.T) {
	sc := NewSlotChain()
	ctx := sc.GetPooledContext()
	rw := NewResourceWrapper("abc", ResTypeCommon, Inbound)
	ctx.Resource = rw
	statNodeMock := &StatNodeMock{}
	statNodeMock.On("AddErrorRequest", mock.Anything).Return()
	ctx.StatNode = statNodeMock
	ctx.Input = &SentinelInput{
		AcquireCount: 1,
		Flag:         0,
		Args:         nil,
		Attachments:  nil,
	}

	rbs := &badPrepareSlotMock{}
	fsm := &mockRuleCheckSlot1{}
	dsm := &mockRuleCheckSlot2{}
	ssm := &statisticSlotMock{}
	sc.AddStatPrepareSlotFirst(rbs)
	sc.AddRuleCheckSlotFirst(fsm)
	sc.AddRuleCheckSlotLast(dsm)
	sc.AddStatSlotFirst(ssm)

	rbs.On("Prepare", mock.Anything).Return()
	fsm.On("Check", mock.Anything).Return(NewTokenResultPass())
	dsm.On("Check", mock.Anything).Return(NewTokenResultBlocked(BlockTypeUnknown))
	ssm.On("OnEntryPassed", mock.Anything).Return()
	ssm.On("OnEntryBlocked", mock.Anything, mock.Anything).Return()
	ssm.On("OnCompleted", mock.Anything).Return()

	r := sc.Entry(ctx)
	assert.Nil(t, r, "internal error in slots should recover and yield nil TokenResult")

	rbs.AssertNumberOfCalls(t, "Prepare", 1)
	fsm.AssertNumberOfCalls(t, "Check", 0)
	dsm.AssertNumberOfCalls(t, "Check", 0)
	ssm.AssertNumberOfCalls(t, "OnEntryPassed", 0)
	ssm.AssertNumberOfCalls(t, "OnEntryBlocked", 0)
}
