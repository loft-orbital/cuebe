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
	"context"
	"fmt"
	"io"
	"os"
	"sort"

	"cuelang.org/go/cue"
	"github.com/loft-orbital/cuebe/pkg/manifest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"sigs.k8s.io/yaml"
)

const fieldManager = "cuebe"

type Release struct {
	cfg *rest.Config

	// Objects managed by this release
	Objects []unstructured.Unstructured
}

// Create a new release by extracting manifests and context from a CUE value.
// If ktxpath fails to resolve to a string, ktxfallback will be used as the Kubernetes context.
func NewReleaseFor(v cue.Value, expressions []string, ktxpath string, ktxfallback string) (*Release, error) {
	// Get kubernetes client
	ktx, err := ExtractContext(v, ktxpath, ktxfallback)
	if err != nil {
		return nil, fmt.Errorf("failed to build release: %w", err)
	}
	cfg, err := DefaultConfig(ktx)
	if err != nil {
		return nil, fmt.Errorf("failed to build release: %w", err)
	}

	// Get manifests
	var mfs []manifest.Manifest
	if len(expressions) <= 0 {
		mfs = manifest.Extract(v)
	} else {
		for _, e := range expressions {
			ve := v.LookupPath(cue.ParsePath(e))
			if ve.Err() != nil {
				return nil, fmt.Errorf("failed to build release: %w", ve.Err())
			}
			mfs = append(mfs, manifest.Extract(ve)...)
		}
	}

	objs := make([]unstructured.Unstructured, len(mfs))
	for i, m := range mfs {
		objs[i], err = m.ToObj()
		if err != nil {
			return nil, fmt.Errorf("failed to build release: %w", err)
		}
	}
	sort.Sort(SortableUnstructured(objs))

	return &Release{
		cfg:     cfg,
		Objects: objs,
	}, nil
}

// Deploy patches every objects contained in the release to the Kubernetes cluster.
// It uses server-side apply method.
func (r *Release) Deploy(ctx context.Context, dryrun bool) error {
	// Get dynamic client
	dync, err := dynamic.NewForConfig(r.cfg)
	if err != nil {
		return err
	}

	// Get discovery client
	disc, err := discovery.NewDiscoveryClientForConfig(r.cfg)
	if err != nil {
		return err
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
					fmt.Fprintf(os.Stderr, "Could not apply %s: %s, retrying later...\n", printObj(obj), err)
					continue
				}

				data, err := obj.MarshalJSON()
				if err != nil {
					return fmt.Errorf("Couldn't marshal %s: %w", printObj(obj), err)
				}

				// Patch object
				var drSuffix string
				po := metav1.PatchOptions{
					FieldManager: fieldManager,
				}
				if dryrun {
					po.DryRun = []string{"All"}
					drSuffix = " (server dry-run)"
				}
				if _, err := dr.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, po); err != nil {
					return fmt.Errorf("Couldn't patch %s: %w", printObj(obj), err)
				}
				fmt.Printf("%s patched%s\n", printObj(obj), drSuffix)
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

// Host returns the kubernetes host this release targets.
func (r *Release) Host() string {
	return r.cfg.Host
}

// Render write a kubectl compatible yaml manifest to w.
func (r *Release) Render(w io.Writer) error {
	for _, obj := range r.Objects {
		by, err := yaml.Marshal(&obj)
		if err != nil {
			return fmt.Errorf("failed to render release: %w", err)
		}
		if _, err := w.Write([]byte("---\n")); err != nil {
			return fmt.Errorf("failed to render release: %w", err)
		}
		if _, err := w.Write(by); err != nil {
			return fmt.Errorf("failed to render release: %w", err)
		}
	}
	return nil
}
