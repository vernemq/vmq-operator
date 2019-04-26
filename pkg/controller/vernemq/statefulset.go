package vernemq

import (
	"encoding/base64"
	"fmt"

	"github.com/blang/semver"
	pkgerr "github.com/pkg/errors"
	vernemqv1alpha1 "github.com/vernemq/vmq-operator/pkg/apis/vernemq/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// statefulSetForVerneMQ returns a VerneMQ StatefulSet object
func makeStatefulSet(instance *vernemqv1alpha1.VerneMQ) (*appsv1.StatefulSet, error) {

	// instance is passed in by value, not by reference. But p contains references like
	// to annotation map, that do not get copied on function invocation. Ensure to
	// prevent side effects before editing p by creating a deep copy. For more
	// details see https://github.com/coreos/prometheus-operator/issues/1659.
	instance = instance.DeepCopy()

	if instance.Spec.BaseImage == "" {
		instance.Spec.BaseImage = defaultVerneMQBaseImage
	}

	if instance.Spec.Version == "" {
		instance.Spec.Version = defaultVerneMQVersion
	}

	if instance.Spec.Size == nil {
		instance.Spec.Size = &minSize
	}
	intZero := int32(0)
	if instance.Spec.Size != nil && *instance.Spec.Size < 0 {
		instance.Spec.Size = &intZero
	}

	spec, err := makeStatefulSetSpec(instance)

	if err != nil {
		return nil, pkgerr.Wrap(err, "make StatefulSet spec")
	}

	boolTrue := true
	statefulset := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        prefixedName(instance.Name),
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

	//if statefulset.ObjectMeta.Annotations == nil {
	//	statefulset.ObjectMeta.Annotations = map[string]string{
	//		sSetInputHashName: inputHash,
	//	}
	//} else {
	//	statefulset.ObjectMeta.Annotations[sSetInputHashName] = inputHash

	//}

	if instance.Spec.ImagePullSecrets != nil && len(instance.Spec.ImagePullSecrets) > 0 {
		statefulset.Spec.Template.Spec.ImagePullSecrets = instance.Spec.ImagePullSecrets
	}

	storageSpec := instance.Spec.Storage
	if storageSpec == nil {
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(instance.Name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		})
	} else if storageSpec.EmptyDir != nil {
		emptyDir := storageSpec.EmptyDir
		statefulset.Spec.Template.Spec.Volumes = append(statefulset.Spec.Template.Spec.Volumes, v1.Volume{
			Name: volumeName(instance.Name),
			VolumeSource: v1.VolumeSource{
				EmptyDir: emptyDir,
			},
		})
	} else {
		pvcTemplate := storageSpec.VolumeClaimTemplate
		if pvcTemplate.Name == "" {
			pvcTemplate.Name = volumeName(instance.Name)
		}
		pvcTemplate.Spec.AccessModes = []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce}
		pvcTemplate.Spec.Resources = storageSpec.VolumeClaimTemplate.Spec.Resources
		pvcTemplate.Spec.Selector = storageSpec.VolumeClaimTemplate.Spec.Selector
		statefulset.Spec.VolumeClaimTemplates = append(statefulset.Spec.VolumeClaimTemplates, pvcTemplate)
	}
	return statefulset, nil
}

