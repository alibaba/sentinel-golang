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
	"fmt"
	"net/http"
	"strconv"

	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	k8sApiError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datasourcev1 "github.com/alibaba/sentinel-golang/pkg/datasource/k8s/api/v1alpha1"
)

const (
	ConcurrencyMetricType string = "Concurrency"
	QPSMetricType         string = "QPS"
)

// HotspotRulesReconciler reconciles a HotspotRules object
type HotspotRulesReconciler struct {
	client.Client
	Logger         logr.Logger
	Scheme         *runtime.Scheme
	Namespace      string
	ExpectedCrName string
}

// +kubebuilder:rbac:groups=datasource.sentinel.io,resources=hotspotrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=datasource.sentinel.io,resources=hotspotrules/status,verbs=get;update;patch
func (r *HotspotRulesReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Logger.WithValues("expectedNamespace", r.Namespace, "expectedCrName", r.ExpectedCrName, "req", req.String())

	if req.Namespace != r.Namespace {
		log.V(int(logging.DebugLevel)).Info("ignore unmatched namespace")
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, nil
	}

	if req.Name != r.ExpectedCrName {
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, nil
	}

	hotspotRulesCR := &datasourcev1.HotspotRules{}
	if err := r.Get(ctx, req.NamespacedName, hotspotRulesCR); err != nil {
		k8sApiErr, ok := err.(*k8sApiError.StatusError)
		if !ok {
			log.Error(err, "Fail to get datasourcev1.HotspotRules.")
			return ctrl.Result{
				Requeue:      false,
				RequeueAfter: 0,
			}, nil
		}
		if k8sApiErr.Status().Code != http.StatusNotFound {
			log.Error(err, "Fail to get datasourcev1.HotspotRules.")
			return ctrl.Result{
				Requeue:      false,
				RequeueAfter: 0,
			}, nil
		}
		log.Info("datasourcev1.CircuitBreakerRules had been deleted.")
		hotspotRulesCR = nil
	}

	var hotspotRules []*hotspot.Rule
	if hotspotRulesCR != nil {
		log.Info("Receive datasourcev1.HotspotRules", "rules:", hotspotRulesCR.Spec.Rules)
		hotspotRules = r.assembleHotspotRules(hotspotRulesCR)
	}

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
			Threshold:         rule.Threshold,
			MaxQueueingTimeMs: rule.MaxQueueingTimeMs,
			BurstCount:        rule.BurstCount,
			DurationInSec:     rule.DurationInSec,
			ParamsMaxCapacity: rule.ParamsMaxCapacity,
			SpecificItems:     parseSpecificItems(rule.SpecificItems),
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
		case "":
			hotspotRule.ControlBehavior = hotspot.Reject
		case RejectControlBehavior:
			hotspotRule.ControlBehavior = hotspot.Reject
		case ThrottlingControlBehavior:
			hotspotRule.ControlBehavior = hotspot.Throttling
		default:
			log.Error(errors.New("unsupported ControlBehavior for hotspot.Rule"), "", "controlBehavior", rule.ControlBehavior)
			continue
		}
		ret = append(ret, hotspotRule)
	}
	return ret
}

// arseSpecificItems parses the SpecificValue as real value.
func parseSpecificItems(source []datasourcev1.SpecificValue) map[interface{}]int64 {
	ret := make(map[interface{}]int64)
	if len(source) == 0 {
		return ret
	}
	for _, item := range source {
		switch item.ValKind {
		case "KindInt":
			realVal, err := strconv.Atoi(item.ValStr)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for int specific item", "itemValKind", item.ValKind, "itemValStr", item.ValStr)
				continue
			}
			ret[realVal] = item.Threshold
		case "KindString":
			ret[item.ValStr] = item.Threshold
		case "KindBool":
			realVal, err := strconv.ParseBool(item.ValStr)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for bool specific item", "itemValStr", item.ValStr)
				continue
			}
			ret[realVal] = item.Threshold
		case "KindFloat64":
			realVal, err := strconv.ParseFloat(item.ValStr, 64)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for float specific item", "itemValStr", item.ValStr)
				continue
			}
			realVal, err = strconv.ParseFloat(fmt.Sprintf("%.5f", realVal), 64)
			if err != nil {
				logging.Error(errors.Wrap(err, "parseSpecificItems error"), "Failed to parse value for float specific item", "itemValStr", item.ValStr)
				continue
			}
			ret[realVal] = item.Threshold
		default:
			logging.Error(errors.New("unsupported kind for specific item"), "", item.ValKind)
		}
	}
	return ret
}

func (r *HotspotRulesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datasourcev1.HotspotRules{}).
		Complete(r)
}
