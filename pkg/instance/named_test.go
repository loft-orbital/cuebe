package instance

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/loft-orbital/cuebe/internal/mock"
	"github.com/loft-orbital/cuebe/internal/utils"
	"github.com/loft-orbital/cuebe/pkg/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func newUniqueManifest() manifest.Manifest {
	m := manifest.New(new(unstructured.Unstructured))
	name := make([]byte, 16)
	rand.Read(name)
	m.SetName(fmt.Sprintf("%s-%d", base64.StdEncoding.EncodeToString(name), time.Now().UnixNano()))

	return m
}

func TestNewNamed(t *testing.T) {
	i := NewNamed("potato")
	assert.IsType(t, (*Named)(nil), i)
	assert.Implements(t, (*Instance)(nil), i)
}

func TestNamedManifests(t *testing.T) {
	i := &Named{
		ObjectMeta: metav1.ObjectMeta{
			Name: "potato",
		},
		manifests: map[manifest.Id]manifest.Manifest{
			{Name: "one"}:   newUniqueManifest(),
			{Name: "two"}:   newUniqueManifest(),
			{Name: "three"}: newUniqueManifest(),
		},
	}
	assert.Len(t, i.Manifests(), 3)
}

func TestNamedAdd(t *testing.T) {
	instance := NewNamed("potato")
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			instance.Add(newUniqueManifest())
		}()
	}
	wg.Wait()

	assert.Len(t, instance.Manifests(), 100)
}

func TestNamedRemove(t *testing.T) {
	mfs := []manifest.Manifest{
		newUniqueManifest(),
		newUniqueManifest(),
		newUniqueManifest(),
	}

	instance := &Named{manifests: map[manifest.Id]manifest.Manifest{
		mfs[0].Id(): mfs[0],
		// we don't add 1 on purpose to test deletion of a manifest that does not exist
		mfs[2].Id(): mfs[2],
	}}
	var wg sync.WaitGroup

	for i := 0; i < len(mfs)-1; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			instance.Remove(mfs[i])
		}(i)
	}
	wg.Wait()

	assert.Len(t, instance.Manifests(), 1)
}

func TestNamedMarshal(t *testing.T) {
	instance := NewNamed("potato")
	b, err := instance.Marshal()
	assert.NoError(t, err)
	expected := `
{
   "apiVersion":"cuebe.loftorbital.com/v1alpha1",
   "kind":"Instance",
   "metadata":{
      "creationTimestamp":null,
      "name":"potato"
   },
   "spec":{

   }
}
`
	assert.JSONEq(t, expected, string(b))
}

func TestNamedUnmarshal(t *testing.T) {
	m := newUniqueManifest()
	u := new(unstructured.Unstructured)
	u.SetName("tomato")
	u.SetUnstructuredContent(map[string]interface{}{
		"spec": InstanceSpec{Resources: []manifest.Id{m.Id()}},
	})
	data, err := u.MarshalJSON()
	require.NoError(t, err)

	ni := NewNamed("potato")
	assert.NoError(t, ni.Unmarshal(data))
	assert.Equal(t, m.Id(), ni.Spec.Resources[0])
}

func TestNamedDelete(t *testing.T) {
	konfig, tfake, client := utils.NewFakeK8sConfig()
	coreresources := &metav1.APIResourceList{
		GroupVersion: Group + "/" + Version,
		APIResources: []metav1.APIResource{
			{
				Name:       Resource,
				Kind:       Kind,
				Verbs:      metav1.Verbs{"delete"},
				Namespaced: false,
			},
		},
	}
	tfake.Resources = append(tfake.Resources, coreresources)

	m := manifest.New(new(unstructured.Unstructured))
	ni := NewNamed("potato")
	ni.Add(m)
	cluster := mock.NewCluster(client, tfake.Resources...)
	rid := ni.Id()
	cluster.Resources.Store(rid, ni)

	assert.NoError(t, ni.Delete(context.Background(), konfig, utils.CommonMetaOptions{}))
	assert.Len(t, ni.Manifests(), 0)
	assert.Len(t, ni.Spec.Resources, 0)
	assert.Falsef(t, cluster.Contains(rid), "Cluster should not contain the instance after delete")
}

func TestNamedCommit(t *testing.T) {
	// prepare cluster
	konfig, tfake, client := utils.NewFakeK8sConfig()
	coreresources := &metav1.APIResourceList{
		GroupVersion: "v1",
		APIResources: []metav1.APIResource{
			{
				Name:       "configmaps",
				Kind:       "ConfigMap",
				Verbs:      metav1.Verbs{"delete, create, patch, get"},
				Namespaced: true,
			},
		},
	}
	instresources := &metav1.APIResourceList{
		GroupVersion: Group + "/" + Version,
		APIResources: []metav1.APIResource{
			{
				Name:       Resource,
				Kind:       Kind,
				Verbs:      metav1.Verbs{"delete, create, patch, get"},
				Namespaced: false,
			},
		},
	}
	tfake.Resources = append(tfake.Resources, coreresources, instresources)
	cluster := mock.NewCluster(client, tfake.Resources...)

	// prepare instance
	ni := NewNamed("potato")
	m := newUniqueManifest()
	m.SetKind("ConfigMap")
	m.SetAPIVersion("v1")
	m.SetNamespace("default")
	ni.Add(m)
	mprune := newUniqueManifest()
	mprune.SetKind("ConfigMap")
	mprune.SetAPIVersion("v1")
	mprune.SetNamespace("default")
	ni.Spec.Resources = []manifest.Id{mprune.Id()}
	cluster.Resources.Store(mprune.Id(), mprune)

	// commit
	assert.NoError(t, ni.Commit(context.Background(), konfig, utils.CommonMetaOptions{}))

	// verify
	assert.True(t, cluster.Contains(ni.Id()), "Cluster should contain the instance after commit")
	assert.True(t, cluster.Contains(m.Id()), "Cluster should contain the manifest after commit")
	assert.False(t, cluster.Contains(mprune.Id()), "Cluster should contain the pruned manifest after commit")
	assert.Len(t, ni.Spec.Resources, 1)
	assert.Equal(t, m.Id(), ni.Spec.Resources[0])
}

func TestNamedGet(t *testing.T) {
	konfig, tfake, client := utils.NewFakeK8sConfig()
	instresources := &metav1.APIResourceList{
		GroupVersion: Group + "/" + Version,
		APIResources: []metav1.APIResource{
			{
				Name:       Resource,
				Kind:       Kind,
				Verbs:      metav1.Verbs{"delete, create, patch, get"},
				Namespaced: false,
			},
		},
	}
	tfake.Resources = append(tfake.Resources, instresources)
	cluster := mock.NewCluster(client, tfake.Resources...)

	i := NewNamed("potato")
	cluster.Resources.Store(i.Id(), i)

	actual, err := Get(context.Background(), "tomato", konfig, utils.CommonMetaOptions{})
	assert.Nil(t, actual)
	assert.EqualError(t, err, "Could not get instance: instances.cuebe.loftorbital.com \"tomato\" not found")

	actual, err = Get(context.Background(), "potato", konfig, utils.CommonMetaOptions{})
	assert.NoError(t, err)
	assert.Implements(t, (*Instance)(nil), actual)
}