func makeStatefulSetSpec(instance *vernemqv1alpha1.VerneMQ) (*appsv1.StatefulSetSpec, error) {
	// VerneMQ may take quite long to shut down to migrate existing
	// sessions. Allow up to 10 minutes for clean termination

	dropoutPeriod := int64(0)
	if instance.Spec.DropoutPeriodSeconds != nil {
		dropoutPeriod = *instance.Spec.DropoutPeriodSeconds
	}

	terminationGracePeriod := int64(3600)
	if instance.Spec.TerminationGracePeriodSeconds != nil {
		terminationGracePeriod = *instance.Spec.TerminationGracePeriodSeconds + dropoutPeriod
	}

	version, err := semver.Parse(instance.Spec.Version)
	if err != nil {
		return nil, pkgerr.Wrap(err, "parse version")
	}

	vernemqCommand := []string{"/bin/sh", "-c", `
	mkdir -p plugins && \
	curl -L http://$VMQ_BUNDLER_HOST/bundle.tar.gz | tar xvz -C plugins && \
	eval "echo \"$(echo $VERNEMQ_CONF | base64 -d)\"" > /vernemq/etc/vernemq.conf && \
	eval "echo \"$(echo $VM_ARGS | base64 -d)\"" > /vernemq/etc/vm.args && \
	/vernemq/bin/vernemq console -noshell -noinput`}

	vernemqPreStopCommand := []string{"/bin/sh", "-c", fmt.Sprintf(`
	vmq-admin cluster leave node=vmq@$VMQ_NODENAME.$VMQ_HOSTNAME && \
	sleep %d && \
	vmq-admin cluster leave node=vmq@$VMQ_NODENAME.$VMQ_HOSTNAME --timeout=%d --kill_sessions
	`, dropoutPeriod, terminationGracePeriod)}
	//vernemqCommand := []string{
	//	"/vernemq/bin/vernemq",
	//}
	//vernemqArgs := []string{
	//	"console",
	//	"-noshell",
	//	"-noinput",
	//}
	switch version.Major {
	case 1:
		if version.Minor < 7 {
			return nil, pkgerr.Errorf("unsupported VerneMQ minor version %s", version)
		}
	default:
		return nil, pkgerr.Errorf("unsupported VerneMQ major version %s", version)
	}

	var securityContext *v1.PodSecurityContext
	if instance.Spec.SecurityContext != nil {
		securityContext = instance.Spec.SecurityContext
	}

	var ports = []v1.ContainerPort{
		{
			Name:          "mqtt",
			ContainerPort: 1883,
			Protocol:      v1.ProtocolTCP,
		},
		{
			Name:          "mqtts",
			ContainerPort: 8883,
			Protocol:      v1.ProtocolTCP,
		},
		{
			Name:          "epmd",
			ContainerPort: 4369,
			Protocol:      v1.ProtocolTCP,
		},
		{
			Name:          "vmq-cluster",
			ContainerPort: 44053,
			Protocol:      v1.ProtocolTCP,
		},
		{
			Name:          "mqtt-ws",
			ContainerPort: 8080,
			Protocol:      v1.ProtocolTCP,
		},
		{
			Name:          "http",
			ContainerPort: 8888,
			Protocol:      v1.ProtocolTCP,
		},
	}
	epmdPortRange := []int32{9100, 9101, 9102, 9103, 9104, 9105, 9106, 9107, 9108, 9109}
	for _, port := range epmdPortRange {
		ports = append(ports, v1.ContainerPort{
			ContainerPort: port,
			Protocol:      v1.ProtocolTCP,
		})
	}

	volumes := []v1.Volume{
		{
			Name: "vernemq-conf",
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: "vernemq-conf",
					},
					Items: []v1.KeyToPath{
						{
							Key:  "config.yaml",
							Path: "config.yaml",
						},
					},
				},
			},
		},
		{
			Name: "vernemq-clusterview",
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: "vernemq-clusterview",
					Items: []v1.KeyToPath{
						{
							Key:  "vernemq.clusterview",
							Path: "vernemq.clusterview",
						},
					},
				},
			},
		},
	}

	volName := volumeName(instance.Name)
	if instance.Spec.Storage != nil {
		if instance.Spec.Storage.VolumeClaimTemplate.Name != "" {
			volName = instance.Spec.Storage.VolumeClaimTemplate.Name
		}
	}

	vernemqVolumeMounts := []v1.VolumeMount{
		{
			Name:      volName,
			MountPath: storageDir,
			SubPath:   subPathForStorage(instance.Spec.Storage),
		},
		{
			Name:      "vernemq-conf",
			MountPath: configmapsDir,
		},
		{
			Name: "vernemq-clusterview",
			//MountPath: configmapsDir,
			MountPath: fmt.Sprintf("%s/clusterview", configmapsDir),
		},
	}

	for _, s := range instance.Spec.Secrets {
		volumes = append(volumes, v1.Volume{
			Name: volumeName("secret-" + s),
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: s,
				},
			},
		})
		vernemqVolumeMounts = append(vernemqVolumeMounts, v1.VolumeMount{
			Name:      volumeName("secret-" + s),
			ReadOnly:  true,
			MountPath: secretsDir + s,
		})
	}

	for _, c := range instance.Spec.ConfigMaps {
		volumes = append(volumes, v1.Volume{
			Name: volumeName("configmap-" + c),
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: c,
					},
				},
			},
		})
		vernemqVolumeMounts = append(vernemqVolumeMounts, v1.VolumeMount{
			Name:      volumeName("configmap-" + c),
			ReadOnly:  true,
			MountPath: configmapsDir + c,
		})
	}

	var livenessFailureThreshold int32 = 60

	var probeHandler = v1.Handler{
		HTTPGet: &v1.HTTPGetAction{
			Path: "/health",
			Port: intstr.FromInt(8888),
		},
	}

	var livenessProbe = &v1.Probe{
		Handler:          probeHandler,
		PeriodSeconds:    5,
		TimeoutSeconds:   probeTimeoutSeconds,
		FailureThreshold: livenessFailureThreshold,
	}
	var readinessProbe = &v1.Probe{
		Handler:          probeHandler,
		PeriodSeconds:    5,
		TimeoutSeconds:   probeTimeoutSeconds,
		FailureThreshold: 120, // allow up to 10m on startup for data recovery
	}

	podLabels := map[string]string{}
	podAnnotations := map[string]string{}
	if instance.Spec.PodMetadata != nil {
		if instance.Spec.PodMetadata.Labels != nil {
			for k, v := range instance.Spec.PodMetadata.Labels {
				podLabels[k] = v
			}
		}
		if instance.Spec.PodMetadata.Annotations != nil {
			for k, v := range instance.Spec.PodMetadata.Annotations {
				podAnnotations[k] = v
			}
		}
	}
	podLabels["app"] = "vernemq"
	podLabels["vernemq"] = instance.Name

	vernemqImage := fmt.Sprintf("%s:%s", instance.Spec.BaseImage, instance.Spec.Version)
	if instance.Spec.Tag != "" {
		vernemqImage = fmt.Sprintf("%s:%s", instance.Spec.BaseImage, instance.Spec.Tag)
	}
	if instance.Spec.SHA != "" {
		vernemqImage = fmt.Sprintf("%s@sha256:%s", instance.Spec.BaseImage, instance.Spec.SHA)
	}
	if instance.Spec.Image != nil && *instance.Spec.Image != "" {
		vernemqImage = *instance.Spec.Image
	}

	UID := int64(10000)
	vmqContainerSecurityContext := v1.SecurityContext{
		RunAsUser:  &UID,
		RunAsGroup: &UID,
	}

	additionalContainers := instance.Spec.Containers
	envVars := instance.Spec.Env

	return &appsv1.StatefulSetSpec{
		ServiceName:         serviceName(instance.Name),
		Replicas:            instance.Spec.Size,
		PodManagementPolicy: appsv1.ParallelPodManagement,
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
		},
		Selector: &metav1.LabelSelector{
			MatchLabels: podLabels,
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      podLabels,
				Annotations: podAnnotations,
			},
			Spec: v1.PodSpec{
				Containers: append([]v1.Container{
					{
						Name:            vernemqName,
						Image:           vernemqImage,
						Ports:           ports,
						Command:         vernemqCommand,
						VolumeMounts:    vernemqVolumeMounts,
						LivenessProbe:   livenessProbe,
						ReadinessProbe:  readinessProbe,
						Resources:       instance.Spec.Resources,
						SecurityContext: &vmqContainerSecurityContext,
						Lifecycle: &v1.Lifecycle{
							PreStop: &v1.Handler{
								Exec: &v1.ExecAction{
									Command: vernemqPreStopCommand,
								},
							},
						},
						Env: append([]v1.EnvVar{
							{
								Name: "VMQ_NODENAME",
								ValueFrom: &v1.EnvVarSource{
									FieldRef: &v1.ObjectFieldSelector{
										FieldPath: "metadata.name",
									},
								},
							},
							{
								Name:  "VMQ_HOSTNAME",
								Value: getHostname(instance),
							},
							{
								Name:  "VMQ_CONFIGMAP",
								Value: fmt.Sprintf("%s/config.yaml", configmapsDir),
							},
							{
								Name:  "VMQ_CLUSTERVIEW",
								Value: fmt.Sprintf("%s/clusterview/vernemq.clusterview", configmapsDir),
							},
							{
								Name:  "VMQ_BUNDLER_HOST",
								Value: bundlerServiceName(instance.Name),
							},
							{
								Name:  "VERNEMQ_CONF",
								Value: makeGlobalVerneMQConf(instance),
							},
							{
								Name:  "VM_ARGS",
								Value: makeGlobalVMArgs(instance),
							},
							{
								Name: "ERLANG_SCHEDULERS",
								ValueFrom: &v1.EnvVarSource{
									ResourceFieldRef: &v1.ResourceFieldSelector{
										ContainerName: vernemqName,
										Resource:      "requests.cpu",
									},
								},
							},
							{
								Name: "MY_POD_IP",
								ValueFrom: &v1.EnvVarSource{
									FieldRef: &v1.ObjectFieldSelector{
										FieldPath: "status.podIP",
									},
								},
							},
						}, envVars...),
					},
				}, additionalContainers...),
				SecurityContext:               securityContext,
				ServiceAccountName:            instance.Spec.ServiceAccountName,
				NodeSelector:                  instance.Spec.NodeSelector,
				PriorityClassName:             instance.Spec.PriorityClassName,
				TerminationGracePeriodSeconds: &terminationGracePeriod,
				Volumes:                       volumes,
				Tolerations:                   instance.Spec.Tolerations,
				Affinity:                      instance.Spec.Affinity,
			},
		},
	}, nil
}

