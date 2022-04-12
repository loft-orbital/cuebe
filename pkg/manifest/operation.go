package manifest

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

const FieldManager = "cuebe"

// Delete deletes the Manifest from the cluster.
// If deletion policy is set to abandon, the resource is not deleted, but will be patched with an empty instance.
func (m Manifest) Delete(ctx context.Context, rm meta.RESTMapper, client dynamic.Interface, opts metav1.DeleteOptions) error {
	// TODO: implements a better logger
	fmt.Println("deleting", m.Id())

	if m.GetDeletionPolicy() == DeletionPolicyAbandon {
		m = m.WithInstance("")
		force := true
		_, err := m.Patch(ctx, rm, client, metav1.PatchOptions{
			TypeMeta: opts.TypeMeta,
			Force:    &force,
			DryRun:   opts.DryRun,
		})
		return err
	}

	resource, err := m.Id().ResourceInterface(rm, client)
	if err != nil {
		return fmt.Errorf("could not get resource interface: %w", err)
	}
	return resource.Delete(ctx, m.GetName(), opts)
}

// Patch does a server-side apply patch of the Manifest.
func (m Manifest) Patch(ctx context.Context, rm meta.RESTMapper, client dynamic.Interface, opts metav1.PatchOptions) (Manifest, error) {
	opts.FieldManager = FieldManager
	// TODO: implements a better logger
	fmt.Println("patching", m.Id())

	resource, err := m.Id().ResourceInterface(rm, client)
	if err != nil {
		return Manifest{}, fmt.Errorf("could not get resource interface: %w", err)
	}
	// marshal manifest
	data, err := m.MarshalJSON()
	if err != nil {
		return Manifest{}, fmt.Errorf("unable to marshal object: %w", err)
	}
	// server-side apply patch
	newo, err := resource.Patch(ctx, m.GetName(), types.ApplyPatchType, data, opts)

	return New(newo), err
}
