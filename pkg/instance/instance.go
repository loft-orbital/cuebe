package instance

import (
	"context"
	"fmt"

	"github.com/loft-orbital/cuebe/internal/utils"
	"github.com/loft-orbital/cuebe/pkg/manifest"
)

type Instance interface {
	fmt.Stringer

	// Commit applies the instance remotely.
	Commit(ctx context.Context, config *utils.K8sConfig, opts utils.CommonMetaOptions) error
	// Delete deletes the instance from the cluster.
	Delete(ctx context.Context, config *utils.K8sConfig, opts utils.CommonMetaOptions) error

	// Add adds a Manifest to the instance.
	//
	// It does not commit changes remotely, use Commit for that.
	Add(m manifest.Manifest)
	// Remove removes a Manifest from the instance
	//
	// It does not commit changes remotely, use Commit for that.
	Remove(m manifest.Manifest)
	// Manifests returns the instance manifests.
	Manifests() []manifest.Manifest
}

// Split groups manifest by instance and returns the instances.
// Manifests that do not belong to an instance will be added to the orphan instance.
func Split(manifests []manifest.Manifest) []Instance {
	instances := make([]Instance, 0)
	mem := make(map[string]Instance)

	for _, m := range manifests {
		iname := m.GetInstance()
		instance, ok := mem[iname]
		if !ok {
			if iname != "" {
				instance = NewNamed(iname)
			} else {
				instance = NewOrphan()
			}
			mem[iname] = instance
			instances = append(instances, instance)
		}
		instance.Add(m)
	}

	return instances
}
