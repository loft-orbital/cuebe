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

const (
	DeletionPolicyAbandon    = "abandon"
	InstanceLabel            = "cuebe.loft-orbital.com/instance"
	DeletionPolicyAnnotation = "cuebe.loft-orbital.com/deletion-policy"
)

// GetInstance retrieve the instance this Manifest belongs to,
// or an empty string if not belonging to any instance.
func (m Manifest) GetInstance() string {
	return m.GetLabels()[InstanceLabel]
}

// GetDeletionPolicy returns the deletion policy of this Manifest.
func (m Manifest) GetDeletionPolicy() string {
	return m.GetAnnotations()[DeletionPolicyAnnotation]
}

// WithInstance sets the manifest instance and returns the modified manifest.
func (m Manifest) WithInstance(name string) Manifest {
	labels := m.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[InstanceLabel] = name
	m.SetLabels(labels)
	return m
}
