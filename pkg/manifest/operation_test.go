package manifest_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/loft-orbital/cuebe/internal/mock"
	"github.com/loft-orbital/cuebe/internal/utils"
	"github.com/loft-orbital/cuebe/pkg/manifest"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	. "github.com/loft-orbital/cuebe/pkg/manifest"
)

// NewUnique returns a manifest with a unique random name.
func NewUnique() Manifest {
	m := manifest.New(new(unstructured.Unstructured))
	name := make([]byte, 16)
	rand.Read(name)
	m.SetName(fmt.Sprintf("%s-%d", base64.StdEncoding.EncodeToString(name), time.Now().UnixNano()))

	return m
}

func TestManifestDelete(t *testing.T) {
	konfig, tfake, client := utils.NewFakeK8sConfig()
	coreresources := &metav1.APIResourceList{
		GroupVersion: "testing",
		APIResources: []metav1.APIResource{
			{
				Name:  "manifests",
				Kind:  "Manifest",
				Verbs: metav1.Verbs{"delete, create, patch, get"},
			},
		},
	}
	tfake.Resources = append(tfake.Resources, coreresources)
	cluster := mock.NewCluster(client, tfake.Resources...)

	m := NewUnique()
	m.SetAPIVersion("testing")
	m.SetKind("Foo")
	assert.Error(t, m.Delete(context.Background(), konfig.RESTMapper, client, metav1.DeleteOptions{}))

	m.SetKind("Manifest")
	cluster.Resources.Store(m.Id(), m)

	assert.True(t, cluster.Contains(m.Id()), "Cluster should contains the manifest before delete")
	assert.NoError(t, m.Delete(context.Background(), konfig.RESTMapper, client, metav1.DeleteOptions{}))
	assert.False(t, cluster.Contains(m.Id()), "Cluster should not contains the manifest after delete")
}

func TestManifestDeleteAbandon(t *testing.T) {
	konfig, tfake, client := utils.NewFakeK8sConfig()
	coreresources := &metav1.APIResourceList{
		GroupVersion: "testing",
		APIResources: []metav1.APIResource{
			{
				Name:       "manifests",
				Kind:       "Manifest",
				Verbs:      metav1.Verbs{"delete, create, patch, get"},
				Namespaced: true,
			},
		},
	}
	tfake.Resources = append(tfake.Resources, coreresources)
	cluster := mock.NewCluster(client, tfake.Resources...)

	m := NewUnique().WithInstance("potato").WithDeletionPolicy(DeletionPolicyAbandon)
	m.SetAPIVersion("testing")
	m.SetKind("Manifest")
	cluster.Resources.Store(m.Id(), m)

	assert.True(t, cluster.Contains(m.Id()), "Cluster should contains the manifest before delete")
	assert.NoError(t, m.Delete(context.Background(), konfig.RESTMapper, client, metav1.DeleteOptions{}))

	actual, found := cluster.Resources.Load(m.Id())
	assert.True(t, found, "Cluster should still contains the manifest after abandon")
	mactual := actual.(Manifest)
	assert.Equal(t, m.Id(), mactual.Id())
	assert.Equal(t, "", mactual.GetInstance())
}

func TestManifestPatch(t *testing.T) {
	konfig, tfake, client := utils.NewFakeK8sConfig()
	coreresources := &metav1.APIResourceList{
		GroupVersion: "testing",
		APIResources: []metav1.APIResource{
			{
				Name:  "manifests",
				Kind:  "Manifest",
				Verbs: metav1.Verbs{"delete, create, patch, get"},
			},
		},
	}
	tfake.Resources = append(tfake.Resources, coreresources)
	cluster := mock.NewCluster(client, tfake.Resources...)

	m := NewUnique()
	m.SetAPIVersion("testing")
	m.SetKind("Foo")

	_, err := m.Patch(context.Background(), konfig.RESTMapper, client, metav1.PatchOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not get resource interface:")

	m.SetKind("Manifest")
	assert.False(t, cluster.Contains(m.Id()), "Cluster should not contains the manifest before patch")
	expected, err := m.Patch(context.Background(), konfig.RESTMapper, client, metav1.PatchOptions{})
	assert.NoError(t, err)
	actual, found := cluster.Resources.Load(m.Id())
	assert.True(t, found, "Cluster should contains the manifest after patch")
	assert.Equal(t, expected, New(actual.(*unstructured.Unstructured)))

	// Fail marshal
	m.SetUnstructuredContent(map[string]interface{}{"foo": math.NaN()})
	m.SetAPIVersion("testing")
	m.SetKind("Manifest")
	_, err = m.Patch(context.Background(), konfig.RESTMapper, client, metav1.PatchOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to marshal object:")
}
