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
	"sync"

	"github.com/hashicorp/go-multierror"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

const (
	defaultNamespace = "default"

	FieldManager = "cuebe"
)

var (
	DefaultPatchOptions = metav1.PatchOptions{FieldManager: FieldManager}
)

// Patch patches an unstructured object.
func Patch(ctx context.Context, obj unstructured.Unstructured, rm meta.RESTMapper, dc dynamic.Interface, opts metav1.PatchOptions) (*unstructured.Unstructured, error) {
	// get mapping
	gvk := obj.GroupVersionKind()
	mapping, err := rm.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("Could not get mapping: %w", err)
	}
	// create dynamic resource interface...
	var dri dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// ...for namespaced resources
		ns := obj.GetNamespace()
		if ns == "" {
			ns = defaultNamespace
		}
		dri = dc.Resource(mapping.Resource).Namespace(ns)
	} else {
		// ...for cluster-wide resources
		dri = dc.Resource(mapping.Resource)
	}
	// marshal object
	data, err := obj.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("unable to marshal object: %w", err)
	}
	// do patch
	return dri.Patch(ctx, obj.GetName(), types.ApplyPatchType, data, opts)
}

func PatchSlice(ctx context.Context, objs []unstructured.Unstructured, rm meta.RESTMapper, dc dynamic.Interface, opts metav1.PatchOptions) (done, failed []unstructured.Unstructured, err error) {
	var rerr *multierror.Error
	var m sync.Mutex
	var wg sync.WaitGroup

	for _, o := range objs {
		wg.Add(1)
		go func(obj unstructured.Unstructured) {
			defer wg.Done()
			_, err := Patch(ctx, obj, rm, dc, opts)

			m.Lock()
			defer m.Unlock()
			if err != nil {
				rerr = multierror.Append(rerr, fmt.Errorf("%s patch failed: %w", printObj(obj), err))
				failed = append(failed, obj)
			} else {
				done = append(done, obj)
			}
		}(o)
	}
	wg.Wait()

	return done, failed, rerr.ErrorOrNil()
}
