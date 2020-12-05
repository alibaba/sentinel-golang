// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package base

import (
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type StatPrepareSlotMock1 struct {
	name  string
	order uint32
}

func (spl *StatPrepareSlotMock1) Name() string {
	return spl.name
}

func (spl *StatPrepareSlotMock1) Order() uint32 {
	return spl.order
}

func (spl *StatPrepareSlotMock1) Prepare(ctx *EntryContext) {
	return
}

func TestSlotChain_AddStatPrepareSlot(t *testing.T) {
	sc := NewSlotChain()
	for i := 9; i >= 0; i-- {
		sc.AddStatPrepareSlot(&StatPrepareSlotMock1{
			name:  "mock2" + strconv.Itoa(i),
			order: uint32(20 + i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.AddStatPrepareSlot(&StatPrepareSlotMock1{
			name:  "mock1" + strconv.Itoa(i),
			order: uint32(10 + i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.AddStatPrepareSlot(&StatPrepareSlotMock1{
			name:  "mock3" + strconv.Itoa(i),
			order: uint32(30 + i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.AddStatPrepareSlot(&StatPrepareSlotMock1{
			name:  "mock" + strconv.Itoa(i),
			order: uint32(i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.AddStatPrepareSlot(&StatPrepareSlotMock1{
			name:  "mock4" + strconv.Itoa(i),
			order: uint32(40 + i),
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
		assert.True(t, reflect.DeepEqual(n, spsm.name))
	}
}

type RuleCheckSlotMock1 struct {
	name  string
	order uint32
}

func (rcs *RuleCheckSlotMock1) Name() string {
	return rcs.name
}

func (rcs *RuleCheckSlotMock1) Order() uint32 {
	return rcs.order
}

func (rcs *RuleCheckSlotMock1) Check(ctx *EntryContext) *TokenResult {
	return nil
}
func TestSlotChain_AddRuleCheckSlot(t *testing.T) {
	sc := NewSlotChain()
	for i := 9; i >= 0; i-- {
		sc.AddRuleCheckSlot(&RuleCheckSlotMock1{
			name:  "mock2" + strconv.Itoa(i),
			order: uint32(20 + i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.AddRuleCheckSlot(&RuleCheckSlotMock1{
			name:  "mock1" + strconv.Itoa(i),
			order: uint32(10 + i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.AddRuleCheckSlot(&RuleCheckSlotMock1{
			name:  "mock3" + strconv.Itoa(i),
			order: uint32(30 + i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.AddRuleCheckSlot(&RuleCheckSlotMock1{
			name:  "mock" + strconv.Itoa(i),
			order: uint32(i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.AddRuleCheckSlot(&RuleCheckSlotMock1{
			name:  "mock4" + strconv.Itoa(i),
			order: uint32(40 + i),
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
		assert.True(t, reflect.DeepEqual(n, spsm.name))
	}
}

type StatSlotMock1 struct {
	name  string
	order uint32
}

func (ss *StatSlotMock1) Name() string {
	return ss.name
}

func (ss *StatSlotMock1) Order() uint32 {
	return ss.order
}

func (ss *StatSlotMock1) OnEntryPassed(ctx *EntryContext) {
	return
}
func (ss *StatSlotMock1) OnEntryBlocked(ctx *EntryContext, blockError *BlockError) {
	return
}
func (ss *StatSlotMock1) OnCompleted(ctx *EntryContext) {
	return
}
func TestSlotChain_AddStatSlot(t *testing.T) {
	sc := NewSlotChain()
	for i := 9; i >= 0; i-- {
		sc.AddStatSlot(&StatSlotMock1{
			name:  "mock2" + strconv.Itoa(i),
			order: uint32(20 + i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.AddStatSlot(&StatSlotMock1{
			name:  "mock1" + strconv.Itoa(i),
			order: uint32(10 + i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.AddStatSlot(&StatSlotMock1{
			name:  "mock3" + strconv.Itoa(i),
			order: uint32(30 + i),
		})
	}
	for i := 9; i >= 0; i-- {
		sc.AddStatSlot(&StatSlotMock1{
			name:  "mock" + strconv.Itoa(i),
			order: uint32(i),
		})
	}
	for i := 0; i < 10; i++ {
		sc.AddStatSlot(&StatSlotMock1{
			name:  "mock4" + strconv.Itoa(i),
			order: uint32(40 + i),
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
		assert.True(t, reflect.DeepEqual(n, spsm.name))
	}
}

type prepareSlotMock struct {
	mock.Mock
}

func (m *prepareSlotMock) Name() string {
	return "mock-sentinel-prepare-slot-1"
}

func (m *prepareSlotMock) Order() uint32 {
	return 0
}

func (m *prepareSlotMock) Prepare(ctx *EntryContext) {
	m.Called(ctx)
	return
}

type mockRuleCheckSlot1 struct {
	mock.Mock
}

func (m *mockRuleCheckSlot1) Name() string {
	return "mock-sentinel-check-slot--1"
}

func (m *mockRuleCheckSlot1) Order() uint32 {
	return 0
}

func (m *mockRuleCheckSlot1) Check(ctx *EntryContext) *TokenResult {
	arg := m.Called(ctx)
	return arg.Get(0).(*TokenResult)
}

type mockRuleCheckSlot2 struct {
	mock.Mock
}

func (m *mockRuleCheckSlot2) Name() string {
	return "mock-sentinel-rule-check-slot-2"
}

func (m *mockRuleCheckSlot2) Order() uint32 {
	return 0
}

func (m *mockRuleCheckSlot2) Check(ctx *EntryContext) *TokenResult {
	arg := m.Called(ctx)
	return arg.Get(0).(*TokenResult)
}

type statisticSlotMock struct {
	mock.Mock
}

func (m *statisticSlotMock) Name() string {
	return "mock-sentinel-statistic-slot"
}

func (m *statisticSlotMock) Order() uint32 {
	return 0
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
		BatchCount:  1,
		Flag:        0,
		Args:        nil,
		Attachments: nil,
	}

	ps1 := &prepareSlotMock{}
	rcs1 := &mockRuleCheckSlot1{}
	rcs2 := &mockRuleCheckSlot2{}
	ssm := &statisticSlotMock{}
	sc.AddStatPrepareSlot(ps1)
	sc.AddRuleCheckSlot(rcs1)
	sc.AddRuleCheckSlot(rcs2)
	sc.AddStatSlot(ssm)

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
		BatchCount:  1,
		Flag:        0,
		Args:        nil,
		Attachments: nil,
	}

	rbs := &prepareSlotMock{}
	fsm := &mockRuleCheckSlot1{}
	dsm := &mockRuleCheckSlot2{}
	ssm := &statisticSlotMock{}
	sc.AddStatPrepareSlot(rbs)
	sc.AddRuleCheckSlot(fsm)
	sc.AddRuleCheckSlot(dsm)
	sc.AddStatSlot(ssm)

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

func (m *badPrepareSlotMock) Name() string {
	return "bad-mock-sentinel-prepare-slot"
}

func (m *badPrepareSlotMock) Order() uint32 {
	return 0
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
		BatchCount:  1,
		Flag:        0,
		Args:        nil,
		Attachments: nil,
	}

	rbs := &badPrepareSlotMock{}
	fsm := &mockRuleCheckSlot1{}
	dsm := &mockRuleCheckSlot2{}
	ssm := &statisticSlotMock{}
	sc.AddStatPrepareSlot(rbs)
	sc.AddRuleCheckSlot(fsm)
	sc.AddRuleCheckSlot(dsm)
	sc.AddStatSlot(ssm)

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

func TestValidateStatPrepareSlotNaming(t *testing.T) {
	sc := NewSlotChain()
	sps1 := &StatPrepareSlotMock1{
		name: "sps1",
	}
	sps2 := &StatPrepareSlotMock1{
		name: "sps2",
	}
	sps3 := &StatPrepareSlotMock1{
		name: "sps3",
	}
	sps4 := &StatPrepareSlotMock1{
		name: "sps4",
	}
	sc.AddStatPrepareSlot(sps1)
	sc.AddStatPrepareSlot(sps2)
	sc.AddStatPrepareSlot(sps3)
	sc.AddStatPrepareSlot(sps4)

	sps5 := &StatPrepareSlotMock1{
		name: "sps5",
	}
	assert.True(t, ValidateStatPrepareSlotNaming(sc, sps5))
	sps6 := &StatPrepareSlotMock1{
		name: "sps1",
	}
	assert.True(t, !ValidateStatPrepareSlotNaming(sc, sps6))
}

func TestValidateRuleCheckSlotNaming(t *testing.T) {
	sc := NewSlotChain()
	rcs1 := &RuleCheckSlotMock1{
		name: "rcs1",
	}
	rcs2 := &RuleCheckSlotMock1{
		name: "rcs2",
	}
	rcs3 := &RuleCheckSlotMock1{
		name: "rcs3",
	}
	rcs4 := &RuleCheckSlotMock1{
		name: "rcs4",
	}
	sc.AddRuleCheckSlot(rcs1)
	sc.AddRuleCheckSlot(rcs2)
	sc.AddRuleCheckSlot(rcs3)
	sc.AddRuleCheckSlot(rcs4)

	rcs5 := &RuleCheckSlotMock1{
		name: "rcs5",
	}
	assert.True(t, ValidateRuleCheckSlotNaming(sc, rcs5))
	rcs6 := &RuleCheckSlotMock1{
		name: "rcs1",
	}
	assert.True(t, !ValidateRuleCheckSlotNaming(sc, rcs6))
}

func TestValidateStatSlotNaming(t *testing.T) {
	sc := NewSlotChain()
	ss1 := &StatSlotMock1{
		name: "ss1",
	}
	ss2 := &StatSlotMock1{
		name: "ss2",
	}
	ss3 := &StatSlotMock1{
		name: "ss3",
	}
	ss4 := &StatSlotMock1{
		name: "ss4",
	}
	sc.AddStatSlot(ss1)
	sc.AddStatSlot(ss2)
	sc.AddStatSlot(ss3)
	sc.AddStatSlot(ss4)

	ss5 := &StatSlotMock1{
		name: "ss5",
	}
	assert.True(t, ValidateStatSlotNaming(sc, ss5))
	ss6 := &StatSlotMock1{
		name: "ss1",
	}
	assert.True(t, !ValidateStatSlotNaming(sc, ss6))
}
