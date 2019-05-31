
# API Docs
This Document documents the types introduced by the VerneMQ Operator to be consumed by users.
> Note this document is generated from code comments. When contributing a change to this document please do so by changing the code comments.

## Table of Contents
* [ConfigItem](#configitem)
* [Listener](#listener)
* [Plugin](#plugin)
* [PluginSource](#pluginsource)
* [ReloadableConfig](#reloadableconfig)
* [StorageSpec](#storagespec)
* [TLSConfig](#tlsconfig)
* [VerneMQ](#vernemq)
* [VerneMQList](#vernemqlist)
* [VerneMQSpec](#vernemqspec)
* [VerneMQStatus](#vernemqstatus)

## ConfigItem

ConfigItem defines a single reloadable VerneMQ config item

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | Defines the name of the config | string | true |
| value | Defines the value of the config | string | true |

[Back to TOC](#table-of-contents)

## Listener

Listener defines the listeners to be started !!! Make sure that the JSON name of the property converted to snake-case results in the value accepted by vmq-admin listener start

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| address | Defines the Network address the listener accepts connections on. Alternatively pass the name of the network interface. | string | true |
| port | Defines the TCP port | uint16 | true |
| mountpoint | Defines the mountpoint for this listener. Defaults to \"\" | string | false |
| nrOfAcceptors | Defines the number of TCP acceptor processes. | uint32 | false |
| maxConnections | Defines the number of allowed concurrent TCP connections. | uint32 | false |
| protocolVersions | Defines the allowed MQTT protocol version. Specified as a comma separated list e.g. \"3,4,5\" | string | false |
| websocket | Specifies that this listener accepts connections over HTTP websockets. | bool | false |
| proxyProtocol | Enable PROXY v2 protocol for this listener | bool | false |
| useCnAsUsername | If PROXY v2 is enabled for this listener use this flag to decide if the common name should replace the MQTT username Enabled by default (use `=false`) to disable | bool | false |
| tlsConfig | The TLS Config. | *[TLSConfig](#tlsconfig) | false |

[Back to TOC](#table-of-contents)

## Plugin

Plugin defines the plugins to be enabled by VerneMQ

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| name | The name of the plugin application | string | true |
| path | The path to the plugin application | string | false |

[Back to TOC](#table-of-contents)

## PluginSource

PluginSource defines the plugins to be fetched, compiled and loaded into the VerneMQ container

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| applicationName | The name of the plugin application | string | true |
| repoURL | The URL of the Git repository | string | true |
| versionType | The type to checkout, can be \"branch\", \"tag\", or \"commit\" | string | true |
| version | The version to checkout, can be name of the branch or tag, or the Git commit ref | string | true |

[Back to TOC](#table-of-contents)

## ReloadableConfig

ReloadableConfig defines the reloadable parts of the VerneMQ configuration

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| plugins | Defines the plugins to enable when VerneMQ starts | [][Plugin](#plugin) | false |
| listeners | Defines the listeners to enable when VerneMQ starts | [][Listener](#listener) | false |
| configs | Configures VerneMQ, valid are all the properties that can be set with the `vmq-admin set` command | [][ConfigItem](#configitem) | false |

[Back to TOC](#table-of-contents)

## StorageSpec

StorageSpec defines the configured storage for VerneMQ Cluster nodes. If neither `emptyDir` nor `volumeClaimTemplate` is specified, then by default an [EmptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) will be used.

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| emptyDir | EmptyDirVolumeSource to be used by the VerneMQ StatefulSets. If specified, used in place of any volumeClaimTemplate. More info: https://kubernetes.io/docs/concepts/storage/volumes/#emptydir | *[v1.EmptyDirVolumeSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#emptydirvolumesource-v1-core) | false |
| volumeClaimTemplate | A PVC spec to be used by the VerneMQ StatefulSets. | [v1.PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#persistentvolumeclaim-v1-core) | false |

[Back to TOC](#table-of-contents)

## TLSConfig

TLSConfig defines the TLS configuration used for a TLS enabled listener !!! Make sure that the JSON name of the property converted to snake-case results in the value accepted by vmq-admin listener start

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| cafile | The path to the cafile containing the PEM encoded CA certificates that are trusted by the server. | string | true |
| certfile | The path to the PEM encoded server certificate | string | true |
| keyfile | The path to the PEM encoded key file | string | true |
| ciphers | The list of allowed ciphers, each separated by a colon | string | false |
| requireCertificate | Use client certificates to authenticate your clients | bool | false |
| useIdentityAsUsername | If RequreCertificate is true then the CN value from the client certificate is used as the username for authentication | bool | false |
| crlfile | If RequreCertificate is true, you can use a certificate revocation list file to revoke access to particular client certificates. The file has to be PEM encoded. | string | false |

[Back to TOC](#table-of-contents)

## VerneMQ

VerneMQ is the Schema for the vernemqs API

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata | Standard object’s metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata | [metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta) | false |
| spec |  | [VerneMQSpec](#vernemqspec) | true |
| status |  | [VerneMQStatus](#vernemqstatus) | false |

[Back to TOC](#table-of-contents)

## VerneMQList

VerneMQList contains a list of VerneMQ

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| metadata |  | [metav1.ListMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#listmeta-v1-meta) | false |
| items |  | [][VerneMQ](#vernemq) | true |

[Back to TOC](#table-of-contents)

## VerneMQSpec

VerneMQSpec defines the desired state of VerneMQ

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| podMetadata | Standard object’s metadata. More info: https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#metadata Metadata Labels and Annotations gets propagated to the vernemq pods. | *[metav1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#objectmeta-v1-meta) | false |
| size | Size is the size of the VerneMQ deployment | *int32 | false |
| version | Version of VerneMQ to be deployed | string | false |
| tag | Tag of VerneMQ container image to be deployed. Defaults to the value of `version`. Version is ignored if Tag is set. | string | false |
| sha | SHA of VerneMQ container image to be deployed. Defaults to the value of `version`. Similar to a tag, but the SHA explicitly deploys an immutable container image. Version and Tag are ignored if SHA is set. | string | false |
| image | Image if specified has precedence over baseImage, tag and sha combinations. Specifying the version is still necessary to ensure the VerneMQ Operator knows what version of VerneMQ is being configured. | *string | false |
| baseImage | Base image to use for a VerneMQ deployment. | string | false |
| imagePullSecrets | An optional list of references to secrets in the same namespace to use for pulling vernemq images from registries see http://kubernetes.io/docs/user-guide/images#specifying-imagepullsecrets-on-a-pod | [][v1.LocalObjectReference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#localobjectreference-v1-core) | false |
| securityContext | SecurityContext holds pod-level security attributes and common container settings. This defaults to non root user with uid 10000 and gid 10000 for VerneMQ >1.7.0 and default PodSecurityContext for other versions. | *v1.PodSecurityContext | false |
| storage | Storage spec to specify how storage shall be used. | *[StorageSpec](#storagespec) | false |
| containers | Containers allows injecting additional containers. | []v1.Container | false |
| resources | Define resources requests and limits for single Pods. | [v1.ResourceRequirements](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#resourcerequirements-v1-core) | false |
| serviceAccountName | ServiceAccountName is the name of the ServiceAccount to use to run the VerneMQ Pods. | string | false |
| nodeSelector | Define which Nodes the Pods are scheduled on. | map[string]string | false |
| priorityClassName | Priority class assigned to the Pods | string | false |
| affinity | If specified, the pod's scheduling constraints. | *v1.Affinity | false |
| tolerations | If specified, the pod's tolerations. | []v1.Toleration | false |
| dropoutPeriodSeconds |  | *int64 | false |
| terminationGracePeriodSeconds |  | *int64 | false |
| secrets | Secrets is a list of Secrets in the same namespace as the VerneMQ object, which shall be mounted into the VerneMQ Pods. The Secrets are mounted into /etc/vernemq/secrets/<secret-name>. | []string | false |
| configMaps | ConfigMaps is a list of ConfigMaps in the same namespace as the VerneMQ object, which shall be mounted into the VerneMQ Pods. The ConfigMaps are mounted into /etc/vernemq/configmaps/<configmap-name>. | []string | false |
| vmqConfig | Defines the config that is used when starting VerneMQ (similar to vernemq.conf) | string | false |
| vmArgs | Defines the arguments passed to the erlang VM when starting VerneMQ | string | false |
| env | Defines additional environment variables for the VerneMQ container The environment variables can be used to template the VMQConfig and VMArgs | []v1.EnvVar | false |
| externalPlugins | Defines external plugins that have to be compiled and loaded into VerneMQ | [][PluginSource](#pluginsource) | false |
| config | Defines the reloadable config that VerneMQ regularly checks and applies | [ReloadableConfig](#reloadableconfig) | false |

[Back to TOC](#table-of-contents)

## VerneMQStatus

VerneMQStatus defines the observed state of VerneMQ

| Field | Description | Scheme | Required |
| ----- | ----------- | ------ | -------- |
| nodes | Nodes are the names of the VerneMQ pods | []string | true |

[Back to TOC](#table-of-contents)
