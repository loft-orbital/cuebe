package manifest

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func (m Manifest) Decode() (runtime.Object, error) {
	b, err := m.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("decoding manifest: %w", err)
	}

	obj, _, err := unstructured.UnstructuredJSONScheme.Decode(b, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("decoding manifest: %w", err)
	}

	return obj, nil
}