func makeGlobalVerneMQConf(instance *vernemqv1alpha1.VerneMQ) string {
	// Static configuration that can't be changed on runtime
	// belongs here:
	config := `metadata_plugin = vmq_swc
listener.vmq.clustering = $MY_POD_IP:44053
listener.http.default = $MY_POD_IP:8888
plugins.vmq_passwd = off
plugins.vmq_acl = off
plugins.vmq_k8s.path = /vernemq/plugins/_build/default
plugins.vmq_k8s = on
leveldb.maximum_memory.percent = 20
log.console = console
`
	config = config + instance.Spec.VMQConfig + "\n"
	fmt.Printf(config)
	return base64.StdEncoding.EncodeToString([]byte(config))
}
func makeGlobalVMArgs(instance *vernemqv1alpha1.VerneMQ) string {
	// -name is added by start script
	vmArgs := `+P 256000
-env ERL_MAX_ETS_TABLES 256000
-env ERL_CRASH_DUMP /vernemq/log/erl_crash.dump
-env ERL_FULLSWEEP_AFTER 0
-env ERL_MAX_PORTS 262144
+A 64
-setcookie ${VMQ_DISTRIBUTED_COOKIE:-vmq}
-name vmq@$VMQ_NODENAME.$VMQ_HOSTNAME
+K true
+W w
-smp enable
`
	vmArgs = vmArgs + instance.Spec.VMArgs + "\n"
	fmt.Printf(vmArgs)
	return base64.StdEncoding.EncodeToString([]byte(vmArgs))
}
