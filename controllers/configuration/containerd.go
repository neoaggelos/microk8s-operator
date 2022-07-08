package configuration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *Reconciler) reconcileContainerdEnv(ctx context.Context, env string) error {
	log := log.FromContext(ctx)
	if env == "" {
		return nil
	}

	updated, err := updateFile(r.ContainerdEnvFile, env, 0660)
	if err != nil {
		return fmt.Errorf("failed to update containerd environment file: %w", err)
	}
	if !updated {
		log.Info("containerd environment file up to date")
		return nil
	}
	log.Info("updated containerd environment file")

	if err := r.RestartContainerd(ctx); err != nil {
		return fmt.Errorf("failed to restart containerd service: %w", err)
	}
	log.Info("restarted containerd service")
	return nil
}

func (r *Reconciler) reconcileRegistryConfigs(ctx context.Context, registries map[string]string) {
	log := log.FromContext(ctx)
	for registry, toml := range registries {
		log := log.WithValues("registry", registry)
		dir := filepath.Join(r.RegistryCertsDir, registry)
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Error(err, "failed to setup directories")
			continue
		}

		updated, err := updateFile(filepath.Join(dir, "hosts.toml"), toml, 0660)
		if err != nil {
			log.Error(err, "failed to update hosts.toml")
			continue
		}
		if updated {
			log.Info("updated registry configuration")
		} else {
			log.Info("registry configuration is up to date")
		}
	}
}
