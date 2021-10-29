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
package kubernetes

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

const defaultNamespace = "default"

// PatchObjects patches a list of *unstructured.Unstructured object
func PatchObjects(ctx context.Context, cfg *rest.Config, objs *list.List) error {
	// Get dynamic client
	dync, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return err
	}

	// Get discovery client
	disc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return err
	}
	// Create the discovery cache
	dcache := memory.NewMemCacheClient(disc)
	// Get rest mapper
	rm := restmapper.NewDeferredDiscoveryRESTMapper(dcache)

	// Start patching
	for initlen := objs.Len(); initlen > 0; initlen = objs.Len() {
		dcache.Invalidate() // Invalidate the cache to get new resource mappings

		for e := objs.Front(); e != nil; e = e.Next() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Retrieve object
				obj, ok := e.Value.(*unstructured.Unstructured)
				if !ok {
					return fmt.Errorf("Unexpected type for %v, should be *unstructured.Unstructured", e.Value)
				}

				// Get mapping
				gvk := obj.GroupVersionKind()
				mapping, err := rm.RESTMapping(gvk.GroupKind(), gvk.Version)
				if err != nil {
					// couldn't find mapping, try next loop
					continue
				}
				var dr dynamic.ResourceInterface
				if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
					// Namespaced resources
					ns := obj.GetNamespace()
					if ns == "" {
						ns = defaultNamespace
					}
					dr = dync.Resource(mapping.Resource).Namespace(ns)
				} else {
					// Cluster-wide resources
					dr = dync.Resource(mapping.Resource)
				}
				data, err := json.Marshal(obj)
				if err != nil {
					return fmt.Errorf("Couldn't marshal %s: %w", printObj(obj), err)
				}

				// TODO maybe retrieve from option, to enable dryrun
				po := metav1.PatchOptions{
					FieldManager: "cuebe",
				}
				// Patch object
				if _, err := dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, po); err != nil {
					return fmt.Errorf("Couldn't patch %s: %w", printObj(obj), err)
				}
				fmt.Printf("%s patched\n", printObj(obj))
				// Object successfully patched, remove from queue
				objs.Remove(e)
			}
		}

		// Found a deadlock
		if objs.Len() >= initlen {
			return fmt.Errorf("Failed to patch every objects, %d remaining", objs.Len())
		}
	}

	return nil
}

func printObj(obj *unstructured.Unstructured) string {
	return strings.ToLower(fmt.Sprintf("%s/%s", obj.GetKind(), obj.GetName()))
}
