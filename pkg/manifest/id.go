package manifest

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Id contains the necessary fields to identify a Manifest
type Id struct {
	Group     string `json:"group" protobuf:"bytes,1,opt,name=group"`
	Version   string `json:"version" protobuf:"bytes,2,opt,name=version"`
	Kind      string `json:"kind" protobuf:"bytes,3,opt,name=kind"`
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`
	Name      string `json:"name,omitempty" protobuf:"bytes,3,opt,name=name"`
}

// Id returns the Id of the Manifest m.
func (m Manifest) Id() Id {
	gvk := m.GroupVersionKind()

	return Id{
		Group:     gvk.Group,
		Version:   gvk.Version,
		Kind:      gvk.Kind,
		Namespace: m.GetNamespace(),
		Name:      m.GetName(),
	}
}

func (id Id) GroupKind() schema.GroupKind {
	return schema.GroupKind{
		Group: id.Group,
		Kind:  id.Kind,
	}
}

// Manifest retrieves the manifest matching this id in the given cluster.
func (id Id) Manifest(ctx context.Context, rm meta.RESTMapper, client dynamic.Interface, opts metav1.GetOptions) (Manifest, error) {
	resource, err := id.ResourceInterface(rm, client)
	if err != nil {
		return Manifest{}, fmt.Errorf("could not get resource interface: %w", err)
	}

	obj, err := resource.Get(ctx, id.Name, opts)
	if err != nil {
		return Manifest{}, fmt.Errorf("could not get resource: %w", err)
	}
	return New(obj), nil
}

// ResourceInterface returns the dynamic resource interface for this id.
func (id Id) ResourceInterface(rm meta.RESTMapper, client dynamic.Interface) (resource dynamic.ResourceInterface, err error) {
	mapping, err := id.RESTMapping(rm)
	if err != nil {
		return nil, fmt.Errorf("could not get rest mapping: %w", err)
	}

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		ns := id.Namespace
		if ns == "" {
			ns = "default" // TODO get actual default namespace
		}
		resource = client.Resource(mapping.Resource).Namespace(ns)
	} else {
		resource = client.Resource(mapping.Resource)
	}

	return
}

// RESTMapping returns the rest mapping for this id.
func (id Id) RESTMapping(rm meta.RESTMapper) (*meta.RESTMapping, error) {
	return rm.RESTMapping(id.GroupKind(), id.Version)
}

func (id Id) String() string {
	return fmt.Sprintf("%s/%s in %s", id.Kind, id.Name, id.Namespace)
}
