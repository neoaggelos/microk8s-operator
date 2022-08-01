package configuration

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	microk8sv1alpha1 "github.com/neoaggelos/microk8s-operator/api/v1alpha1"
)

func mergeMaps(base map[string]string, overrides map[string]string) map[string]string {
	m := make(map[string]string, len(base)+len(overrides))
	for key, val := range base {
		m[key] = val
	}
	for key, val := range overrides {
		m[key] = val
	}
	return m
}

func mergeArguments(base map[string]*string, overrides map[string]*string) map[string]*string {
	m := make(map[string]*string, len(base)+len(overrides))
	for key, val := range base {
		m[fmt.Sprintf("--%s", strings.TrimLeft(key, "-"))] = val
	}
	for key, val := range overrides {
		m[fmt.Sprintf("--%s", strings.TrimLeft(key, "-"))] = val
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
	result.ExtraKubeletArgs = mergeArguments(base.ExtraKubeletArgs, overrides.ExtraKubeletArgs)
	result.ExtraAPIServerArgs = mergeArguments(base.ExtraAPIServerArgs, overrides.ExtraAPIServerArgs)

	return result
}

func updateFile(file string, newContents string, perm fs.FileMode) (bool, error) {
	b, err := os.ReadFile(file)
	if err != nil && !os.IsNotExist(err) {
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
