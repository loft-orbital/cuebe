package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

func TestPrintObj(t *testing.T) {
	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	deploymentYAML := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cuebe-test
`
	obj := &unstructured.Unstructured{}
	_, _, err := decUnstructured.Decode([]byte(deploymentYAML), nil, obj)
	require.NoError(t, err)

	assert.Equal(t, "configmap/cuebe-test", printObj(obj))
}
