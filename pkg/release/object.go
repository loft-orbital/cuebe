/*
Copyright © 2021 Loft Orbital

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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
