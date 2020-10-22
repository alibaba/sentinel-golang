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

	"github.com/alibaba/sentinel-golang/core/hotspot"
	datasourcev1 "github.com/alibaba/sentinel-golang/ext/datasource/k8s/api/v1"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HotspotRulesReconciler reconciles a HotspotRules object
type HotspotRulesReconciler struct {
	client.Client
	Logger          logr.Logger
	Scheme          *runtime.Scheme
	EffectiveCrName string
}

// +kubebuilder:rbac:groups=datasource.sentinel.io,resources=hotspotrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=datasource.sentinel.io,resources=hotspotrules/status,verbs=get;update;patch

func (r *HotspotRulesReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Logger
	log.Info("receive HotspotRules", "namespace", req.NamespacedName.String())

	if req.Name != r.EffectiveCrName {
		log.V(int(logging.WarnLevel)).Info("ignore unregister cr.", "ns", req.Namespace, "crName", req.Name)
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, nil
	}

	hotspotRulesCR := &datasourcev1.HotspotRules{}
	if err := r.Get(ctx, req.NamespacedName, hotspotRulesCR); err != nil {
		log.Error(err, "Fail to get datasourcev1.HotspotRules")
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, err
	}
	log.Info("Receive datasourcev1.HotspotRules", "rules:", hotspotRulesCR.Spec.Rules)

	hotspotRules := r.assembleHotspotRules(hotspotRulesCR)
	_, err := hotspot.LoadRules(hotspotRules)
	if err != nil {
		log.Error(err, "Fail to Load hotspot.Rules")
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, err
	}

	return ctrl.Result{}, nil
}

func (r *HotspotRulesReconciler) assembleHotspotRules(rs *datasourcev1.HotspotRules) []*hotspot.Rule {
	ret := make([]*hotspot.Rule, 0, len(rs.Spec.Rules))
	log := r.Logger
	for _, rule := range rs.Spec.Rules {
		hotspotRule := &hotspot.Rule{
			ID:                rule.Id,
			Resource:          rule.Resource,
			MetricType:        0,
			ControlBehavior:   0,
			ParamIndex:        int(rule.ParamIndex),
			Threshold:         float64(rule.Threshold),
			MaxQueueingTimeMs: int64(rule.MaxQueueingTimeMs),
			BurstCount:        int64(rule.BurstCount),
			DurationInSec:     int64(rule.DurationInSec),
			ParamsMaxCapacity: int64(rule.ParamsMaxCapacity),
			SpecificItems:     make([]hotspot.SpecificValue, 0, len(rule.SpecificItems)),
		}
		switch rule.MetricType {
		case ConcurrencyMetricType:
			hotspotRule.MetricType = hotspot.Concurrency
		case QPSMetricType:
			hotspotRule.MetricType = hotspot.QPS
		default:
			log.Error(errors.New("unsupported MetricType for hotspot.Rule"), "", "metricType", rule.MetricType)
			continue
		}

		switch rule.ControlBehavior {
		case RejectControlBehavior:
			hotspotRule.ControlBehavior = hotspot.Reject
		case ThrottlingControlBehavior:
			hotspotRule.ControlBehavior = hotspot.Throttling
		default:
			log.Error(errors.New("unsupported ControlBehavior for hotspot.Rule"), "", "controlBehavior", rule.ControlBehavior)
			continue
		}

		for _, specificItem := range rule.SpecificItems {
			hotspotSpecificValue := hotspot.SpecificValue{
				ValStr:    specificItem.ValStr,
				Threshold: int64(specificItem.Threshold),
			}
			switch specificItem.ValKind {
			case "KindInt":
				hotspotSpecificValue.ValKind = hotspot.KindInt
			case "KindString":
				hotspotSpecificValue.ValKind = hotspot.KindString
			case "KindBool":
				hotspotSpecificValue.ValKind = hotspot.KindBool
			case "KindFloat64":
				hotspotSpecificValue.ValKind = hotspot.KindFloat64
			default:
				log.Error(errors.New("unsupported hotspot.SpecificValue.ValKind"), "", "valKind", specificItem.ValKind)
				continue
			}
			hotspotRule.SpecificItems = append(hotspotRule.SpecificItems, hotspotSpecificValue)
		}
		ret = append(ret, hotspotRule)
	}
	return ret
}

func (r *HotspotRulesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datasourcev1.HotspotRules{}).
		Complete(r)
}
