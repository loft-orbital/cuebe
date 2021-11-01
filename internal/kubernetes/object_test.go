package kubernetes

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type raw struct {
	v1.TypeMeta   `json:",inline"`
	v1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
}

func (r *raw) ToUnstructured() unstructured.Unstructured {
	u := unstructured.Unstructured{}
	b, _ := json.Marshal(r)
	json.Unmarshal(b, &u.Object)
	return u
}

func toSortableUnstructured(rs []raw) SortableUnstructured {
	s := []unstructured.Unstructured{}
	for _, r := range rs {
		s = append(s, r.ToUnstructured())
	}
	return SortableUnstructured(s)
}

func TestPrintObj(t *testing.T) {
	obj := &raw{v1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"}, v1.ObjectMeta{Name: "cuebe-test"}}

	assert.Equal(t, "configmap/cuebe-test", printObj(obj.ToUnstructured()))
}

func TestSortableUnstructuredLen(t *testing.T) {
	objs := make(SortableUnstructured, 2)
	assert.Equal(t, 2, len(objs))
}

func TestSortableUnstructuredSwap(t *testing.T) {
	first := (&raw{}).ToUnstructured()
	first.SetName("first")
	second := (&raw{}).ToUnstructured()
	second.SetName("second")
	l := SortableUnstructured{first, second}
	l.Swap(0, 1)

	assert.Equal(t, l[0], second)
	assert.Equal(t, l[1], first)
}

func TestSortableUnstructuredLess(t *testing.T) {
	tcs := map[string]struct {
		high *raw
		low  *raw
	}{
		"namespace vs crd": {
			high: &raw{v1.TypeMeta{Kind: "CustomResourceDefinition"}, v1.ObjectMeta{Namespace: "aaa"}},
			low:  &raw{v1.TypeMeta{Kind: "Namespace"}, v1.ObjectMeta{Namespace: "zzz"}},
		},
		"crd vs configmap": {
			high: &raw{v1.TypeMeta{Kind: "ConfigMap"}, v1.ObjectMeta{}},
			low:  &raw{v1.TypeMeta{Kind: "CustomResourceDefinition"}, v1.ObjectMeta{}},
		},
		"namespace": {
			high: &raw{v1.TypeMeta{}, v1.ObjectMeta{Namespace: "zzz"}},
			low:  &raw{v1.TypeMeta{}, v1.ObjectMeta{Namespace: "aaa"}},
		},
		"name": {
			high: &raw{v1.TypeMeta{}, v1.ObjectMeta{Name: "zzz"}},
			low:  &raw{v1.TypeMeta{}, v1.ObjectMeta{Name: "aaa"}},
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			l := SortableUnstructured{tc.low.ToUnstructured(), tc.high.ToUnstructured()}
			assert.True(t, l.Less(0, 1))
		})
	}
}
