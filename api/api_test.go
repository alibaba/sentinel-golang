package api

import (
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type prepareSlotMock struct {
	base.SlotOrder
	mock.Mock
}

func (m *prepareSlotMock) Name() string {
	return "mock-sentinel-prepare-slot"
}

func (m *prepareSlotMock) Prepare(ctx *base.EntryContext) {
	m.Called(ctx)
	return
}

type mockRuleCheckSlot1 struct {
	base.SlotOrder
	mock.Mock
}

func (m *mockRuleCheckSlot1) Name() string {
	return "mock-sentinel-rule-check-slot1"
}

func (m *mockRuleCheckSlot1) Check(ctx *base.EntryContext) *base.TokenResult {
	arg := m.Called(ctx)
	return arg.Get(0).(*base.TokenResult)
}

type mockRuleCheckSlot2 struct {
	base.SlotOrder
	mock.Mock
}

func (m *mockRuleCheckSlot2) Name() string {
	return "mock-sentinel-rule-check-slot2"
}

func (m *mockRuleCheckSlot2) Check(ctx *base.EntryContext) *base.TokenResult {
	arg := m.Called(ctx)
	return arg.Get(0).(*base.TokenResult)
}

type statisticSlotMock struct {
	base.SlotOrder
	mock.Mock
}

func (m *statisticSlotMock) Name() string {
	return "mock-sentinel-stat-check-slot"
}

func (m *statisticSlotMock) OnEntryPassed(ctx *base.EntryContext) {
	m.Called(ctx)
	return
}
func (m *statisticSlotMock) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
	m.Called(ctx, blockError)
	return
}
func (m *statisticSlotMock) OnCompleted(ctx *base.EntryContext) {
	m.Called(ctx)
	return
}

func Test_entryWithArgsAndChainPass(t *testing.T) {
	sc := base.NewSlotChain()
	ps1 := &prepareSlotMock{}
	rcs1 := &mockRuleCheckSlot1{}
	rcs2 := &mockRuleCheckSlot2{}
	ssm := &statisticSlotMock{}
	sc.InsertStatPrepareSlotByOrder(ps1)
	sc.InsertRuleCheckSlotByOrder(rcs1)
	sc.InsertRuleCheckSlotByOrder(rcs2)
	sc.InsertStatSlotByOrder(ssm)

	ps1.On("Prepare", mock.Anything).Return()
	rcs1.On("Check", mock.Anything).Return(base.NewTokenResultPass())
	rcs2.On("Check", mock.Anything).Return(base.NewTokenResultPass())
	ssm.On("OnEntryPassed", mock.Anything).Return()
	ssm.On("OnCompleted", mock.Anything).Return()

	entry, b := entry("abc", &EntryOptions{
		resourceType: base.ResTypeCommon,
		entryType:    base.Inbound,
		batchCount:   1,
		flag:         0,
		slotChain:    sc,
	})
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
	sc := base.NewSlotChain()
	ps1 := &prepareSlotMock{}
	rcs1 := &mockRuleCheckSlot1{}
	rcs2 := &mockRuleCheckSlot2{}
	ssm := &statisticSlotMock{}
	sc.InsertStatPrepareSlotByOrder(ps1)
	sc.InsertRuleCheckSlotByOrder(rcs1)
	sc.InsertRuleCheckSlotByOrder(rcs2)
	sc.InsertStatSlotByOrder(ssm)

	blockType := base.BlockTypeFlow

	ps1.On("Prepare", mock.Anything).Return()
	rcs1.On("Check", mock.Anything).Return(base.NewTokenResultBlocked(blockType))
	rcs2.On("Check", mock.Anything).Return(base.NewTokenResultPass())
	ssm.On("OnEntryPassed", mock.Anything).Return()
	ssm.On("OnEntryBlocked", mock.Anything, mock.Anything).Return()
	ssm.On("OnCompleted", mock.Anything).Return()

	entry, b := entry("abc", &EntryOptions{
		resourceType: base.ResTypeCommon,
		entryType:    base.Inbound,
		batchCount:   1,
		flag:         0,
		slotChain:    sc,
	})
	assert.Nil(t, entry)
	assert.NotNil(t, b)
	assert.Equal(t, blockType, b.BlockType())

	ps1.AssertNumberOfCalls(t, "Prepare", 1)
	rcs1.AssertNumberOfCalls(t, "Check", 1)
	rcs2.AssertNumberOfCalls(t, "Check", 0)
	ssm.AssertNumberOfCalls(t, "OnEntryPassed", 0)
	ssm.AssertNumberOfCalls(t, "OnEntryBlocked", 1)
	ssm.AssertNumberOfCalls(t, "OnCompleted", 0)
}
