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

package controllers

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.snapp.ir/snappcloud/consul-gslb-driver/pkg/connection"
	"gitlab.snapp.ir/snappcloud/consul-gslb-driver/pkg/rpc"
	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	gslbv1alpha1 "github.com/snapp-cab/gslb-controller/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

const (
	// Default timeout of short GSLBI calls like GetPluginInfo
	gslbiTimeout = time.Second
	gslbiAddress = "/Users/my/gitlab/consul-gslb-driver/socket" // Address of the GSLBI driver socket.
	timeout      = 15 * time.Second                             // Timeout for waiting for creating or deleting the gslb."
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = gslbv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = gslbv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	gslbiConn, err := connection.Connect(gslbiAddress, []grpc.DialOption{}, connection.OnConnectionLoss(connection.ExitOnConnectionLoss()))
	Expect(err).NotTo(HaveOccurred())

	err = rpc.ProbeForever(gslbiConn, timeout)
	Expect(err).NotTo(HaveOccurred())

	// Find driver name.
	_, err = rpc.GetDriverName(context.TODO(), gslbiConn)
	Expect(err).NotTo(HaveOccurred())
	NewClient(gslbiConn)

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&GslbReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = (&GslbContentReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	// err = (&gslbv1alpha1.Gslb{}).SetupWebhookWithManager(k8sManager)
	// Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func errString(err error) string {
	if err == nil {
		return ""
	}
	errStr := err.Error()
	if len(errStr) == 0 {
		panic("invalid error")
	}
	return errStr
}

func TestSth(t *testing.T) {

	tests := []struct {
		name    string
		hi      string
		gslb    *gslbv1alpha1.Gslb
		wantErr string
	}{
		{
			name: "all good",
			gslb: &gslbv1alpha1.Gslb{
				Spec: gslbv1alpha1.GslbSpec{
					ServiceName: "good",
				},
			},
		},
		{
			name: "bad gslb",
			gslb: &gslbv1alpha1.Gslb{
				Spec: gslbv1alpha1.GslbSpec{
					ServiceName: "learn",
				},
			},
			wantErr: "must not equal learn",
		},
		{
			name: "wrong hi",
			gslb: &gslbv1alpha1.Gslb{
				Spec: gslbv1alpha1.GslbSpec{
					ServiceName: "learn",
				},
			},
			wantErr: "must not equal learn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := validate(tt.gslb)
			if errString(err) != tt.wantErr {
				t.Fatalf("got: %v, want: %s", err, tt.wantErr)
			}
		})
	}
}

func validate(gslb *gslbv1alpha1.Gslb) error {
	if gslb.Spec.ServiceName == "learn" {
		return fmt.Errorf("must not equal learn")
	}
	return nil
}
