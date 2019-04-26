package vernemq

import (
	"fmt"

	vernemqv1alpha1 "github.com/vernemq/vmq-operator/pkg/apis/vernemq/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func makeStatefulSetService(instance *vernemqv1alpha1.VerneMQ) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: instance.APIVersion,
					Kind:       instance.Kind,
					Name:       instance.Name,
					UID:        instance.UID,
				},
			},
			Labels: map[string]string{
				"operated-vernemq": "true",
			},
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "None",
			Ports: []v1.ServicePort{
				{
					Name:       "mqtt",
					Port:       1883,
					TargetPort: intstr.FromString("mqtt"),
				},
				{
					Name:       "http",
					Port:       8888,
					TargetPort: intstr.FromString("http"),
				},
			},
			Selector: map[string]string{
				"app": "vernemq",
			},
		},
	}
	svc.Name = serviceName(instance.Name)
	svc.Namespace = instance.Namespace
	return svc
}

func makeDeploymentService(instance *vernemqv1alpha1.VerneMQ) *v1.Service {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: instance.APIVersion,
					Kind:       instance.Kind,
					Name:       instance.Name,
					UID:        instance.UID,
				},
			},
			Labels: map[string]string{
				"operated-vernemq": "true",
			},
		},
		Spec: v1.ServiceSpec{
			Type: "ClusterIP",
			Ports: []v1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromString("http"),
				},
			},
			Selector: map[string]string{
				"app": "vmq-bundler",
			},
		},
	}
	svc.Name = serviceName(instance.Name + "-vmq-bundler")
	svc.Namespace = instance.Namespace
	return svc
}

func serviceName(name string) string {
	return fmt.Sprintf("%s-service", prefixedName(name))
}

func bundlerServiceName(name string) string {
	return serviceName(name + "-vmq-bundler")
}
