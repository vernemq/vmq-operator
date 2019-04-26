package vernemq

import (
	"context"
	"fmt"
	"reflect"

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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	pkgerr "github.com/pkg/errors"
)

// Constants for VerneMQ StatefulSet & Volumes
const (
	vernemqName       = "vernemq"
	storageDir        = "/vernemq/data"
	configmapsDir     = "/vernemq/etc/configmaps/"
	secretsDir        = "/vernemq/etc/secrets/"
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

	deploymentService := makeDeploymentService(instance)
	err = r.client.Create(context.TODO(), deploymentService)
	if err != nil && errors.IsAlreadyExists(err) == false {
		return reconcile.Result{}, pkgerr.Wrap(err, "generating deployment service failed")
	}

	deployment := makeDeployment(instance)
	err = r.createOrUpdate(deployment.Name, deployment.Namespace, deployment)
	if err != nil {
		return reconcile.Result{}, pkgerr.Wrap(err, "generating deployment failed")
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

	emptyConfig := makeEmptyVerneMQConfigMap(instance)
	err = r.client.Create(context.TODO(), emptyConfig)
	if err != nil && errors.IsAlreadyExists(err) == false {
		return reconcile.Result{}, pkgerr.Wrap(err, "creating empty config map failed")
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

func makeEmptyVerneMQConfigMap(instance *vernemqv1alpha1.VerneMQ) *v1.ConfigMap {
	boolTrue := true
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vernemq-conf",
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
		Data: map[string]string{"config.yaml": ""},
	}
}

func makeClusterViewSecret(instance *vernemqv1alpha1.VerneMQ, podList *corev1.PodList) *v1.Secret {
	str := ""
	for _, pod := range podList.Items {
		str += fmt.Sprintf("vmq@%s.%s;", pod.Spec.Hostname, getHostname(instance))
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
