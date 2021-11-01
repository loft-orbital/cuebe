package kubernetes

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

const defaultNamespace = "default"

// SortableUnstructured implements sort.Interface based on the namespace first, then the name of an unstructured object.
type SortableUnstructured []unstructured.Unstructured

func (u SortableUnstructured) Len() int { return len(u) }

func (u SortableUnstructured) Less(i, j int) bool {
	ik := u[i].GetKind()
	jk := u[j].GetKind()
	if ik == "Namespace" {
		return true
	}
	if jk == "Namespace" {
		return false
	}

	if ik == "CustomResourceDefinition" {
		return true
	}
	if jk == "CustomResourceDefinition" {
		return false
	}

	nsc := strings.Compare(u[i].GetNamespace(), u[j].GetNamespace())
	if nsc != 0 {
		return nsc < 0
	}

	return strings.Compare(u[i].GetName(), u[j].GetName()) < 0
}

func (u SortableUnstructured) Swap(i, j int) { u[i], u[j] = u[j], u[i] }

func dynamicResourceInterfaceFor(obj unstructured.Unstructured, rm *restmapper.DeferredDiscoveryRESTMapper, dc dynamic.Interface) (dynamic.ResourceInterface, error) {
	// Get mapping
	gvk := obj.GroupVersionKind()
	mapping, err := rm.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		// Namespaced resources
		ns := obj.GetNamespace()
		if ns == "" {
			ns = defaultNamespace
		}
		dr = dc.Resource(mapping.Resource).Namespace(ns)
	} else {
		// Cluster-wide resources
		dr = dc.Resource(mapping.Resource)
	}
	return dr, nil
}

func printObj(obj unstructured.Unstructured) string {
	return strings.ToLower(fmt.Sprintf("%s/%s", obj.GetKind(), obj.GetName()))
}
