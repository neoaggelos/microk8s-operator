package controllers

import (
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

func mergeConfigSpecs(base, overrides microk8sv1alpha1.ConfigurationSpec) microk8sv1alpha1.ConfigurationSpec {
	result := microk8sv1alpha1.ConfigurationSpec{}

	result.ContainerdRegistryConfigs = mergeMaps(base.ContainerdRegistryConfigs, overrides.ContainerdRegistryConfigs)
	result.AddonRepositories = append(base.AddonRepositories, overrides.AddonRepositories...)
	result.ExtraSANIPs = append(base.ExtraSANIPs, overrides.ExtraSANIPs...)
	result.ExtraSANs = append(base.ExtraSANs, overrides.ExtraSANs...)
	result.ExtraKubeletArgs = append(base.ExtraKubeletArgs, overrides.ExtraKubeletArgs...)
	result.PodCIDR = base.PodCIDR
	if o := overrides.PodCIDR; o != "" {
		result.PodCIDR = o
	}

	return result
}
