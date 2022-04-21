package manifest

import (
	"context"
	"fmt"

	"github.com/loft-orbital/cuebe/internal/utils"
	"github.com/loft-orbital/cuebe/pkg/log"
	"k8s.io/apimachinery/pkg/types"
)

const FieldManager = "cuebe"

// Delete deletes the Manifest from the cluster.
// If deletion policy is set to abandon, the resource is not deleted, but will be patched with an empty instance.
func (m Manifest) Delete(ctx context.Context, konfig *utils.K8sConfig, opts utils.CommonMetaOptions) error {
	if m.GetDeletionPolicy() == DeletionPolicyAbandon {
		m = m.WithInstance("")
		_, err := m.Patch(ctx, konfig, opts)
		return err
	}

	resource, err := m.Id().ResourceInterface(konfig.RESTMapper, konfig.DynamicClient)
	if err != nil {
		return fmt.Errorf("could not get resource interface: %w", err)
	}
	if err := resource.Delete(ctx, m.GetName(), opts.DeleteOptions()); err != nil {
		return err
	}

	logger := log.GetLogger(ctx)
	logger.Info("%s deleted\n", m.Id())
	return nil
}

// Patch does a server-side apply patch of the Manifest.
func (m Manifest) Patch(ctx context.Context, konfig *utils.K8sConfig, opts utils.CommonMetaOptions) (Manifest, error) {
	resource, err := m.Id().ResourceInterface(konfig.RESTMapper, konfig.DynamicClient)
	if err != nil {
		return Manifest{}, fmt.Errorf("could not get resource interface: %w", err)
	}
	// marshal manifest
	data, err := m.MarshalJSON()
	if err != nil {
		return Manifest{}, fmt.Errorf("unable to marshal object: %w", err)
	}
	// server-side apply patch
	newo, err := resource.Patch(ctx, m.GetName(), types.ApplyPatchType, data, opts.PatchOptions())
	if err != nil {
		return Manifest{}, err
	}

	logger := log.GetLogger(ctx)
	logger.Info("%s patched\n", m.Id())
	return New(newo), nil
}
