/*
Copyright 2022 Angelos Kolaitis.

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

package manager

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"go.uber.org/zap/zapcore"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/snapcore/snapd/client"
	snapdclient "github.com/snapcore/snapd/client"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	microk8sv1alpha1 "github.com/neoaggelos/microk8s-operator/api/v1alpha1"
	"github.com/neoaggelos/microk8s-operator/controllers/configuration"
	"github.com/neoaggelos/microk8s-operator/controllers/microk8snode"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(microk8sv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func restartService(ctx context.Context, snapClient *snapdclient.Client, service string) error {
	changeID, err := snapClient.Restart([]string{service}, client.RestartOptions{Reload: false})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for {
		if change, err := snapClient.Change(changeID); err != nil {
			return err
		} else if change.Status == "Done" {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for restart")
		case <-time.After(time.Second):
		}
	}
}

func Main() {
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
		TimeEncoder: zapcore.ISO8601TimeEncoder,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		setupLog.Info("NODE_NAME is not set. It must be set to the current node.")
		os.Exit(1)
	}

	snapData := os.Getenv("SNAP_DATA")
	if snapData == "" {
		setupLog.Info("SNAP_DATA is not set. It must be set to the SNAP_DATA directory of MicroK8s")
		os.Exit(1)
	}

	snapCommon := os.Getenv("SNAP_COMMON")
	if snapCommon == "" {
		setupLog.Info("SNAP_COMMON is not set. It must be set to the SNAP_COMMON directory of MicroK8s")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "bf1786fa.canonical.com",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	snapClient := snapdclient.New(&snapdclient.Config{
		Socket: os.Getenv("SNAP_SOCKET"),
	})

	if err = (&configuration.Reconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),

		Node: nodeName,

		RestartContainerd: func(ctx context.Context) error {
			return restartService(ctx, snapClient, "microk8s.daemon-containerd")
		},
		RestartKubelet: func(ctx context.Context) error {
			return restartService(ctx, snapClient, "microk8s.daemon-kubelite")
		},
		RestartKubeAPIServer: func(ctx context.Context) error {
			return restartService(ctx, snapClient, "microk8s.daemon-kubelite")
		},
		RefreshCertificates: func(ctx context.Context) error {
			if _, err := snapClient.SetConf("microk8s", map[string]interface{}{
				"operator-configure-hook": time.Now().String(),
			}); err != nil {
				return fmt.Errorf("failed to change snap config: %w", err)
			}
			return nil
		},
		CSRConfFile:           filepath.Join(snapData, "certs", "csr.conf.template"),
		RegistryCertsDir:      filepath.Join(snapData, "args", "certs.d"),
		ContainerdEnvFile:     filepath.Join(snapData, "args", "containerd-env"),
		KubeletArgsFile:       filepath.Join(snapData, "args", "kubelet"),
		KubeAPIServerArgsFile: filepath.Join(snapData, "args", "kube-apiserver"),

		AddonsDir: filepath.Join(snapCommon, "addons"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Configuration")
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

	nodeController := &microk8snode.Controller{
		Client:   mgr.GetClient(),
		Interval: time.Minute,
		Node:     nodeName,
		SnapInfo: func(ctx context.Context) (microk8snode.SnapInfo, error) {
			r, err := snapClient.List([]string{"microk8s"}, nil)
			if err != nil {
				return microk8snode.SnapInfo{}, err
			}
			if len(r) == 0 {
				return microk8snode.SnapInfo{}, fmt.Errorf("no microk8s snap found")
			}
			return microk8snode.SnapInfo{
				Revision:    r[0].Revision.String(),
				Channel:     r[0].Channel,
				Version:     r[0].Version,
				Confinement: r[0].Confinement,
			}, nil
		},
	}

	ctx, cancel := context.WithCancel(ctrl.SetupSignalHandler())
	defer cancel()
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		setupLog.Info("starting node controller")
		if err := nodeController.Run(ctx); err != nil {
			setupLog.Error(err, "problem running node controller")
			cancel()
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		setupLog.Info("starting manager")
		if err := mgr.Start(ctx); err != nil {
			setupLog.Error(err, "problem running manager")
			cancel()
		}
		wg.Done()
	}()

	wg.Wait()
}
