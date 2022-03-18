package vernemq

import (
	"fmt"

	vernemqv1alpha1 "github.com/vernemq/vmq-operator/pkg/apis/vernemq/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makeDeployment(instance *vernemqv1alpha1.VerneMQ) *appsv1.Deployment {
	boolTrue := true
	spec := makeDeploymentSpec(instance)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        deploymentName(instance.Name),
			Namespace:   instance.Namespace,
			Annotations: instance.ObjectMeta.Annotations,
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
		Spec: *spec,
	}
}

func makeDeploymentSpec(instance *vernemqv1alpha1.VerneMQ) *appsv1.DeploymentSpec {
	if instance.Spec.BundlerBaseImage == "" {
		instance.Spec.BundlerBaseImage = defaultBundlerBaseImage
	}
	if instance.Spec.BundlerVersion == "" {
		instance.Spec.BundlerVersion = defaultBundlerVersion
	}
	bundlerImage := fmt.Sprintf("%s:%s", instance.Spec.BundlerBaseImage, instance.Spec.BundlerVersion)
	if instance.Spec.BundlerTag != "" {
		bundlerImage = fmt.Sprintf("%s:%s", instance.Spec.BundlerBaseImage, instance.Spec.BundlerTag)
	}
	if instance.Spec.BundlerSHA != "" {
		bundlerImage = fmt.Sprintf("%s@sha256:%s", instance.Spec.BundlerBaseImage, instance.Spec.BundlerSHA)
	}
	if instance.Spec.BundlerImage != nil && *instance.Spec.BundlerImage != "" {
		bundlerImage = *instance.Spec.BundlerImage
	}

	podLabels := map[string]string{"app": "vmq-bundler"}
	podAnnotations := map[string]string{}

	return &appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{
			MatchLabels: podLabels,
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      podLabels,
				Annotations: podAnnotations,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "vmq-bundler",
						Image: bundlerImage,
						Ports: []v1.ContainerPort{
							{
								Name:          "http",
								ContainerPort: 80,
								Protocol:      v1.ProtocolTCP,
							},
						},
						Env: []v1.EnvVar{
							{
								Name:  "BUNDLER_CONFIG",
								Value: makeBundlerConfig(instance),
							},
							{
								Name:  "HTTP_PORT",
								Value: "80",
							},
						},
					},
				},
			},
		},
	}
}

func makeBundlerConfig(instance *vernemqv1alpha1.VerneMQ) string {
	config := `
{plugins, [
	{rebar3_cargo, {git, "https://github.com/benoitc/rebar3_cargo", {ref, "379115f"}}}
]}.
{deps, [
	`
	for _, p := range instance.Spec.ExternalPlugins {
		config = config + fmt.Sprintf("{%s, {git, \"%s\", {%s, \"%s\"}}},\n", p.ApplicationName, p.RepoURL, p.VersionType, p.Version)
	}
	config = config + `
	{vmq_k8s, {git, "https://github.com/vernemq/vmq-operator", {branch, "master"}}}
]}.
	`
	return config
}
