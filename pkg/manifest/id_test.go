package manifest_test

import (
	"context"
	"testing"

	"github.com/loft-orbital/cuebe/internal/mock"
	"github.com/loft-orbital/cuebe/internal/utils"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	. "github.com/loft-orbital/cuebe/pkg/manifest"
)

func TestManifestId(t *testing.T) {
	u := new(unstructured.Unstructured)
	u.SetKind("Deployment")
	u.SetAPIVersion("apps/v1")
	u.SetName("potato")
	u.SetNamespace("tomato")
	m := New(u)

	expected := Id{
		Kind:      "Deployment",
		Version:   "v1",
		Group:     "apps",
		Namespace: "tomato",
		Name:      "potato",
	}
	assert.Equal(t, expected, m.Id())
}

func TestIdManifest(t *testing.T) {
	konfig, tfake, client := utils.NewFakeK8sConfig()
	coreresources := &metav1.APIResourceList{
		GroupVersion: "testing",
		APIResources: []metav1.APIResource{
			{
				Name:       "manifests",
				Kind:       "Manifest",
				Verbs:      metav1.Verbs{"delete, create, patch, get"},
				Namespaced: false,
			},
		},
	}
	tfake.Resources = append(tfake.Resources, coreresources)
	cluster := mock.NewCluster(client, tfake.Resources...)

	m := NewUnique()
	m.SetUnstructuredContent(map[string]interface{}{"foo": "bar"})
	m.SetAPIVersion("testing")
	// Unknown kind
	m.SetKind("Foo")
	_, err := m.Id().Manifest(context.Background(), konfig.RESTMapper, client, metav1.GetOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not get resource interface:")

	// Not found
	m.SetKind("Manifest")
	_, err = m.Id().Manifest(context.Background(), konfig.RESTMapper, client, metav1.GetOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not get resource:")

	// Nominal
	cluster.Resources.Store(m.Id(), m)
	assert.True(t, cluster.Contains(m.Id()), "Cluster should contains the manifest before retrieving it (duh!)")
	actual, err := m.Id().Manifest(context.Background(), konfig.RESTMapper, client, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.Equal(t, m, actual)
}

func TestIdResourceInterface(t *testing.T) {
	konfig, tfake, client := utils.NewFakeK8sConfig()
	coreresources := &metav1.APIResourceList{
		GroupVersion: "testing",
		APIResources: []metav1.APIResource{
			{
				Name:       "manifests",
				Kind:       "Manifest",
				Verbs:      metav1.Verbs{"delete, create, patch, get"},
				Namespaced: false,
			},
		},
	}
	tfake.Resources = append(tfake.Resources, coreresources)

	m := NewUnique()
	m.SetAPIVersion("testing")
	m.SetKind("Foo")
	r, err := m.Id().ResourceInterface(konfig.RESTMapper, client)
	assert.Error(t, err)
	assert.Nil(t, r)

	m.SetKind("Manifest")
	r, err = m.Id().ResourceInterface(konfig.RESTMapper, client)
	assert.NoError(t, err)
	assert.Implements(t, (*dynamic.ResourceInterface)(nil), r)
}
