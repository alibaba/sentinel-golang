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

	"github.com/alibaba/sentinel-golang/core/flow"
	datasourcev1 "github.com/alibaba/sentinel-golang/ext/datasource/k8s/api/v1"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FlowRulesReconciler reconciles a FlowRules object
type FlowRulesReconciler struct {
	client.Client
	Logger          logr.Logger
	Scheme          *runtime.Scheme
	EffectiveCrName string
}

const (
	ConcurrencyMetricType string = "Concurrency"
	QPSMetricType         string = "QPS"

	CurrentResourceRelationStrategy    string = "CurrentResource"
	AssociatedResourceRelationStrategy string = "AssociatedResource"

	DirectTokenCalculateStrategy string = "Direct"
	WarmUpTokenCalculateStrategy string = "WarmUp"

	RejectControlBehavior     string = "Reject"
	ThrottlingControlBehavior string = "Throttling"
)

// +kubebuilder:rbac:groups=datasource.sentinel.io,resources=flowrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=datasource.sentinel.io,resources=flowrules/status,verbs=get;update;patch

func (r *FlowRulesReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Logger
	log.Info("receive FlowRules", "namespace", req.NamespacedName.String())

	if req.Name != r.EffectiveCrName {
		log.V(int(logging.WarnLevel)).Info("ignore unregister cr.", "ns", req.Namespace, "crName", req.Name)
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, nil
	}

	// your logic here
	flowRulesCR := &datasourcev1.FlowRules{}
	if err := r.Get(ctx, req.NamespacedName, flowRulesCR); err != nil {
		log.Error(err, "Fail to get datasourcev1.FlowRules.")
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, err
	}
	log.Info("Receive datasourcev1.FlowRules", "rules:", flowRulesCR.Spec.Rules)

	flowRules := r.assembleFlowRules(flowRulesCR)
	_, err := flow.LoadRules(flowRules)
	if err != nil {
		log.Error(err, "Fail to Load flow.Rules")
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, err
	}
	return ctrl.Result{}, nil
}

func (r *FlowRulesReconciler) assembleFlowRules(rs *datasourcev1.FlowRules) []*flow.Rule {
	ret := make([]*flow.Rule, 0, len(rs.Spec.Rules))
	log := r.Logger
	for _, rule := range rs.Spec.Rules {
		cbRule := &flow.Rule{
			ID:                     rule.Id,
			Resource:               rule.Resource,
			MetricType:             0,
			TokenCalculateStrategy: 0,
			ControlBehavior:        0,
			Count:                  float64(rule.Count),
			RelationStrategy:       0,
			RefResource:            rule.RefResource,
			MaxQueueingTimeMs:      uint32(rule.MaxQueueingTimeMs),
			WarmUpPeriodSec:        uint32(rule.WarmUpPeriodSec),
			WarmUpColdFactor:       uint32(rule.WarmUpColdFactor),
		}
		switch rule.MetricType {
		case ConcurrencyMetricType:
			cbRule.MetricType = flow.Concurrency
		case QPSMetricType:
			cbRule.MetricType = flow.QPS
		default:
			log.Error(errors.New("unsupported MetricType for flow.Rule"), "", "metricType", rule.MetricType)
			continue
		}

		switch rule.TokenCalculateStrategy {
		case DirectTokenCalculateStrategy:
			cbRule.TokenCalculateStrategy = flow.Direct
		case WarmUpTokenCalculateStrategy:
			cbRule.TokenCalculateStrategy = flow.WarmUp
		default:
			log.Error(errors.New("unsupported TokenCalculateStrategy for flow.Rule"), "", "TokenCalculateStrategy", rule.TokenCalculateStrategy)
			continue
		}

		switch rule.ControlBehavior {
		case RejectControlBehavior:
			cbRule.ControlBehavior = flow.Reject
		case ThrottlingControlBehavior:
			cbRule.ControlBehavior = flow.Throttling
		default:
			log.Error(errors.New("unsupported ControlBehavior for flow.Rule"), "", "controlBehavior", rule.ControlBehavior)
			continue
		}

		switch rule.RelationStrategy {
		case CurrentResourceRelationStrategy:
			cbRule.RelationStrategy = flow.CurrentResource
		case AssociatedResourceRelationStrategy:
			cbRule.RelationStrategy = flow.AssociatedResource
		default:
			log.Error(errors.New("unsupported RelationStrategy for flow.Rule"), "", "relationStrategy", rule.RelationStrategy)
			continue
		}

		ret = append(ret, cbRule)
	}
	return ret
}

func (r *FlowRulesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datasourcev1.FlowRules{}).
		Complete(r)
}
