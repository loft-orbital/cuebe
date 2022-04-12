package instance

import (
	"context"
	"testing"

	"github.com/loft-orbital/cuebe/internal/utils"
	"github.com/loft-orbital/cuebe/pkg/manifest"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestOrphanNew(t *testing.T) {
	o := NewOrphan()
	assert.IsType(t, (*Orphan)(nil), o)
	assert.Implements(t, (*Instance)(nil), o)
}

func TestOrphanAdd(t *testing.T) {
	o := NewOrphan()
	m := newUniqueManifest()

	o.Add(m)

	mfs := o.Manifests()
	assert.Len(t, mfs, 1)
	assert.Equal(t, m, mfs[0])
}

func TestOrphanRemove(t *testing.T) {
	m := newUniqueManifest()
	o := &Orphan{manifests: map[manifest.Id]manifest.Manifest{m.Id(): m}}

	assert.Len(t, o.Manifests(), 1)
	o.Remove(m)
	assert.Len(t, o.Manifests(), 0)
}

func TestOrphanManifests(t *testing.T) {
	o := NewOrphan()

	assert.Len(t, o.Manifests(), 0)
	o.Add(newUniqueManifest())
	assert.Len(t, o.Manifests(), 1)
}

func TestOrphanCommit(t *testing.T) {
	konfig, tfake, client := utils.NewFakeK8sConfig()

	coreresources := &metav1.APIResourceList{
		GroupVersion: "v1",
		APIResources: []metav1.APIResource{
			{
				Name:       "configmaps",
				Kind:       "ConfigMap",
				Verbs:      metav1.Verbs{"patch"},
				Namespaced: true,
			},
		},
	}
	tfake.Resources = append(tfake.Resources, coreresources)

	o := NewOrphan()
	m := manifest.New(new(unstructured.Unstructured))
	m.SetName("cm-test")
	m.SetNamespace("default")
	m.SetAPIVersion("v1")
	m.SetKind("ConfigMap")
	o.Add(m)

	client.PrependReactor("patch", "configmaps", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		m.SetGenerateName("patched")
		return true, m, nil
	})

	assert.NoError(t, o.Commit(context.Background(), konfig, utils.CommonMetaOptions{}))
	assert.Equal(t, "patched", o.Manifests()[0].GetGenerateName())
}

func TestOrphanDelete(t *testing.T) {
	konfig, tfake, client := utils.NewFakeK8sConfig()
	coreresources := &metav1.APIResourceList{
		GroupVersion: "v1",
		APIResources: []metav1.APIResource{
			{
				Name:       "configmaps",
				Kind:       "ConfigMap",
				Verbs:      metav1.Verbs{"delete"},
				Namespaced: true,
			},
		},
	}
	tfake.Resources = append(tfake.Resources, coreresources)

	m := manifest.New(new(unstructured.Unstructured))
	m.SetName("cm-test")
	m.SetNamespace("default")
	m.SetAPIVersion("v1")
	m.SetKind("ConfigMap")

	o := NewOrphan()
	o.Add(m)

	client.PrependReactor("delete", "configmaps", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, nil
	})

	assert.NoError(t, o.Delete(context.Background(), konfig, utils.CommonMetaOptions{}))
	assert.Len(t, o.Manifests(), 0)
}
