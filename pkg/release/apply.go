/*
Copyright © 2021 Loft Orbital

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
package release

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultNamespace = "default"

	FieldManager = "cuebe"
)

var (
	DefaultPatchOptions = metav1.PatchOptions{FieldManager: FieldManager}
)

func (r *Release) Apply(ctx context.Context, po metav1.PatchOptions) error {
	// get config for context
	cfg, err := DefaultConfig(r.Context)
	if err != nil {
		return fmt.Errorf("failed to get k8s config: %w", err)
	}

	// Get dynamic client
	dync, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to get dynamic client: %w", err)
	}

	// Get discovery client
	disc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to get discovery client: %w", err)
	}
	// Get rest mapper
	rm := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(disc))

	todo := make([]unstructured.Unstructured, len(r.Objects))
	copy(todo, r.Objects)
	failed := []unstructured.Unstructured{}
	// Start patching
	for len(todo) > 0 {
		rm.Reset() // Invalidate the cache to get new resource mappings

		for _, obj := range todo {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				dr, err := dynamicResourceInterfaceFor(obj, rm, dync)
				if err != nil {
					failed = append(failed, obj) // Couldn't find any mapping, retrying next loop
					fmt.Printf("Could not apply %s: %s, retrying later...\n", printObj(obj), err)
					continue
				}

				data, err := obj.MarshalJSON()
				if err != nil {
					return fmt.Errorf("Couldn't marshal %s: %w", printObj(obj), err)
				}

				// Patch object
				if _, err := dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, po); err != nil {
					return fmt.Errorf("Couldn't patch %s: %w", printObj(obj), err)
				}
				fmt.Printf("%s patched", printObj(obj))
				if len(po.DryRun) > 0 {
					fmt.Printf(" (dry-run %s)", po.DryRun)
				}
				fmt.Println()
			}
		}

		// Found a deadlock
		if len(failed) >= len(todo) {
			return fmt.Errorf("Failed to patch every objects, %d remaining", len(failed))
		}
		todo = failed
		failed = []unstructured.Unstructured{}
	}

	return nil
}

// DefaultConfig returns the kubernetes config and client from default configuration.
// The default context is used if context is empty.
func DefaultConfig(context string) (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}

	if context != "" {
		overrides.CurrentContext = context
	}
	ccfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

	config, err := ccfg.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get Kubernetes config for context %q: %w", context, err)
	}
	return config, nil
}

func dynamicResourceInterfaceFor(obj unstructured.Unstructured, rm *restmapper.DeferredDiscoveryRESTMapper, dc dynamic.Interface) (dynamic.ResourceInterface, error) {
	// Get mapping
	gvk := obj.GroupVersionKind()
	mapping, err := rm.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// Namespaced resources
		ns := obj.GetNamespace()
		if ns == "" {
			ns = defaultNamespace
		}
		dr = dc.Resource(mapping.Resource).Namespace(ns)
	} else {
		// Cluster-wide resources
		dr = dc.Resource(mapping.Resource)
	}
	return dr, nil
}

func printObj(obj unstructured.Unstructured) string {
	return strings.ToLower(fmt.Sprintf("%s/%s", obj.GetKind(), obj.GetName()))
}
