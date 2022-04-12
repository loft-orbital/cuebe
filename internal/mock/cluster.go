package mock

import (
	"fmt"
	"sync"

	"github.com/imdario/mergo"
	"github.com/loft-orbital/cuebe/pkg/manifest"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured/unstructuredscheme"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/testing"
)

type Cluster struct {
	Resources sync.Map

	GVRtoGVK map[schema.GroupVersionResource]schema.GroupVersionKind
}

func NewCluster(client *fake.FakeDynamicClient, resourcelist ...*metav1.APIResourceList) *Cluster {
	cluster := new(Cluster)

	cluster.GVRtoGVK = make(map[schema.GroupVersionResource]schema.GroupVersionKind)
	for _, rl := range resourcelist {
		gv, _ := schema.ParseGroupVersion(rl.GroupVersion)
		for _, r := range rl.APIResources {
			cluster.GVRtoGVK[gv.WithResource(r.Name)] = gv.WithKind(r.Kind)
		}
	}

	client.PrependReactor("delete", "*", cluster.ReactDelete)
	client.PrependReactor("get", "*", cluster.ReactGet)
	client.PrependReactor("patch", "*", cluster.ReactPatch)
	return cluster
}

func (c *Cluster) Contains(id manifest.Id) bool {
	_, ok := c.Resources.Load(id)
	return ok
}

type NamedAction interface {
	testing.Action
	GetName() string
}

func (c *Cluster) idFor(action NamedAction) manifest.Id {
	gvk := c.GVRtoGVK[action.GetResource()]
	return manifest.Id{
		Group:     gvk.Group,
		Kind:      gvk.Kind,
		Version:   gvk.Version,
		Namespace: action.GetNamespace(),
		Name:      action.GetName(),
	}
}

func (c *Cluster) ReactDelete(action testing.Action) (handled bool, ret runtime.Object, err error) {
	da, ok := action.(testing.DeleteAction)
	if !ok {
		return true, nil, errors.NewBadRequest("ReactDelete can only react to delete action")
	}

	rid := c.idFor(da)
	if _, ok := c.Resources.Load(rid); !ok {
		return true, nil, errors.NewNotFound(da.GetResource().GroupResource(), da.GetName())
	}

	c.Resources.Delete(rid)

	return true, nil, nil
}

func (c *Cluster) ReactGet(action testing.Action) (handled bool, ret runtime.Object, err error) {
	ga, ok := action.(testing.GetAction)
	if !ok {
		return true, nil, errors.NewBadRequest("ReactGet can only react to get action")
	}

	rid := c.idFor(ga)
	res, ok := c.Resources.Load(rid)
	if !ok {
		return true, nil, errors.NewNotFound(ga.GetResource().GroupResource(), ga.GetName())
	}

	ret, ok = res.(runtime.Object)
	if !ok {
		return true, nil, fmt.Errorf("cannot convert %v to runtime.Object", res)
	}

	return true, ret, nil
}

func (c *Cluster) ReactPatch(action testing.Action) (handled bool, ret runtime.Object, err error) {
	pa, ok := action.(testing.PatchAction)
	if !ok {
		return true, nil, errors.NewBadRequest("ReactPatch can only react to patch action")
	}

	rid := c.idFor(pa)

	// decode object
	srz := json.NewSerializerWithOptions(json.DefaultMetaFactory, unstructuredscheme.NewUnstructuredCreator(), unstructuredscheme.NewUnstructuredObjectTyper(), json.SerializerOptions{})
	gvk := rid.GroupKind().WithVersion(rid.Version)
	obj, _, err := srz.Decode(pa.GetPatch(), &gvk, nil)
	if err != nil {
		return true, nil, errors.NewInternalError(fmt.Errorf("could not decode object: %w", err))
	}

	// get current object if it exists
	r, exists := c.Resources.Load(rid)
	if !exists {
		r = obj
	}
	curr, ok := r.(runtime.Object)
	if exists && !ok {
		return true, nil, errors.NewInternalError(fmt.Errorf("cannot convert %v to runtime.Object", r))
	}

	// apply patch
	if err := mergo.MergeWithOverwrite(curr, obj); err != nil {
		return true, nil, errors.NewInternalError(fmt.Errorf("cannot patch: %s", err))
	}
	c.Resources.Store(rid, curr)

	return true, curr, nil
}
