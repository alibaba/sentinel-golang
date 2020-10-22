/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	datasourcev1 "github.com/alibaba/sentinel-golang/ext/datasource/k8s/api/v1"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CircuitBreakerRulesReconciler reconciles a CircuitBreakerRules object
type CircuitBreakerRulesReconciler struct {
	client.Client
	Logger          logr.Logger
	Scheme          *runtime.Scheme
	EffectiveCrName string
}

const (
	SlowRequestRatioStrategy string = "SlowRequestRatio"
	ErrorRatioStrategy       string = "ErrorRatio"
	ErrorCountStrategy       string = "ErrorCount"
)

// +kubebuilder:rbac:groups=datasource.sentinel.io,resources=circuitbreakerrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=datasource.sentinel.io,resources=circuitbreakerrules/status,verbs=get;update;patch

func (r *CircuitBreakerRulesReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Logger
	log.Info("receive CircuitBreakerRules", "namespace", req.NamespacedName.String())

	if req.Name != r.EffectiveCrName {
		log.V(int(logging.WarnLevel)).Info("ignore unregister cr.", "ns", req.Namespace, "crName", req.Name)
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, nil
	}

	cbRulesCR := &datasourcev1.CircuitBreakerRules{}
	if err := r.Get(ctx, req.NamespacedName, cbRulesCR); err != nil {
		log.Error(err, "Fail to get datasourcev1.CircuitBreakerRules.")
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, err
	}
	log.Info("Get datasourcev1.CircuitBreakerRules", "rules:", cbRulesCR.Spec.Rules)

	cbRules := r.assembleCircuitBreakerRules(cbRulesCR)
	_, err := circuitbreaker.LoadRules(cbRules)
	if err != nil {
		log.Error(err, "Fail to Load circuitbreaker.Rules")
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, err
	}
	return ctrl.Result{}, nil
}

func (r *CircuitBreakerRulesReconciler) assembleCircuitBreakerRules(rs *datasourcev1.CircuitBreakerRules) []*circuitbreaker.Rule {
	ret := make([]*circuitbreaker.Rule, 0, len(rs.Spec.Rules))

	for _, rule := range rs.Spec.Rules {
		cbRule := &circuitbreaker.Rule{
			Id:               rule.Id,
			Resource:         rule.Resource,
			RetryTimeoutMs:   uint32(rule.RetryTimeoutMs),
			MinRequestAmount: uint64(rule.MinRequestAmount),
			StatIntervalMs:   uint32(rule.StatIntervalMs),
			MaxAllowedRtMs:   uint64(rule.MaxAllowedRtMs),
		}
		switch rule.Strategy {
		case SlowRequestRatioStrategy:
			cbRule.Strategy = circuitbreaker.SlowRequestRatio
			cbRule.Threshold = float64(rule.Threshold) / 100
		case ErrorRatioStrategy:
			cbRule.Strategy = circuitbreaker.ErrorRatio
			cbRule.Threshold = float64(rule.Threshold) / 100
		case ErrorCountStrategy:
			cbRule.Strategy = circuitbreaker.ErrorCount
			cbRule.Threshold = float64(rule.Threshold)
		default:
			r.Logger.Error(errors.New("unsupported circuit breaker strategy"), "", "strategy", rule.Strategy)
			continue
		}

		ret = append(ret, cbRule)
	}
	return ret
}

func (r *CircuitBreakerRulesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datasourcev1.CircuitBreakerRules{}).
		Complete(r)
}
