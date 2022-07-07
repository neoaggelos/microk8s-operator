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

package controllers

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	microk8sv1alpha1 "github.com/neoaggelos/microk8s-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ConfigurationReconciler reconciles a Configuration object
type ConfigurationReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Node       string
	SnapData   string
	SnapCommon string
}

//+kubebuilder:rbac:groups=microk8s.canonical.com,resources=configurations;microk8snodes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=microk8s.canonical.com,resources=configurations/status;microk8snodes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=microk8s.canonical.com,resources=configurations/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Configuration object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *ConfigurationReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	if req.Name != "default" && req.Name != r.Node {
		log.Info("Ignoring change for other node", "name", req.Name)
		return ctrl.Result{}, nil
	}

	defaultConfig := &microk8sv1alpha1.Configuration{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: "default"}, defaultConfig); err != nil && !apierrors.IsNotFound(err) {
		log.Error(err, "Failed to get default config")
		return ctrl.Result{}, err
	}
	config := &microk8sv1alpha1.Configuration{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: r.Node}, config); err != nil && !apierrors.IsNotFound(err) {
		log.Error(err, "Failed to get config object for node")
		return ctrl.Result{}, err
	}
	spec := mergeConfigSpecs(defaultConfig.Spec, config.Spec)

	for registry, toml := range spec.ContainerdRegistryConfigs {
		log := log.WithValues("registry", registry)
		dir := filepath.Join(r.SnapData, "args", "certs.d", registry)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Error(err, "Failed to setup directories for registry")
			continue
		}

		if err := os.WriteFile(filepath.Join(dir, "hosts.toml"), []byte(toml), 0660); err != nil {
			log.Error(err, "Failed to write registry configuration")
		}

		log.Info("Successfully configured registry")
	}

	for _, repo := range spec.AddonRepositories {
		log := log.WithValues("repository", repo.Name)
		dir := filepath.Join(r.SnapCommon, "addons", repo.Name)
		if err := os.RemoveAll(dir); err != nil {
			log.Error(err, "Failed to cleanup dir")
			continue
		}

		if err := exec.CommandContext(ctx, "git", "clone", "--depth=1", repo.Repository, dir).Run(); err != nil {
			log.Error(err, "Failed to fetch repository")
			continue
		}

		log.Info("Configured addon repository")
	}

	node := &microk8sv1alpha1.MicroK8sNode{}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: r.Node}, node); err != nil {
		if !errors.IsNotFound(err) {
			log.Error(err, "Failed to retrieve current node")
			return ctrl.Result{}, err
		}

		if err := r.Client.Create(ctx, &microk8sv1alpha1.MicroK8sNode{ObjectMeta: metav1.ObjectMeta{Name: r.Node}}); err != nil {
			log.Error(err, "Failed to create microk8s node")
			return ctrl.Result{}, err
		}
		log.Info("Initialized microk8s node")
	}
	if err := r.Client.Get(ctx, types.NamespacedName{Name: r.Node}, node); err != nil {
		log.Error(err, "Failed to retrieve current node")
		return ctrl.Result{}, err
	}

	node.Status.Version = "!!!!"
	if err := r.Client.Status().Update(ctx, node); err != nil {
		log.Error(err, "Failed to patch microk8s node")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigurationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&microk8sv1alpha1.Configuration{}).
		Complete(r)
}
