/*
Copyright 2021.

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

package main

import (
	"context"
	"flag"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"github.com/snapp-incubator/consul-gslb-driver/pkg/connection"
	"github.com/snapp-incubator/consul-gslb-driver/pkg/rpc"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	gslbv1alpha1 "github.com/snapp-cab/gslb-controller/api/v1alpha1"
	"github.com/snapp-cab/gslb-controller/controllers"
	//+kubebuilder:scaffold:imports
)

const (
	// Default timeout of short GSLBI calls like GetPluginInfo
	gslbiTimeout = time.Second
	gslbiAddress = "/var/run/gslbi/gslbi.sock" // Address of the GSLBI driver socket.
	// timeout      = "15*time.Second"                             // Timeout for waiting for creating or deleting the gslb."
)

var (
	scheme     = runtime.NewScheme()
	setupLog   = ctrl.Log.WithName("setup")
	syncPeriod = 4 * time.Hour
	timeout    = flag.Duration("timeout", 15*time.Second, "Timeout for waiting for creating or deleting the gslb.")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(gslbv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	klog.InitFlags(nil)
	// Connect to GSLBI.
	gslbiConn, err := connection.Connect(gslbiAddress, []grpc.DialOption{}, connection.OnConnectionLoss(connection.ExitOnConnectionLoss()))
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}
	err = rpc.ProbeForever(gslbiConn, *timeout)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}

	// Find driver name.
	ctx, cancel := context.WithTimeout(context.Background(), gslbiTimeout)
	defer cancel()
	gslbiAttacher, err := rpc.GetDriverName(ctx, gslbiConn)
	if err != nil {
		klog.Error(err.Error())
		os.Exit(1)
	}
	klog.V(2).Infof("gslbi driver name: %q", gslbiAttacher)

	controllers.NewClient(gslbiConn)

	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "0865a5bd.snappcloud.io",
		SyncPeriod:             &syncPeriod,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.GslbReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Gslb")
		os.Exit(1)
	}
	if err = (&controllers.GslbContentReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GslbContent")
		os.Exit(1)
	}
	if err = (&gslbv1alpha1.Gslb{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Gslb")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
