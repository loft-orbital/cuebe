package instance

import (
	"context"

	"github.com/loft-orbital/cuebe/internal/utils"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	instanceDefinition = extv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "instances.cuebe.loftorbital.com",
		},

		Spec: extv1.CustomResourceDefinitionSpec{
			Group: "cuebe.loftorbital.com",
			Scope: extv1.ClusterScoped,
			Names: extv1.CustomResourceDefinitionNames{
				Plural:     "instances",
				Singular:   "instance",
				Kind:       "Instance",
				ShortNames: []string{"inst"},
			},
			Versions: []extv1.CustomResourceDefinitionVersion{v1alpha1},
		},
	}

	v1alpha1 = extv1.CustomResourceDefinitionVersion{
		Name:    "v1alpha1",
		Served:  true,
		Storage: true,
		Schema: &extv1.CustomResourceValidation{
			OpenAPIV3Schema: &extv1.JSONSchemaProps{
				Type: "object",
				Properties: map[string]extv1.JSONSchemaProps{
					"spec": {
						Type: "object",
						Properties: map[string]extv1.JSONSchemaProps{
							"resources": {
								Type: "array",
								Items: &extv1.JSONSchemaPropsOrArray{
									Schema: &extv1.JSONSchemaProps{
										Type:     "object",
										Required: []string{"group", "version", "kind", "name"},
										Properties: map[string]extv1.JSONSchemaProps{
											"group":     {Type: "string"},
											"version":   {Type: "string"},
											"kind":      {Type: "string"},
											"namespace": {Type: "string"},
											"name":      {Type: "string"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
)

// InstallCRD install the cuebe instance custom definition to the cluster.
func InstallCRD(ctx context.Context, config *utils.K8sConfig, opts utils.CommonMetaOptions) error {
	resource := config.ExtensionClient.ApiextensionsV1().CustomResourceDefinitions()
	_, err := resource.Create(ctx, &instanceDefinition, opts.CreateOptions())
	return err
}
