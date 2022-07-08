package configuration

import (
	"context"
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

// updateServiceArguments updates the arguments file for a service.
// updateMap is a map of key-value pairs. It will replace the argument with the new value (or just append).
// if a value is nil, then the argument is removed if present.
// returns true/false whether the service file was updated as well as the error that occured, or nil.
func updateServiceArguments(argumentsFile string, updateMap map[string]*string) (bool, error) {
	// If no updates are requested, exit early
	if len(updateMap) == 0 {
		return false, nil
	}

	arguments, err := os.ReadFile(argumentsFile)
	if err != nil {
		return false, fmt.Errorf("failed to read arguments file: %w", err)
	}

	for key, value := range updateMap {
		delete(updateMap, key)
		updateMap[fmt.Sprintf("--%s", strings.TrimLeft(key, "-"))] = value
	}

	existingArguments := make(map[string]struct{}, len(arguments))
	newArguments := make([]string, 0, len(arguments))
	for _, line := range strings.Split(string(arguments), "\n") {
		line = strings.TrimSpace(line)
		// ignore empty lines
		if line == "" {
			continue
		}
		// handle "--argument value" and "--argument=value" variants
		key := strings.SplitN(line, " ", 2)[0]
		key = strings.SplitN(key, "=", 2)[0]
		key = fmt.Sprintf("--%s", strings.TrimLeft(key, "-"))
		existingArguments[key] = struct{}{}
		if newValue, ok := updateMap[key]; ok {
			if newValue == nil {
				// remove argument
				continue
			} else {
				// update argument with new value
				newArguments = append(newArguments, fmt.Sprintf("%s=%s", key, *newValue))
			}
		} else {
			// no change
			newArguments = append(newArguments, line)
		}
	}

	for key, value := range updateMap {
		key = fmt.Sprintf("--%s", strings.TrimLeft(key, "-"))
		if _, argExists := existingArguments[key]; !argExists && value != nil {
			newArguments = append(newArguments, fmt.Sprintf("%s=%s", key, *value))
		}
	}

	updated, err := updateFile(argumentsFile, strings.Join(newArguments, "\n")+"\n", 0660)
	if err != nil {
		return updated, fmt.Errorf("failed to update arguments file: %w", err)
	}
	return updated, nil
}

func (r *Reconciler) reconcileKubeletArgs(ctx context.Context, args map[string]*string) error {
	if len(args) == 0 {
		return nil
	}
	log := log.FromContext(ctx)
	updated, err := updateServiceArguments(r.KubeletArgsFile, args)
	if err != nil {
		return fmt.Errorf("failed to update kubelet args file: %w", err)
	}
	if !updated {
		log.Info("kubelet arguments up to date")
		return nil
	}
	log.Info("updated kubelet arguments file")

	if err := r.RestartKubelet(ctx); err != nil {
		return fmt.Errorf("failed to restart kubelet: %w", err)
	}
	log.Info("restarted kubelet")
	return nil
}

func (r *Reconciler) reconcileKubeAPIServerArgs(ctx context.Context, args map[string]*string) error {
	if len(args) == 0 {
		return nil
	}
	log := log.FromContext(ctx)
	updated, err := updateServiceArguments(r.KubeAPIServerArgsFile, args)
	if err != nil {
		return fmt.Errorf("failed to update kube-apiserver args file: %w", err)
	}
	if !updated {
		log.Info("kube-apiserver arguments up to date")
		return nil
	}
	log.Info("updated kube-apiserver arguments file")

	if err := r.RestartKubeAPIServer(ctx); err != nil {
		return fmt.Errorf("failed to restart kube-apiserver: %w", err)
	}
	log.Info("restarted kube-apiserver")
	return nil
}
