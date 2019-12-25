package api

import (
	"github.com/sentinel-group/sentinel-golang/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type prepareSlotMock struct {
	mock.Mock
}

func (m *prepareSlotMock) Prepare(ctx *core.EntryContext) {
	m.Called(ctx)
	return
}

type mockRuleCheckSlot1 struct {
	mock.Mock
}

func (m *mockRuleCheckSlot1) Check(ctx *core.EntryContext) *core.TokenResult {
	arg := m.Called(ctx)
	return arg.Get(0).(*core.TokenResult)
}

type mockRuleCheckSlot2 struct {
	mock.Mock
}

func (m *mockRuleCheckSlot2) Check(ctx *core.EntryContext) *core.TokenResult {
	arg := m.Called(ctx)
	return arg.Get(0).(*core.TokenResult)
}

type statisticSlotMock struct {
	mock.Mock
}

func (m *statisticSlotMock) OnEntryPassed(ctx *core.EntryContext) {
	m.Called(ctx)
	return
}
func (m *statisticSlotMock) OnEntryBlocked(ctx *core.EntryContext, blockError *core.BlockError) {
	m.Called(ctx, blockError)
	return
}
func (m *statisticSlotMock) OnCompleted(ctx *core.EntryContext) {
	m.Called(ctx)
	return
}

func Test_entryWithArgsAndChainPass(t *testing.T) {
	sc := core.NewSlotChain()
	ps1 := &prepareSlotMock{}
	rcs1 := &mockRuleCheckSlot1{}
	rcs2 := &mockRuleCheckSlot2{}
	ssm := &statisticSlotMock{}
	sc.AddStatPrepareSlotFirst(ps1)
	sc.AddRuleCheckSlotFirst(rcs1)
	sc.AddRuleCheckSlotFirst(rcs2)
	sc.AddStatSlotFirst(ssm)

	ps1.On("Prepare", mock.Anything).Return()
	rcs1.On("Check", mock.Anything).Return(core.NewTokenResultPass())
	rcs2.On("Check", mock.Anything).Return(core.NewTokenResultPass())
	ssm.On("OnEntryPassed", mock.Anything).Return()
	ssm.On("OnCompleted", mock.Anything).Return()

	entry, b := entryWithArgsAndChain("abc", core.ResTypeCommon, core.Inbound, 1, 0, sc)
	assert.Nil(t, b, "the entry should not be blocked")
	assert.Equal(t, "abc", entry.Resource().Name())

	entry.Exit()

	ps1.AssertNumberOfCalls(t, "Prepare", 1)
	rcs1.AssertNumberOfCalls(t, "Check", 1)
	rcs2.AssertNumberOfCalls(t, "Check", 1)
	ssm.AssertNumberOfCalls(t, "OnEntryPassed", 1)
	ssm.AssertNumberOfCalls(t, "OnEntryBlocked", 0)
	ssm.AssertNumberOfCalls(t, "OnCompleted", 1)
}

func Test_entryWithArgsAndChainBlock(t *testing.T) {
	sc := core.NewSlotChain()
	ps1 := &prepareSlotMock{}
	rcs1 := &mockRuleCheckSlot1{}
	rcs2 := &mockRuleCheckSlot2{}
	ssm := &statisticSlotMock{}
	sc.AddStatPrepareSlotFirst(ps1)
	sc.AddRuleCheckSlotLast(rcs1)
	sc.AddRuleCheckSlotLast(rcs2)
	sc.AddStatSlotFirst(ssm)

	blockType := core.BlockTypeFlow

	ps1.On("Prepare", mock.Anything).Return()
	rcs1.On("Check", mock.Anything).Return(core.NewTokenResultBlocked(blockType, "Flow"))
	rcs2.On("Check", mock.Anything).Return(core.NewTokenResultPass())
	ssm.On("OnEntryPassed", mock.Anything).Return()
	ssm.On("OnEntryBlocked", mock.Anything, mock.Anything).Return()
	ssm.On("OnCompleted", mock.Anything).Return()

	entry, b := entryWithArgsAndChain("abc", core.ResTypeCommon, core.Inbound, 1, 0, sc)
	assert.Nil(t, entry)
	assert.NotNil(t, b)
	assert.Equal(t, blockType, b.BlockType())

	ps1.AssertNumberOfCalls(t, "Prepare", 1)
	rcs1.AssertNumberOfCalls(t, "Check", 1)
	rcs2.AssertNumberOfCalls(t, "Check", 0)
	ssm.AssertNumberOfCalls(t, "OnEntryPassed", 0)
	ssm.AssertNumberOfCalls(t, "OnEntryBlocked", 1)
	ssm.AssertNumberOfCalls(t, "OnCompleted", 1)
}
