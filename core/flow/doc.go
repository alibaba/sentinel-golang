// Package flow implements the flow shaping control.
//
// flow module supports two statistic metric: QPS and Concurrency.
//
// The TrafficShapingController consists of two part: TrafficShapingCalculator and TrafficShapingChecker
//
//  1. TrafficShapingCalculator calculates the actual traffic shaping token threshold. Currently, Sentinel supports two token calculate strategy: Direct and WarmUp.
//  2. TrafficShapingChecker performs checking logic according to current metrics and the traffic shaping strategy, then yield the token result. Currently, Sentinel supports two control behavior: Reject and Throttling.
//
// Besides, Sentinel supports customized TrafficShapingCalculator and TrafficShapingChecker. User could call function SetTrafficShapingGenerator to register customized TrafficShapingController and call function RemoveTrafficShapingGenerator to unregister TrafficShapingController.
// There are a few notes users need to be aware of:
//
//  1. The function both SetTrafficShapingGenerator and RemoveTrafficShapingGenerator is not thread safe.
//  2. Users can not override the Sentinel supported TrafficShapingController.
//
package flow
