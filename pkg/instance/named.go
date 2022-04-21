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
package instance

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/loft-orbital/cuebe/internal/utils"
	"github.com/loft-orbital/cuebe/pkg/manifest"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type action int

const (
	actionDelete action = iota
	actionPatch
)

const (
	Kind     = "Instance"
	Group    = "cuebe.loftorbital.com"
	Version  = "v1alpha1"
	Resource = "instances"
)

var gvk = schema.GroupVersionResource{
	Group:    Group,
	Version:  Version,
	Resource: Resource,
}

// InstanceSpec contains the ids of the manifests this instance manage.
type InstanceSpec struct {
	Resources []manifest.Id `json:"resources,omitempty"`
}

// Named represents a named instance.
type Named struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InstanceSpec `json:"spec,omitempty"`

	manifests map[manifest.Id]manifest.Manifest
	mguard    sync.Mutex
}

// NewNamed creates a new named instance.
func NewNamed(name string) *Named {
	return &Named{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Instance",
			APIVersion: "cuebe.loftorbital.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},

		manifests: make(map[manifest.Id]manifest.Manifest),
	}
}

// Id returns the instance id as a manifest.Id.
// Mostly useful for test or debug purposes.
func (n *Named) Id() manifest.Id {
	gvk := n.GroupVersionKind()
	return manifest.Id{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
		Name:    n.Name,
	}
}

// Manifests returns the instance manifests.
func (i *Named) Manifests() []manifest.Manifest {
	i.mguard.Lock()
	defer i.mguard.Unlock()

	mfs := make([]manifest.Manifest, 0, len(i.manifests))
	for _, m := range i.manifests {
		mfs = append(mfs, m)
	}
	return mfs
}

// Add adds a manifest to the instance.
// It's thread safe.
func (i *Named) Add(m manifest.Manifest) {
	i.mguard.Lock()
	defer i.mguard.Unlock()
	// do not override local manifest with remote ones
	current, exists := i.manifests[m.Id()]
	if exists && m.IsRemote() && !current.IsRemote() {
		return
	}

	i.manifests[m.Id()] = m
}

// Remove deletes a manifest from the instance.
// It's thread safe.
func (i *Named) Remove(m manifest.Manifest) {
	i.mguard.Lock()
	defer i.mguard.Unlock()
	delete(i.manifests, m.Id())
}

// OwnerReference returns a metav1.OwnerReference to be used
// for managed resources.
func (i *Named) OwnerReference() metav1.OwnerReference {
	block := true
	return metav1.OwnerReference{
		APIVersion:         i.APIVersion,
		Kind:               i.Kind,
		Name:               i.Name,
		UID:                i.UID,
		BlockOwnerDeletion: &block,
	}
}

// Marshal encode the instance.
func (i *Named) Marshal() ([]byte, error) {
	return json.Marshal(i)
}

// Unmarshal decode the instance inplace.
func (i *Named) Unmarshal(b []byte) error {
	return json.Unmarshal(b, i)
}

// Delete deletes the instance from the cluster.
func (i *Named) Delete(ctx context.Context, config *utils.K8sConfig, opts utils.CommonMetaOptions) error {
	resource := config.DynamicClient.Resource(gvk)

	policy := metav1.DeletePropagationForeground
	opts.PropagationPolicy = &policy

	// lock when we're deleting
	i.mguard.Lock()
	defer i.mguard.Unlock()

	if err := resource.Delete(ctx, i.Name, opts.DeleteOptions()); err != nil {
		return err
	}

	i.Spec.Resources = make([]manifest.Id, 0)
	i.manifests = make(map[manifest.Id]manifest.Manifest)
	return nil
}

