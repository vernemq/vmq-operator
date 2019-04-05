package vernemq

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"reflect"
	"text/template"

	"github.com/go-logr/logr"
	vernemqv1alpha1 "github.com/vernemq/vmq-operator/pkg/apis/vernemq/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/blang/semver"
	pkgerr "github.com/pkg/errors"
)

// Constants for VerneMQ StatefulSet & Volumes
const (
	storageDir        = "/vernemq/data"
	configmapsDir     = "/vernemq/etc/configmaps"
	configFilename    = "vernemq.yaml.gz"
	sSetInputHashName = "vernemq-operator-input-hash"

	defaultVerneMQVersion   = "1.7.1-2-alpine"
	defaultVerneMQBaseImage = "vernemq/vernemq"
)

var (
	minSize             int32 = 1
	probeTimeoutSeconds int32 = 3
)

var log = logf.Log.WithName("controller_vernemq")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new VerneMQ Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))

}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileVerneMQ{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("vernemq-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource VerneMQ
	err = c.Watch(&source.Kind{Type: &vernemqv1alpha1.VerneMQ{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner VerneMQ
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &vernemqv1alpha1.VerneMQ{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileVerneMQ{}

// ReconcileVerneMQ reconciles a VerneMQ object
type ReconcileVerneMQ struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	logger logr.Logger
}

// Reconcile reads that state of the cluster for a VerneMQ object and makes changes based on the state read
// and what is in the VerneMQ.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileVerneMQ) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	r.logger = reqLogger
	reqLogger.Info("Reconciling VerneMQ")

	// Fetch the VerneMQ instance
	instance := &vernemqv1alpha1.VerneMQ{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Create empty Secret if it doesn't exist. See comment above.
	secret := makeEmptyConfigurationSecret(instance)
	err = r.createOrUpdate(secret.Name, secret.Namespace, secret)
	if err != nil {
		return reconcile.Result{}, pkgerr.Wrap(err, "creating empty config secret failed")
	}

	service := makeStatefulSetService(instance)
	err = r.createOrUpdate(service.Name, service.Namespace, service)
	if err != nil {
		return reconcile.Result{}, pkgerr.Wrap(err, "generating service failed")
	}
	statefulset, err := makeStatefulSet(instance)
	if err != nil {
		return reconcile.Result{}, pkgerr.Wrap(err, "generating statefulset failed")
	}
	err = r.createOrUpdate(statefulset.Name, statefulset.Namespace, statefulset)
	if err != nil {
		return reconcile.Result{}, pkgerr.Wrap(err, "creating statefulset failed")
	}

	podList, err := r.listPods(instance.Name, instance.Namespace)
	if err != nil {
		return reconcile.Result{}, pkgerr.Wrap(err, "listing pods failed")
	}
	clusterViewSecret := makeClusterViewSecret(instance, podList)
	err = r.createOrUpdate(clusterViewSecret.Name, clusterViewSecret.Namespace, clusterViewSecret)
	if err != nil {
		return reconcile.Result{}, pkgerr.Wrap(err, "creating clusterview secret failed")
	}

	return reconcile.Result{Requeue: true}, nil
}

func (r *ReconcileVerneMQ) listPods(name string, namespace string) (*corev1.PodList, error) {
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForVerneMQ(name))
	listOps := &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: labelSelector,
	}
	err := r.client.List(context.TODO(), listOps, podList)

	if err != nil {
		return podList, pkgerr.Wrap(err, "listing pods failed")
	}
	return podList, nil
}

func (r *ReconcileVerneMQ) createOrUpdate(name string, namespace string, object runtime.Object) error {

	found := object.DeepCopyObject()
	key := types.NamespacedName{Name: name, Namespace: namespace}
	err := r.client.Get(context.TODO(), key, found)
	if err != nil && errors.IsNotFound(err) {
		// define a new resource
		err = r.client.Create(context.TODO(), object)
		if err != nil {
			return pkgerr.Wrap(err, "failed to create object")
		}
		r.logger.Info("created", "object", reflect.TypeOf(object))
		return nil
	} else if err != nil {
		return pkgerr.Wrap(err, "failed to retrieve object")
	} else {
		a := meta.NewAccessor()
		resourceVersion, err := a.ResourceVersion(found)
		if err != nil {
			return pkgerr.Wrap(err, "coudln't extract resource version of object")
		}
		err = a.SetResourceVersion(object, resourceVersion)
		if err != nil {
			return pkgerr.Wrap(err, "coudln't set resource version on object")
		}
		err = r.client.Update(context.TODO(), object)
		if err != nil {
			return pkgerr.Wrap(err, "failed to update object")
		}
		r.logger.Info("updated", "object", reflect.TypeOf(object))
		return nil
	}
}

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

func labelsForVerneMQ(name string) map[string]string {
	return map[string]string{"app": "vernemq", "vernemq": name}
}

func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

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

func makeStatefulSetSpec(instance *vernemqv1alpha1.VerneMQ) (*appsv1.StatefulSetSpec, error) {
	// VerneMQ may take quite long to shut down to migrate existing
	// sessions. Allow up to 10 minutes for clean termination
	terminationGracePeriod := int64(600)

	version, err := semver.Parse(instance.Spec.Version)
	if err != nil {
		return nil, pkgerr.Wrap(err, "parse version")
	}

	vernemqCommand := []string{"/bin/sh", "-c", `
	mkdir -p plugins && \
	curl -L https://github.com/dergraf/downloads/raw/master/plugin.tar.gz | tar xvz -C plugins && \
	echo $VERNEMQ_CONF | base64 -d > /vernemq/etc/vernemq.conf && \
	echo $VM_ARGS | base64 -d > /vernemq/etc/vm.args && \
	echo "-name VerneMQ@$VERNEMQ_NODENAME" >> /vernemq/etc/vm.args && \
	cat /vernemq/etc/vm.args && \
	/vernemq/bin/vernemq console -noshell -noinput`}
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

	var livenessFailureThreshold int32 = 60

	var livenessProbeHandler = v1.Handler{
		Exec: &v1.ExecAction{
			Command: []string{"/bin/sh", "-c", "/vernemq/bin/vernemq ping | grep pong"},
		},
	}
	var readinessProbeHandler = v1.Handler{
		Exec: &v1.ExecAction{
			Command: []string{"/bin/sh", "-c", "/vernemq/bin/vernemq ping | grep pong"},
		},
	}

	var livenessProbe = &v1.Probe{
		Handler:          livenessProbeHandler,
		PeriodSeconds:    5,
		TimeoutSeconds:   probeTimeoutSeconds,
		FailureThreshold: livenessFailureThreshold,
	}
	var readinessProbe = &v1.Probe{
		Handler:          readinessProbeHandler,
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
	podLabels = labelsForVerneMQ(instance.Name)

	additionalContainers := instance.Spec.Containers

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
						Name:    "vernemq",
						Image:   vernemqImage,
						Ports:   ports,
						Command: vernemqCommand,
						//Args:           vernemqArgs,
						VolumeMounts:    vernemqVolumeMounts,
						LivenessProbe:   livenessProbe,
						ReadinessProbe:  readinessProbe,
						Resources:       instance.Spec.Resources,
						SecurityContext: &vmqContainerSecurityContext,
						Env: []v1.EnvVar{
							{
								Name:  "VERNEMQ_NODENAME",
								Value: "127.0.0.1",
								//ValueFrom: &v1.EnvVarSource{
								//	FieldRef: &v1.ObjectFieldSelector{
								//		FieldPath: "metadata.name",
								//	},
								//},
							},
							{
								Name:  "VMQ_CONFIGMAP",
								Value: fmt.Sprintf("%s/config.yaml", configmapsDir),
							},
							{
								Name:  "VERNEMQ_CONF",
								Value: makeGlobalVerneMQConf(instance),
							},
							{
								Name:  "VM_ARGS",
								Value: makeGlobalVMArgs(instance),
							},
						},
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

func configSecretName(name string) string {
	return prefixedName(name)
}

func clusterViewSecretName(name string) string {
	return fmt.Sprintf("%s-clusterview", prefixedName(name))
}

func volumeName(name string) string {
	return fmt.Sprintf("%s-db", prefixedName(name))
}

func serviceName(name string) string {
	return fmt.Sprintf("%s-service", prefixedName(name))
}

func prefixedName(name string) string {
	return fmt.Sprintf("vernemq-%s", name)
}

func subPathForStorage(s *vernemqv1alpha1.StorageSpec) string {
	if s == nil {
		return ""
	}
	return "vernemq-db"
}
func makeEmptyConfigurationSecret(instance *vernemqv1alpha1.VerneMQ) *v1.Secret {
	s := makeConfigSecret(instance)
	s.Namespace = instance.Namespace

	s.ObjectMeta.Annotations = map[string]string{
		"empty": "true",
	}

	return s
}

func makeConfigSecret(instance *vernemqv1alpha1.VerneMQ) *v1.Secret {
	boolTrue := true
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   configSecretName(instance.Name),
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
		StringData: map[string]string{},
	}
}

func makeClusterViewSecret(instance *vernemqv1alpha1.VerneMQ, podList *corev1.PodList) *v1.Secret {
	str := ""
	clusterName := instance.ClusterName
	if clusterName == "" {
		clusterName = "cluster.local"
	}

	for _, pod := range podList.Items {
		str += fmt.Sprintf("%s.%s.%s.svc.%s;", pod.Spec.Hostname, serviceName(instance.Name), instance.Namespace, clusterName)
	}
	boolTrue := true
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vernemq-clusterview",
			Namespace: instance.Namespace,
			Labels:    labelsForVerneMQ(instance.Name),
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
		StringData: map[string]string{"vernemq.clusterview": str},
	}
}

func makeGlobalVerneMQConf(instance *vernemqv1alpha1.VerneMQ) string {
	// Static configuration that can't be changed on runtime
	// belongs here:
	config := `metadata_plugin = vmq_swc
plugins.vmq_passwd = off
plugins.vmq_acl = off
plugins.vmq_k8s.path = /vernemq/plugins/_build/default
plugins.vmq_k8s = on
leveldb.maximum_memory.percent = 20`
	return base64.StdEncoding.EncodeToString([]byte(config))
}

func makeGlobalVMArgs(instance *vernemqv1alpha1.VerneMQ) string {
	// -name is added by start script
	tmpl, err := template.New("config").Parse(
		`+P 256000
-env ERL_MAX_ETS_TABLES 256000
-env ERL_CRASH_DUMP /var/log/vernemq/erl_crash.dump
-env ERL_FULLSWEEP_AFTER 0
-env ERL_MAX_PORTS 262144
+A 64
-setcookie {{.Cookie}}
+K true
+W w
-smp enable
`)
	if err != nil {
		panic(err)
	}
	var res bytes.Buffer
	err = tmpl.Execute(&res, struct {
		Cookie string
	}{
		Cookie: "vmq",
	})
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString([]byte(res.String()))
}
