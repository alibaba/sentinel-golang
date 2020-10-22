package k8s

import (
	"strings"

	"github.com/alibaba/sentinel-golang/util"

	datasourcev1 "github.com/alibaba/sentinel-golang/ext/datasource/k8s/api/v1"
	"github.com/alibaba/sentinel-golang/ext/datasource/k8s/controllers"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = datasourcev1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

type CrdType int

const (
	FlowRulesCrd           CrdType = 1
	CircuitBreakerRulesCrd CrdType = 2
	HotspotRulesCrd        CrdType = 3
	SystemRulesCrd         CrdType = 4
)

func (c CrdType) String() string {
	switch c {
	case FlowRulesCrd:
		return "FlowRulesCrd"
	case CircuitBreakerRulesCrd:
		return "CircuitBreakerRulesCrd"
	case HotspotRulesCrd:
		return "HotspotRulesCrd"
	case SystemRulesCrd:
		return "SystemRulesCrd"
	default:
		return "Undefined"
	}
}

type K8sDatasource struct {
	crdManager  ctrl.Manager
	controllers map[CrdType]reconcile.Reconciler
	stopChan    chan struct{}
}

func NewK8sDatasource() (*K8sDatasource, error) {
	ctrl.SetLogger(&k8sLogger{
		l:             logging.GetGlobalLogger(),
		level:         logging.GetGlobalLoggerLevel(),
		names:         make([]string, 0),
		keysAndValues: make([]interface{}, 0),
	})
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		// disable metric server
		MetricsBindAddress:     "0",
		HealthProbeBindAddress: "0",
		LeaderElection:         false,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return nil, err
	}
	k := &K8sDatasource{
		crdManager:  mgr,
		controllers: make(map[CrdType]reconcile.Reconciler, 0),
		stopChan:    make(chan struct{}),
	}
	return k, nil
}

func (k *K8sDatasource) RegisterController(crd CrdType, crName string) error {
	if len(strings.TrimSpace(crName)) == 0 {
		return errors.New("empty crName")
	}

	_, exist := k.controllers[crd]
	if exist {
		return errors.Errorf("Duplicated register crd for: %s", crd.String())
	}

	switch crd {
	case FlowRulesCrd:
		controller := &controllers.FlowRulesReconciler{
			Client:          k.crdManager.GetClient(),
			Logger:          ctrl.Log.WithName("controllers").WithName("FlowRules"),
			Scheme:          k.crdManager.GetScheme(),
			EffectiveCrName: crName,
		}
		err := controller.SetupWithManager(k.crdManager)
		if err != nil {
			return err
		}
		k.controllers[FlowRulesCrd] = controller
		setupLog.Info("succeed to register FlowRulesCrd Controller.")
		return nil
	case CircuitBreakerRulesCrd:
		controller := &controllers.CircuitBreakerRulesReconciler{
			Client:          k.crdManager.GetClient(),
			Logger:          ctrl.Log.WithName("controllers").WithName("CircuitBreakerRules"),
			Scheme:          k.crdManager.GetScheme(),
			EffectiveCrName: crName,
		}
		err := controller.SetupWithManager(k.crdManager)
		if err != nil {
			return err
		}
		k.controllers[CircuitBreakerRulesCrd] = controller
		setupLog.Info("succeed to register CircuitBreakerRulesCrd Controller.")
		return nil
	case HotspotRulesCrd:
		controller := &controllers.HotspotRulesReconciler{
			Client:          k.crdManager.GetClient(),
			Logger:          ctrl.Log.WithName("controllers").WithName("HotspotRules"),
			Scheme:          k.crdManager.GetScheme(),
			EffectiveCrName: crName,
		}
		err := controller.SetupWithManager(k.crdManager)
		if err != nil {
			return err
		}
		k.controllers[HotspotRulesCrd] = controller
		setupLog.Info("succeed to register HotspotRulesCrd Controller.")
		return nil
	case SystemRulesCrd:
		controller := &controllers.SystemRulesReconciler{
			Client:          k.crdManager.GetClient(),
			Logger:          ctrl.Log.WithName("controllers").WithName("SystemRules"),
			Scheme:          k.crdManager.GetScheme(),
			EffectiveCrName: crName,
		}
		err := controller.SetupWithManager(k.crdManager)
		if err != nil {
			return err
		}
		k.controllers[SystemRulesCrd] = controller
		setupLog.Info("succeed to register SystemRulesCrd Controller.")
		return nil
	default:
		return errors.Errorf("unsupported CrdType: %d", int(crd))
	}
}

func (k *K8sDatasource) Close() error {
	k.stopChan <- struct{}{}
	return nil
}

func (k *K8sDatasource) Run() error {
	// +kubebuilder:scaffold:builder
	go util.RunWithRecover(func() {
		setupLog.Info("starting manager")
		if err := k.crdManager.Start(k.stopChan); err != nil {
			setupLog.Error(err, "problem running manager")
		}
		setupLog.Info("k8s datasource exited")
	})
	return nil
}
