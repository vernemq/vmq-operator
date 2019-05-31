package vernemq

import (
	"fmt"

	vernemqv1alpha1 "github.com/vernemq/vmq-operator/pkg/apis/vernemq/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func makeConfigSecretFromSpec(instance *vernemqv1alpha1.VerneMQ) *v1.Secret {
	boolTrue := true
	config := createStringData(instance)
	configSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "vernemq-yaml",
			Labels: labelsForVerneMQ(instance.Name),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         instance.APIVersion,
					BlockOwnerDeletion: &boolTrue,
					Controller:         &boolTrue,
					Kind:               instance.Kind,
					Name:               instance.Name,
					UID:                instance.UID,
				},
			},
		},
		Type:       "Opaque",
		StringData: map[string]string{"config.yaml": config},
	}
	configSecret.Namespace = instance.Namespace
	return configSecret
}

func createStringData(instance *vernemqv1alpha1.VerneMQ) string {
	d, err := yaml.Marshal(instance.Spec.Config)
	if err != nil {
		return ""
	}
	fmt.Printf("========================\n %s\n======================\n", string(d))
	return string(d)
}
