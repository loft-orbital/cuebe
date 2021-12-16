package release

import (
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

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
