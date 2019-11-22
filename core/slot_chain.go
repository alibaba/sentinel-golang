package core

import (
	"log"
	"sync"
)

// _defaultSlotChain is a default slot chain built by framework
// _defaultSlotChain is global unique chain
var _defaultSlotChain = buildDefaultSlotChain()

// StatPrepareSlot is responsible for some preparation before statistic
// For example: init structure and so on
type StatPrepareSlot interface {
	// Prepare function do some initialization
	// Such as: init statistic structure、node and etc
	// The result of preparing would store in Context
	// All StatPrepareSlots execute in sequence
	// Prepare function should not throw panic.
	Prepare(ctx *Context)
}

// RuleCheckSlot is rule based checking strategy
// All checking rule must implement this interface.
type RuleCheckSlot interface {
	// Check function do some validation
	// It can break off the slot pipeline
	// Each RuleCheckSlot will return check result
	// The upper logic will control pipeline according to SlotResult.
	Check(ctx *Context) *RuleCheckResult
}

// StatSlot is responsible for counting all custom biz metrics.
type StatSlot interface {
	// OnEntryPass function will be invoked when StatPrepareSlots and RuleCheckSlots execute pass
	// StatSlots will do some statistic logic, such as QPS、log、etc
	OnEntryPassed(ctx *Context)
	// OnEntryBlocked function will be invoked when StatPrepareSlots and RuleCheckSlots execute fail
	// StatSlots will do some statistic logic, such as QPS、log、etc
	// blockEvent is a enum{RuleBasedCheckBlockedEvent} indicate the block event
	OnEntryBlocked(ctx *Context, blockEvent RuleBasedCheckBlockedEvent)
	// onComplete function will be invoked when chain exits.
	OnCompleted(ctx *Context)
}

// SlotChain hold all system Slots and customized slot.
// SlotChain support plug-in slots developed by developer.
type SlotChain struct {
	statPres   []StatPrepareSlot
	ruleChecks []RuleCheckSlot
	stats      []StatSlot
	// Context Pool, used for reuse Context object
	pool sync.Pool
}

func NewSlotChain() *SlotChain {
	return &SlotChain{
		statPres:   make([]StatPrepareSlot, 0, 5),
		ruleChecks: make([]RuleCheckSlot, 0, 5),
		stats:      make([]StatSlot, 0, 5),
		pool: sync.Pool{
			New: func() interface{} {
				return NewContext()
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

// Get a Context from Context pool, if pool doesn't have enough Context then new one.
func (sc *SlotChain) GetContext() *Context {
	ctx := sc.pool.Get().(*Context)
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
func (sc *SlotChain) entry(ctx *Context) {
	log.Println("entry slot chain")
	startTime := GetTimeMilli()

	// This should not happen, unless there are errors existing in Sentinel internal.
	// defer to handle it
	defer func() {
		if err := recover(); err != nil {
			log.Printf("unknown panic in SlotChain, err: %+v \n", err)
			return
		}
	}()

	// execute prepare slot
	log.Println("execute prepare slot")
	sps := sc.statPres
	if len(sps) > 0 {
		for _, s := range sps {
			s.Prepare(ctx)
		}
	}

	// execute rule based checking slot
	log.Println("execute rule based checking slot")
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
			log.Printf("[%v] check fail, reason is %s \n", s, sr.toString())
			ruleCheckRet.Status = sr.Status
			ruleCheckRet.BlockedMsg = sr.BlockedMsg
			ruleCheckRet.BlockedEvent = sr.BlockedEvent
			break
		}
	}

	// execute statistic slot
	log.Println("execute statistic slot")
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

func (sc *SlotChain) exit(ctx *Context) {
	log.Println("exit slot chain")
	startTime := GetTimeMilli()
	for _, s := range sc.stats {
		s.OnCompleted(ctx)
	}
	logAccess(ctx, startTime)
}

func logAccess(ctx *Context, startTime uint64) {
	log.Printf("start: %d, end: %d, Context info %v \n", startTime, GetTimeMilli(), *ctx)
}
