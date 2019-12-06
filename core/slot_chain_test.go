package core

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func init() {
	InitDefaultLoggerToConsole()
}

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
		sc.addStatPrepareSlotFirst(&StatPrepareSlotMock1{
			Name: "mock2" + strconv.Itoa(i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.addStatPrepareSlotFirst(&StatPrepareSlotMock1{
			Name: "mock1" + strconv.Itoa(i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.addStatPrepareSlotLast(&StatPrepareSlotMock1{
			Name: "mock3" + strconv.Itoa(i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.addStatPrepareSlotFirst(&StatPrepareSlotMock1{
			Name: "mock" + strconv.Itoa(i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.addStatPrepareSlotLast(&StatPrepareSlotMock1{
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

func (rcs *RuleCheckSlotMock1) Check(ctx *EntryContext) *RuleCheckResult {
	fmt.Println(rcs.Name)
	return nil
}
func TestSlotChain_addRuleCheckSlotFirstAndLast(t *testing.T) {
	sc := NewSlotChain()
	for i := 9; i >= 0; i-- {
		sc.addRuleCheckSlotFirst(&RuleCheckSlotMock1{
			Name: "mock2" + strconv.Itoa(i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.addRuleCheckSlotFirst(&RuleCheckSlotMock1{
			Name: "mock1" + strconv.Itoa(i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.addRuleCheckSlotLast(&RuleCheckSlotMock1{
			Name: "mock3" + strconv.Itoa(i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.addRuleCheckSlotFirst(&RuleCheckSlotMock1{
			Name: "mock" + strconv.Itoa(i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.addRuleCheckSlotLast(&RuleCheckSlotMock1{
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
func (ss *StatSlotMock1) OnEntryBlocked(ctx *EntryContext, blockEvent RuleBasedCheckBlockedEvent) {
	fmt.Println(ss.Name)
}
func (ss *StatSlotMock1) OnCompleted(ctx *EntryContext) {
	fmt.Println(ss.Name)
}
func TestSlotChain_addStatSlotFirstAndLast(t *testing.T) {
	sc := NewSlotChain()
	for i := 9; i >= 0; i-- {
		sc.addStatSlotFirst(&StatSlotMock1{
			Name: "mock2" + strconv.Itoa(i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.addStatSlotFirst(&StatSlotMock1{
			Name: "mock1" + strconv.Itoa(i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.addStatSlotLast(&StatSlotMock1{
			Name: "mock3" + strconv.Itoa(i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.addStatSlotFirst(&StatSlotMock1{
			Name: "mock" + strconv.Itoa(i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.addStatSlotLast(&StatSlotMock1{
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

type ResourceBuilderSlotMock struct {
	mock.Mock
}

func (m *ResourceBuilderSlotMock) Prepare(ctx *EntryContext) {
	m.Called(ctx)
	return
}

type FlowSlotMock struct {
	mock.Mock
}

func (m *FlowSlotMock) Check(ctx *EntryContext) *RuleCheckResult {
	arg := m.Called(ctx)
	return arg.Get(0).(*RuleCheckResult)
}

type DegradeSlotMock struct {
	mock.Mock
}

func (m *DegradeSlotMock) Check(ctx *EntryContext) *RuleCheckResult {
	arg := m.Called(ctx)
	return arg.Get(0).(*RuleCheckResult)
}

type StatisticSlotMock struct {
	mock.Mock
}

func (m *StatisticSlotMock) OnEntryPassed(ctx *EntryContext) {
	m.Called(ctx)
	return
}
func (m *StatisticSlotMock) OnEntryBlocked(ctx *EntryContext, blockEvent RuleBasedCheckBlockedEvent) {
	m.Called(ctx, blockEvent)
	return
}
func (m *StatisticSlotMock) OnCompleted(ctx *EntryContext) {
	m.Called(ctx)
	return
}

func TestSlotChain_Entry_Pass_And_Exit(t *testing.T) {
	sc := NewSlotChain()
	ctx := sc.GetContext()
	rw := &ResourceWrapper{
		ResourceName: "abc",
		FlowType:     InBound,
	}
	ctx.ResWrapper = rw
	ctx.StatNode = &NodeMock{}
	ctx.Count = 1
	ctx.Entry = NewCtEntry(ctx, rw, sc, ctx.StatNode)

	rbs := &ResourceBuilderSlotMock{}
	fsm := &FlowSlotMock{}
	dsm := &DegradeSlotMock{}
	ssm := &StatisticSlotMock{}
	sc.addStatPrepareSlotFirst(rbs)
	sc.addRuleCheckSlotFirst(fsm)
	sc.addRuleCheckSlotFirst(dsm)
	sc.addStatSlotFirst(ssm)

	rbs.On("Prepare", mock.Anything).Return()
	fsm.On("Check", mock.Anything).Return(NewSlotResultPass())
	dsm.On("Check", mock.Anything).Return(NewSlotResultPass())
	ssm.On("OnEntryPassed", mock.Anything).Return()
	ssm.On("OnCompleted", mock.Anything).Return()

	sc.entry(ctx)
	time.Sleep(time.Second * 1)
	sc.exit(ctx)
	rbs.AssertNumberOfCalls(t, "Prepare", 1)
	fsm.AssertNumberOfCalls(t, "Check", 1)
	dsm.AssertNumberOfCalls(t, "Check", 1)
	ssm.AssertNumberOfCalls(t, "OnEntryPassed", 1)
	ssm.AssertNumberOfCalls(t, "OnEntryBlocked", 0)
	ssm.AssertNumberOfCalls(t, "OnCompleted", 1)
}

func TestSlotChain_Entry_Block(t *testing.T) {
	sc := NewSlotChain()
	ctx := sc.GetContext()
	rw := &ResourceWrapper{
		ResourceName: "abc",
		FlowType:     InBound,
	}
	ctx.ResWrapper = rw
	ctx.StatNode = &NodeMock{}
	ctx.Count = 1
	ctx.Entry = NewCtEntry(ctx, rw, sc, ctx.StatNode)

	rbs := &ResourceBuilderSlotMock{}
	fsm := &FlowSlotMock{}
	dsm := &DegradeSlotMock{}
	ssm := &StatisticSlotMock{}
	sc.addStatPrepareSlotFirst(rbs)
	sc.addRuleCheckSlotFirst(fsm)
	sc.addRuleCheckSlotLast(dsm)
	sc.addStatSlotFirst(ssm)

	rbs.On("Prepare", mock.Anything).Return()
	fsm.On("Check", mock.Anything).Return(NewSlotResultPass())
	dsm.On("Check", mock.Anything).Return(NewSlotResultBlocked(UnknownEvent, "UnknownEvent"))
	ssm.On("OnEntryPassed", mock.Anything).Return()
	ssm.On("OnEntryBlocked", mock.Anything, mock.Anything).Return()
	ssm.On("OnCompleted", mock.Anything).Return()

	sc.entry(ctx)
	time.Sleep(time.Second * 1)
	sc.exit(ctx)

	rbs.AssertNumberOfCalls(t, "Prepare", 1)
	fsm.AssertNumberOfCalls(t, "Check", 1)
	dsm.AssertNumberOfCalls(t, "Check", 1)
	ssm.AssertNumberOfCalls(t, "OnEntryPassed", 0)
	ssm.AssertNumberOfCalls(t, "OnEntryBlocked", 1)
	ssm.AssertNumberOfCalls(t, "OnCompleted", 1)
}

type ResourceBuilderSlotMockPanic struct {
	mock.Mock
}

func (m *ResourceBuilderSlotMockPanic) Prepare(ctx *EntryContext) {
	m.Called(ctx)
	panic("unexpected panic")
	return
}

func TestSlotChain_Entry_With_Panic(t *testing.T) {
	sc := NewSlotChain()
	ctx := sc.GetContext()
	rw := &ResourceWrapper{
		ResourceName: "abc",
		FlowType:     InBound,
	}
	ctx.ResWrapper = rw
	ctx.StatNode = &NodeMock{}
	ctx.Count = 1
	ctx.Entry = NewCtEntry(ctx, rw, sc, ctx.StatNode)

	rbs := &ResourceBuilderSlotMockPanic{}
	fsm := &FlowSlotMock{}
	dsm := &DegradeSlotMock{}
	ssm := &StatisticSlotMock{}
	sc.addStatPrepareSlotFirst(rbs)
	sc.addRuleCheckSlotFirst(fsm)
	sc.addRuleCheckSlotLast(dsm)
	sc.addStatSlotFirst(ssm)

	rbs.On("Prepare", mock.Anything).Return()
	fsm.On("Check", mock.Anything).Return(NewSlotResultPass())
	dsm.On("Check", mock.Anything).Return(NewSlotResultBlocked(UnknownEvent, "UnknownEvent"))
	ssm.On("OnEntryPassed", mock.Anything).Return()
	ssm.On("OnEntryBlocked", mock.Anything, mock.Anything).Return()
	ssm.On("OnCompleted", mock.Anything).Return()

	sc.entry(ctx)
	time.Sleep(time.Second * 1)
	sc.exit(ctx)

	rbs.AssertNumberOfCalls(t, "Prepare", 1)
	fsm.AssertNumberOfCalls(t, "Check", 0)
	dsm.AssertNumberOfCalls(t, "Check", 0)
	ssm.AssertNumberOfCalls(t, "OnEntryPassed", 0)
	ssm.AssertNumberOfCalls(t, "OnEntryBlocked", 0)
	ssm.AssertNumberOfCalls(t, "OnCompleted", 1)
}
