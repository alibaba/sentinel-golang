package core

import (
	"github.com/sentinel-group/sentinel-golang/util"
	"sync"
)

// _defaultSlotChain is a default slot chain built by framework
// _defaultSlotChain is global unique chain
var _defaultSlotChain = buildDefaultSlotChain()
var logger = util.GetDefaultLogger()

// StatPrepareSlot is responsible for some preparation before statistic
// For example: init structure and so on
type StatPrepareSlot interface {
	// Prepare function do some initialization
	// Such as: init statistic structure、node and etc
	// The result of preparing would store in EntryContext
	// All StatPrepareSlots execute in sequence
	// Prepare function should not throw panic.
	Prepare(ctx *EntryContext)
}

// RuleCheckSlot is rule based checking strategy
// All checking rule must implement this interface.
type RuleCheckSlot interface {
	// Check function do some validation
	// It can break off the slot pipeline
	// Each RuleCheckSlot will return check result
	// The upper logic will control pipeline according to SlotResult.
	Check(ctx *EntryContext) *RuleCheckResult
}

// StatSlot is responsible for counting all custom biz metrics.
type StatSlot interface {
	// OnEntryPass function will be invoked when StatPrepareSlots and RuleCheckSlots execute pass
	// StatSlots will do some statistic logic, such as QPS、log、etc
	OnEntryPassed(ctx *EntryContext)
	// OnEntryBlocked function will be invoked when StatPrepareSlots and RuleCheckSlots execute fail
	// StatSlots will do some statistic logic, such as QPS、log、etc
	// blockEvent is a enum{RuleBasedCheckBlockedEvent} indicate the block event
	OnEntryBlocked(ctx *EntryContext, blockEvent RuleBasedCheckBlockedEvent)
	// onComplete function will be invoked when chain exits.
	OnCompleted(ctx *EntryContext)
}

// SlotChain hold all system Slots and customized slot.
// SlotChain support plug-in slots developed by developer.
type SlotChain struct {
	statPres   []StatPrepareSlot
	ruleChecks []RuleCheckSlot
	stats      []StatSlot
	// EntryContext Pool, used for reuse EntryContext object
	pool sync.Pool
}

func NewSlotChain() *SlotChain {
	return &SlotChain{
		statPres:   make([]StatPrepareSlot, 0, 5),
		ruleChecks: make([]RuleCheckSlot, 0, 5),
		stats:      make([]StatSlot, 0, 5),
		pool: sync.Pool{
			New: func() interface{} {
				return NewEntryContext()
			},
		},
	}
}

func buildDefaultSlotChain() *SlotChain {
	sc := NewSlotChain()
	// insert slots
	return sc
}

func GetDefaultSlotChain() *SlotChain {
	return _defaultSlotChain
}

// Get a EntryContext from EntryContext pool, if pool doesn't have enough EntryContext then new one.
func (sc *SlotChain) GetContext() *EntryContext {
	ctx := sc.pool.Get().(*EntryContext)
	defer sc.pool.Put(ctx)
	return ctx
}

func (sc *SlotChain) addStatPrepareSlotFirst(s StatPrepareSlot) {
	ns := make([]StatPrepareSlot, 0, len(sc.statPres)+1)
	// add to first
	ns = append(ns, s)
	sc.statPres = append(ns, sc.statPres...)
}

func (sc *SlotChain) addStatPrepareSlotLast(s StatPrepareSlot) {
	sc.statPres = append(sc.statPres, s)
}

func (sc *SlotChain) addRuleCheckSlotFirst(s RuleCheckSlot) {
	ns := make([]RuleCheckSlot, 0, len(sc.ruleChecks)+1)
	ns = append(ns, s)
	sc.ruleChecks = append(ns, sc.ruleChecks...)
}

func (sc *SlotChain) addRuleCheckSlotLast(s RuleCheckSlot) {
	sc.ruleChecks = append(sc.ruleChecks, s)
}

func (sc *SlotChain) addStatSlotFirst(s StatSlot) {
	ns := make([]StatSlot, 0, len(sc.stats)+1)
	ns = append(ns, s)
	sc.stats = append(ns, sc.stats...)
}

func (sc *SlotChain) addStatSlotLast(s StatSlot) {
	sc.stats = append(sc.stats, s)
}

// The entrance of Slot Chain
func (sc *SlotChain) entry(ctx *EntryContext) {
	logger.Debugln("entry slot chain")
	startTime := util.CurrentTimeMillis()

	// This should not happen, unless there are errors existing in Sentinel internal.
	// defer to handle it
	defer func() {
		if err := recover(); err != nil {
			logger.Panicf("unknown panic in SlotChain, err: %+v \n", err)
			return
		}
	}()

	// execute prepare slot
	logger.Debugln("execute prepare slot")
	sps := sc.statPres
	if len(sps) > 0 {
		for _, s := range sps {
			s.Prepare(ctx)
		}
	}

	// execute rule based checking slot
	logger.Debugln("execute rule based checking slot")
	rcs := sc.ruleChecks
	ruleCheckRet := NewSlotResultPass()
	if len(rcs) > 0 {
		for _, s := range rcs {
			sr := s.Check(ctx)
			// check slot result
			if sr.Status == ResultStatusPass {
				continue
			}
			// block or other logic
			logger.Warningf("[%v] check fail, reason is %s \n", s, sr.toString())
			ruleCheckRet.Status = sr.Status
			ruleCheckRet.BlockedMsg = sr.BlockedMsg
			ruleCheckRet.BlockedEvent = sr.BlockedEvent
			break
		}
	}

	// execute statistic slot
	logger.Debugln("execute statistic slot")
	ss := sc.stats
	if len(ss) > 0 {
		for _, s := range ss {
			// indicate the result of rule based checking slot.
			if ruleCheckRet.Status == ResultStatusPass {
				s.OnEntryPassed(ctx)
			} else {
				s.OnEntryBlocked(ctx, ruleCheckRet.BlockedEvent)
			}
		}
	}
	logAccess(ctx, startTime)
}

func (sc *SlotChain) exit(ctx *EntryContext) {
	logger.Debugln("exit slot chain")
	startTime := util.CurrentTimeMillis()
	for _, s := range sc.stats {
		s.OnCompleted(ctx)
	}
	logAccess(ctx, startTime)
}

func logAccess(ctx *EntryContext, startTime uint64) {
	logger.Debugf("start: %d, end: %d, EntryContext info %v \n", startTime, util.CurrentTimeMillis(), *ctx)
}