// Commit applies the instance remotely.
func (i *Named) Commit(ctx context.Context, config *utils.K8sConfig, opts utils.CommonMetaOptions) error {
	// make sure we're up to date
	if err := i.Sync(ctx, config, opts); err != nil {
		return fmt.Errorf("could not synchronize instance: %w", err)
	}

	// lock until we finish commiting
	i.mguard.Lock()
	defer i.mguard.Unlock()

	// gather what we need to do
	mfs, err := i.prepareCommit(ctx, config, opts)
	if err != nil {
		return err
	}

	// apply manifest changes
	cerr := make(chan error, len(mfs)+1)
	cid := make(chan manifest.Id, len(mfs))
	var wg sync.WaitGroup
	config.RESTMapper.Reset()
	for m, a := range mfs {
		wg.Add(1)
		go func(m manifest.Manifest, a action) {
			defer wg.Done()
			cerr <- i.applyManifest(m, a, cid, ctx, config, opts)
		}(m, a)
	}
	wg.Wait()
	close(cid)

	// apply instance changes
	i.Spec.Resources = make([]manifest.Id, 0, len(i.Spec.Resources)+len(i.manifests))
	for id := range cid {
		i.Spec.Resources = append(i.Spec.Resources, id)
	}
	cerr <- i.patch(ctx, config, opts)
	close(cerr)

	return utils.CollectErrors(cerr)
}

func (i *Named) prepareCommit(ctx context.Context, config *utils.K8sConfig, opts utils.CommonMetaOptions) (map[manifest.Manifest]action, error) {
	res := make(map[manifest.Manifest]action, len(i.manifests))

	// look for prune
	for _, id := range i.Spec.Resources {
		if _, ok := i.manifests[id]; !ok {
			m, err := id.Manifest(ctx, config.RESTMapper, config.DynamicClient, opts.GetOptions())
			if err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return nil, fmt.Errorf("could not get manifest %s: %w", id, err)
			}
			res[m] = actionDelete
		}
	}

	// add the rest
	for _, m := range i.manifests {
		res[m] = actionPatch
	}

	return res, nil
}

func (i *Named) applyManifest(m manifest.Manifest, a action, cid chan<- manifest.Id, ctx context.Context, config *utils.K8sConfig, opts utils.CommonMetaOptions) error {
	switch a {
	case actionDelete:
		if err := m.Delete(ctx, config, opts); err != nil {
			// we could not delete the manifest, so keep its reference in the instance.
			cid <- m.Id()
			return err
		}
		return nil
	case actionPatch:
		// Add owner reference if deletion policy allows it
		if m.GetDeletionPolicy() != manifest.DeletionPolicyAbandon {
			m.SetOwnerReferences([]metav1.OwnerReference{i.OwnerReference()})
		}

		// patch
		_, err := m.Patch(ctx, config, opts)
		if err != nil {
			return err
		}
		cid <- m.Id()
		return nil
	default:
		return fmt.Errorf("unexpected action %d", a)
	}
}

func (i *Named) patch(ctx context.Context, config *utils.K8sConfig, opts utils.CommonMetaOptions) error {
	i.ObjectMeta.ManagedFields = nil
	// marshal
	data, err := i.Marshal()
	if err != nil {
		return fmt.Errorf("could not marshal instance: %w", err)
	}
	// apply
	resource := config.DynamicClient.Resource(gvk)
	u, err := resource.Patch(ctx, i.Name, types.ApplyPatchType, data, opts.PatchOptions())
	if err != nil {
		return fmt.Errorf("could not create instance: %w", err)
	}
	// reflect changes
	return i.reflect(u)
}

func (i *Named) reflect(u *unstructured.Unstructured) error {
	raw, err := u.MarshalJSON()
	if err != nil {
		return err
	}
	return i.Unmarshal(raw)
}

func (i *Named) Sync(ctx context.Context, config *utils.K8sConfig, opts utils.CommonMetaOptions) error {
	resource := config.DynamicClient.Resource(gvk)
	u, err := resource.Get(ctx, i.Name, opts.GetOptions())
	if err != nil {
		if errors.IsNotFound(err) {
			return i.patch(ctx, config, opts)
		}
		return fmt.Errorf("could not get instance: %w", err)
	}
	if err := i.reflect(u); err != nil {
		return fmt.Errorf("could not reflect changes: %w", err)
	}

	return nil
}

func (i *Named) String() string {
	return i.Name
}
