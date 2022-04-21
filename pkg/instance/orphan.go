package instance

import (
	"context"
	"fmt"
	"sync"

	"github.com/loft-orbital/cuebe/internal/utils"
	"github.com/loft-orbital/cuebe/pkg/manifest"
)

type Orphan struct {
	manifests map[manifest.Id]manifest.Manifest
	guard     sync.Mutex
}

// NewOrphan creates a new Orphan instance.
func NewOrphan() *Orphan {
	return new(Orphan)
}

// Commit applies the instance remotely.
func (o *Orphan) Commit(ctx context.Context, config *utils.K8sConfig, opts utils.CommonMetaOptions) error {
	mfs := o.Manifests()

	cerr := make(chan error, len(mfs))
	var wg sync.WaitGroup
	for _, m := range mfs {
		wg.Add(1)
		go func(m manifest.Manifest) {
			defer wg.Done()
			newM, err := m.Patch(ctx, config, opts)
			if err != nil {
				cerr <- fmt.Errorf("applying manifest %s: %w", m.Id(), err)
				return
			}
			o.Add(newM)
		}(m)
	}
	wg.Wait()
	close(cerr)

	return utils.CollectErrors(cerr)
}

// Delete deletes the instance from the cluster.
func (o *Orphan) Delete(ctx context.Context, config *utils.K8sConfig, opts utils.CommonMetaOptions) error {
	mfs := o.Manifests()

	cerr := make(chan error, len(mfs))
	var wg sync.WaitGroup
	for _, m := range mfs {
		wg.Add(1)
		go func(m manifest.Manifest) {
			defer wg.Done()
			err := m.Delete(ctx, config, opts)
			if err != nil {
				cerr <- fmt.Errorf("deleting manifest %s: %w", m.Id(), err)
				return
			}
			o.Remove(m)
		}(m)
	}
	wg.Wait()
	close(cerr)

	return utils.CollectErrors(cerr)
}

// Add adds a Manifest to the instance.
//
// It does not commit changes remotely, use Commit for that.
func (o *Orphan) Add(m manifest.Manifest) {
	o.guard.Lock()
	defer o.guard.Unlock()
	if o.manifests == nil {
		o.manifests = make(map[manifest.Id]manifest.Manifest, 1)
	}
	o.manifests[m.Id()] = m
}

// Remove removes a Manifest from the instance
//
// It does not commit changes remotely, use Commit for that.
func (o *Orphan) Remove(m manifest.Manifest) {
	o.guard.Lock()
	defer o.guard.Unlock()
	delete(o.manifests, m.Id())
}

// Manifests returns the instance manifests.
func (o *Orphan) Manifests() []manifest.Manifest {
	o.guard.Lock()
	defer o.guard.Unlock()

	mfs := make([]manifest.Manifest, 0, len(o.manifests))
	for _, m := range o.manifests {
		mfs = append(mfs, m)
	}
	return mfs
}

func (o *Orphan) String() string {
	return "orphan"
}
