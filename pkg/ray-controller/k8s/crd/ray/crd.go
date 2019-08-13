package raycrd

import (
	"reflect"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ray-operator/pkg/ray-controller/k8s/apis/ray.io"
	"github.com/ray-operator/pkg/ray-controller/k8s/apis/ray.io/v1"
)

// CRD metadata.
const (
	Plural   = "rays"
	Singular = "ray"
	Group    = rayio.RayGroupName
	Version  = rayio.Version
	FullName = Plural + "." + Group
)

func GetCRD() *apiextensionsv1beta1.CustomResourceDefinition {
	return &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: FullName,
		},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   Group,
			Version: Version,
			Scope:   apiextensionsv1beta1.NamespaceScoped,
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   Plural,
				Singular: Singular,
				Kind:     reflect.TypeOf(v1.Ray{}).Name(),
			},
			Validation: getCustomResourceValidation(),
		},
	}
}

//todo ADD Validation
func getCustomResourceValidation() *apiextensionsv1beta1.CustomResourceValidation {
	return &apiextensionsv1beta1.CustomResourceValidation{}
}
