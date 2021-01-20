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
	"net/http"

	"github.com/alibaba/sentinel-golang/core/isolation"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/go-logr/logr"
	k8sApiError "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	datasourcev1 "github.com/alibaba/sentinel-golang/pkg/datasource/k8s/api/v1alpha1"
)

// IsolationRulesReconciler reconciles a IsolationRules object
type IsolationRulesReconciler struct {
	client.Client
	Logger         logr.Logger
	Scheme         *runtime.Scheme
	Namespace      string
	ExpectedCrName string
}

// +kubebuilder:rbac:groups=datasource.sentinel.io,resources=isolationrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=datasource.sentinel.io,resources=isolationrules/status,verbs=get;update;patch

func (r *IsolationRulesReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
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

	isolationRulesCR := &datasourcev1.IsolationRules{}
	if err := r.Get(ctx, req.NamespacedName, isolationRulesCR); err != nil {
		k8sApiErr, ok := err.(*k8sApiError.StatusError)
		if !ok {
			log.Error(err, "Fail to get datasourcev1.IsolationRules.")
			return ctrl.Result{
				Requeue:      false,
				RequeueAfter: 0,
			}, nil
		}
		if k8sApiErr.Status().Code != http.StatusNotFound {
			log.Error(err, "Fail to get datasourcev1.IsolationRules.")
			return ctrl.Result{
				Requeue:      false,
				RequeueAfter: 0,
			}, nil
		}
		log.Info("datasourcev1.IsolationRules had been deleted.")
		isolationRulesCR = nil
	}

	var isolationRules []*isolation.Rule
	if isolationRulesCR != nil {
		log.Info("Get datasourcev1.IsolationRules", "rules:", isolationRulesCR.Spec.Rules)
		isolationRules = make([]*isolation.Rule, 0, len(isolationRulesCR.Spec.Rules))
		for _, r := range isolationRulesCR.Spec.Rules {
			isolationRules = append(isolationRules, &isolation.Rule{
				ID:         r.ID,
				Resource:   r.Resource,
				MetricType: isolation.Concurrency,
				Threshold:  uint32(r.Threshold),
			})
		}
	}

	_, err := isolation.LoadRules(isolationRules)
	if err != nil {
		log.Error(err, "Fail to Load isolation.Rules")
		return ctrl.Result{
			Requeue:      false,
			RequeueAfter: 0,
		}, err
	}
	return ctrl.Result{}, nil
}

func (r *IsolationRulesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&datasourcev1.IsolationRules{}).
		Complete(r)
}
