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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

const SystemNamespace = "cuebe-system"

// Instance represents an association of multiples resources that together form an application.
type Instance struct {
	name string

	objects []unstructured.Unstructured
}

// Name returns the name of the instance.
func (i *Instance) Name() string {
	return i.name
}

func (i *Instance) GetManagedResources(ctx context.Context, client *kubernetes.Clientset) (map[string]string, error) {
	cm, err := client.CoreV1().ConfigMaps(SystemNamespace).Get(ctx, i.name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("could not get config map: %w", err)
	}
	return cm.Data, nil
}

// Patch applies every instance's objects to a k8s cluster.
func (i *Instance) Patch(ctx context.Context, rm *restmapper.DeferredDiscoveryRESTMapper, dync dynamic.Interface, opts metav1.PatchOptions) error {
	todo := make([]unstructured.Unstructured, len(i.objects)+1)
	copy(todo, i.objects)
	for {
		rm.Reset() // invalidate the cache to get new resource mappings
		var done []unstructured.Unstructured
		var err error
		done, todo, err = PatchSlice(ctx, todo, rm, dync, opts)
		for _, obj := range done {
			fmt.Printf("%s patched", printObj(obj))
			if len(opts.DryRun) > 0 {
				fmt.Printf(" (dry-run %s)", opts.DryRun)
			}
			fmt.Println()
		}

		if len(todo) <= 0 {
			// finished
			return err
		}

		if len(done) <= 0 {
			// found a deadlock
			return fmt.Errorf("failed to patch (deadlock): %w", err)
		}

		// Wait to retry
		ticker := time.Tick(time.Second)
		for i := 5; i >= 0; i-- {
			<-ticker
			fmt.Printf("\r%d patch failed, retrying in %d...", len(todo), i)
		}
		fmt.Print("\n\n")
	}
}
