package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VerneMQSpec defines the desired state of VerneMQ
// +k8s:openapi-gen=true
type VerneMQSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// Standard object’s metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	// Metadata Labels and Annotations gets propagated to the vernemq pods.
	PodMetadata *metav1.ObjectMeta `json:"podMetadata,omitempty"`
	// Size is the size of the VerneMQ deployment
	Size *int32 `json:"size,omitempty"`
	// Version of VerneMQ to be deployed
	Version string `json:"version,omitempty"`
	// Tag of VerneMQ container image to be deployed. Defaults to the value of `version`.
	// Version is ignored if Tag is set.
	Tag string `json:"tag,omitempty"`
	// SHA of VerneMQ container image to be deployed. Defaults to the value of `version`.
	// Similar to a tag, but the SHA explicitly deploys an immutable container image.
	// Version and Tag are ignored if SHA is set.
	SHA string `json:"sha,omitempty"`
	// Image if specified has precedence over baseImage, tag and sha
	// combinations. Specifying the version is still necessary to ensure the
	// VerneMQ Operator knows what version of VerneMQ is being
	// configured.
	Image *string `json:"image,omitempty"`
	// Base image to use for a VerneMQ deployment.
	BaseImage string `json:"baseImage,omitempty"`
	// An optional list of references to secrets in the same namespace
	// to use for pulling vernemq images from registries
	// see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
	// SecurityContext holds pod-level security attributes and common container settings.
	// This defaults to non root user with uid 10000 and gid 10000 for VerneMQ >1.7.0 and
	// default PodSecurityContext for other versions.
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`
	// Storage spec to specify how storage shall be used.
	Storage *StorageSpec `json:"storage,omitempty"`
	// Containers allows injecting additional containers.
	Containers []v1.Container `json:"containers,omitempty"`
	// Define resources requests and limits for single Pods.
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
	// ServiceAccountName is the name of the ServiceAccount to use to run the
	// VerneMQ Pods.
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Define which Nodes the Pods are scheduled on.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Priority class assigned to the Pods
	PriorityClassName string `json:"priorityClassName,omitempty"`
	// If specified, the pod's scheduling constraints.
	Affinity *v1.Affinity `json:"affinity,omitempty"`
	// If specified, the pod's tolerations.
	Tolerations                   []v1.Toleration `json:"tolerations,omitempty"`
	DropoutPeriodSeconds          *int64          `json:"dropoutPeriodSeconds,omitempty"`
	TerminationGracePeriodSeconds *int64          `json:"terminationGracePeriodSeconds,omitempty"`
	// Secrets is a list of Secrets in the same namespace as the VerneMQ
	// object, which shall be mounted into the VerneMQ Pods.
	// The Secrets are mounted into /etc/vernemq/secrets/<secret-name>.
	Secrets []string `json:"secrets,omitempty"`
	// ConfigMaps is a list of ConfigMaps in the same namespace as the VerneMQ
	// object, which shall be mounted into the VerneMQ Pods.
	// The ConfigMaps are mounted into /etc/vernemq/configmaps/<configmap-name>.
	ConfigMaps []string `json:"configMaps,omitempty"`
	// Defines the config that is used when starting VerneMQ (similar to vernemq.conf)
	VMQConfig string `json:"vmqConfig,omitempty"`
	// Defines the arguments passed to the erlang VM when starting VerneMQ
	VMArgs string `json:"vmArgs,omitempty"`
	// Defines additional environment variables for the VerneMQ container
	// The environment variables can be used to template the VMQConfig and VMArgs
	Env []v1.EnvVar `json:"env,omitempty"`
	// Defines external plugins that have to be compiled and loaded into VerneMQ
	ExternalPlugins []PluginSource `json:"externalPlugins,omitempty"`
	// Defines the reloadable config that VerneMQ regularly checks and applies
	Config ReloadableConfig `json:"config,omitempty"`
}

// ReloadableConfig defines the reloadable parts of the VerneMQ configuration
// +k8s:openapi-gen=true
type ReloadableConfig struct {
	Plugins   []Plugin     `json:"plugins,omitempty"`
	Listeners []Listener   `json:"listeners,omitempty"`
	Configs   []ConfigItem `json:"configs,omitempty"`
}

// PluginSource defines the plugins to be fetched, compiled and loaded into the VerneMQ container
// +k8s:openapi-gen=true
type PluginSource struct {
	ApplicationName string `json:"applicationName"`
	RepoURL         string `json:"repoURL"`
	VersionType     string `json:"versionType"`
	Version         string `json:"version"`
}

// Plugin defines the plugins to be enabled by VerneMQ
// +k8s:openapi-gen=true
type Plugin struct {
	// The name of the plugin application
	Name string `json:"name"`
	// The path to the plugin application
	Path string `json:"path,omitempty"`
}

// ConfigItem defines a single reloadable VerneMQ config item
// +k8s:openapi-gen=true
type ConfigItem struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Listener defines the listeners to be started
// !!! Make sure that the JSON name of the property converted to snake-case results in the value accepted by vmq-admin listener start
// +k8s:openapi-gen=true
type Listener struct {
	Address          string `json:"address"`
	Port             uint16 `json:"port"`
	Mountpoint       string `json:"mountpoint,omitempty"`
	NrOfAcceptors    uint32 `json:"nrOfAcceptors,omitempty"`
	MaxConnections   uint32 `json:"maxConnections,omitempty"`
	ProtocolVersions string `json:"protocolVersions,omitempty"`
	Websocket        bool   `json:"websocket,omitempty"`
	// Enable PROXY v2 protocol for this listener
	ProxyProtocol bool `json:"proxyProtocol,omitempty"`
	// If PROXY v2 is enabled for this listener use this flag to decide if the common name should replace the MQTT username
	// Enabled by default (use `=false`) to disable
	UseCnAsUsername bool `json:"useCnAsUsername,omitempty"`
	// The TLS Config.
	TLSConfig *TLSConfig `json:"tlsConfig,omitempty"`
}

// TLSConfig defines the TLS configuration used for a TLS enabled listener
// !!! Make sure that the JSON name of the property converted to snake-case results in the value accepted by vmq-admin listener start
// +k8s:openapi-gen=true
type TLSConfig struct {
	// The path to the cafile containing the PEM encoded CA certificates that are trusted by the server.
	Cafile string `json:"cafile"`
	// The path to the PEM encoded server certificate
	Certfile string `json:"certfile"`
	// The path to the PEM encoded key file
	Keyfile string `json:"keyfile"`
	// The list of allowed ciphers, each separated by a colon
	Ciphers string `json:"ciphers,omitempty"`
	// Use client certificates to authenticate your clients
	RequireCertificate bool `json:"requireCertificate,omitempty"`
	// If RequreCertificate is true then the CN value from the client certificate is used as the username for authentication
	UseIdentityAsUsername bool `json:"useIdentityAsUsername,omitempty"`
	// If RequreCertificate is true, you can use a certificate revocation list
	// file to revoke access to particular client certificates. The file has to be PEM encoded.
	Crlfile string `json:"crlfile,omitempty"`
}

// StorageSpec defines the configured storage for VerneMQ Cluster nodes.
// If neither `emptyDir` nor `volumeClaimTemplate` is specified, then by default an [EmptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) will be used.
// +k8s:openapi-gen=true
type StorageSpec struct {
	// EmptyDirVolumeSource to be used by the VerneMQ StatefulSets. If specified, used in place of any volumeClaimTemplate. More
	// info: https://kubernetes.io/docs/concepts/storage/volumes/#emptydir
	EmptyDir *v1.EmptyDirVolumeSource `json:"emptyDir,omitempty"`
	// A PVC spec to be used by the VerneMQ StatefulSets.
	VolumeClaimTemplate v1.PersistentVolumeClaim `json:"volumeClaimTemplate,omitempty"`
}

// VerneMQStatus defines the observed state of VerneMQ
// +k8s:openapi-gen=true
type VerneMQStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html

	// Nodes are the names of the VerneMQ pods
	Nodes []string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VerneMQ is the Schema for the vernemqs API
// +k8s:openapi-gen=true
type VerneMQ struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object’s metadata. More info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata
	// +k8s:openapi-gen=false
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VerneMQSpec   `json:"spec"`
	Status VerneMQStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VerneMQList contains a list of VerneMQ
type VerneMQList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VerneMQ `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VerneMQ{}, &VerneMQList{})
}
