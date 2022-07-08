package configuration

import (
	"fmt"
	"io/fs"
	"os"

	microk8sv1alpha1 "github.com/neoaggelos/microk8s-operator/api/v1alpha1"
)

func mergeMaps[T any](base map[string]T, overrides map[string]T) map[string]T {
	m := make(map[string]T, len(base)+len(overrides))
	for key, val := range base {
		m[key] = val
	}
	for key, val := range overrides {
		m[key] = val
	}
	return m
}

func mergeConfigSpecs(base, overrides microk8sv1alpha1.ConfigurationSpec) microk8sv1alpha1.ConfigurationSpec {
	result := microk8sv1alpha1.ConfigurationSpec{}

	result.ContainerdRegistryConfigs = mergeMaps(base.ContainerdRegistryConfigs, overrides.ContainerdRegistryConfigs)
	result.AddonRepositories = append(base.AddonRepositories, overrides.AddonRepositories...)
	result.ExtraSANIPs = append(base.ExtraSANIPs, overrides.ExtraSANIPs...)
	result.ExtraSANs = append(base.ExtraSANs, overrides.ExtraSANs...)
	result.PodCIDR = base.PodCIDR
	if o := overrides.PodCIDR; o != "" {
		result.PodCIDR = o
	}
	result.ContainerdEnv = base.ContainerdEnv
	if o := overrides.ContainerdEnv; o != "" {
		result.ContainerdEnv = o
	}
	result.ExtraKubeletArgs = mergeMaps(base.ExtraKubeletArgs, overrides.ExtraKubeletArgs)
	result.ExtraAPIServerArgs = mergeMaps(base.ExtraAPIServerArgs, overrides.ExtraAPIServerArgs)

	return result
}

func updateFile(file string, newContents string, perm fs.FileMode) (bool, error) {
	b, err := os.ReadFile(file)
	if err != nil {
		return false, fmt.Errorf("failed to read file: %w", err)
	}
	if string(b) == newContents {
		return false, nil
	}

	if err := os.WriteFile(file, []byte(newContents), perm); err != nil {
		return false, fmt.Errorf("failed to write file: %w", err)
	}
	return true, nil
}
