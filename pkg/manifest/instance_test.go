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
package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetInstance(t *testing.T) {
	u := new(unstructured.Unstructured)
	u.SetLabels(map[string]string{
		InstanceLabel: "my-instance",
	})
	m := New(u)
	assert.Equal(t, "my-instance", m.GetInstance())
}

func TestGetDeletionPolicy(t *testing.T) {
	u := new(unstructured.Unstructured)
	u.SetAnnotations(map[string]string{
		DeletionPolicyAnnotation: DeletionPolicyAbandon,
	})
	m := New(u)
	assert.Equal(t, DeletionPolicyAbandon, m.GetDeletionPolicy())
}

func TestWithInstance(t *testing.T) {
	u := new(unstructured.Unstructured)
	m := New(u)
	m2 := m.WithInstance("potato")

	assert.Equal(t, m, m2)
	assert.Equal(t, "potato", m.GetInstance())
	assert.Equal(t, "potato", m2.GetInstance())
}
