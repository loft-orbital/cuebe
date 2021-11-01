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
package kubernetes

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestRender(t *testing.T) {
	r := &Release{
		Objects: []unstructured.Unstructured{
			(&raw{v1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"}, v1.ObjectMeta{Name: "cuebe-test"}}).ToUnstructured(),
			(&raw{v1.TypeMeta{Kind: "Namespace", APIVersion: "v1"}, v1.ObjectMeta{Name: "default"}}).ToUnstructured(),
		},
	}

	b := new(strings.Builder)
	assert.NoError(t, r.Render(b))
	expected := `---
apiVersion: v1
kind: ConfigMap
metadata:
  creationTimestamp: null
  name: cuebe-test
---
apiVersion: v1
kind: Namespace
metadata:
  creationTimestamp: null
  name: default
`
	assert.Equal(t, expected, b.String())
}
