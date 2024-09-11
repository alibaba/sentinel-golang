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

package outlier

import (
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
)

// Rule encompasses the fields of outlier ejection rule.
type Rule struct {
	*circuitbreaker.Rule

	// Whether to enable active detection mode for recovery
	EnableActiveRecovery bool

	// An upper limit on the percentage of service nodes to be removed, which
	// defines the maximum percentage of nodes allowed to be excluded from
	// the service's load balancing pool.
	MaxEjectionPercent float64

	// The initial value of the time interval (in ms) to resume detection.
	// Enabling active detection mode will disable passive detection.
	RecoveryInterval uint32

	// Maximum number of recovery attempts allowed during recovery detection
	MaxRecoveryAttempts uint32
}
