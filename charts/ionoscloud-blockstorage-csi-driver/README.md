# ionoscloud-blockstorage-csi-driver

![Version: 0.4.2](https://img.shields.io/badge/Version-0.4.2-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v1.9.0-rc.2](https://img.shields.io/badge/AppVersion-v1.9.0--rc.2-informational?style=flat-square)

**Homepage:** <https://github.com/ionos-cloud/ionoscloud-blockstorage-csi-driver>

The [Container Storage Interface][csi-spec] (CSI) driver plugin communicates with the [IONOS Cloud API][cloud-api] to manage storage.
The visibility and permissions it has depend on the authentication token it is given.

Check out [this page][token-docs] to learn more about managing tokens.

## Prerequisites

* Kubernetes 1.25+
* Helm 3+

Before installing create a secret that contains your IONOS Cloud authentication token:

```console
kubectl -n kube-system create secret generic csi-secret --from-literal token=<your-token>
```

The key **must** be named `token`.

The CSI node server expects the file `/etc/ie-csi/cfg.json` to exist on every VM.
The file must contain the datacenter ID of the VM in the following format:

```json
{"datacenter-id": "<DATACENTER_ID>"}
```

## Installation

Provide the secret name during installation:

```console
helm install -n kube-system ionoscloud-blockstorage-csi-driver \
    oci://ghcr.io/ionos-cloud/helm-charts/ionoscloud-blockstorage-csi-driver \
    --set tokenSecretName=csi-secret
```

> [!IMPORTANT]
> Be aware that tokens have a limited lifetime. CSI pods must be restarted every time the token is updated.

### Multi-tenancy setup

The default settings of the CSI driver helm chart are meant to be used in a single-tenancy manner.
Should you need to install multiple CSI drivers using tokens from the same users or contracts, e.g. if you manage more than 1 cluster,
you need to set the `clusterName` value on installation.

```console
helm install -n kube-system ionoscloud-blockstorage-csi-driver \
    oci://ghcr.io/ionos-cloud/helm-charts/ionoscloud-blockstorage-csi-driver \
    --set tokenSecretName=csi-secret --set clusterName=production
helm install -n kube-system ionoscloud-blockstorage-csi-driver \
    oci://ghcr.io/ionos-cloud/helm-charts/ionoscloud-blockstorage-csi-driver \
    --set tokenSecretName=csi-secret --set clusterName=staging
```

> [!WARNING]
> The `clusterName` must not be changed after storage has already been provisioned.

## Upgrade

### From 0.2.0 to 0.3.0

The provided CRDs were updated to contain CEL markers for validation, which is why the **minimum required version** is now 1.25.
The helm CLI will ignore any changes to CRDs which is why they must be updated manually before running `helm upgrade`.

```console
helm show crds ./charts/ionoscloud-blockstorage-csi-driver | kubectl apply -f -
```

## Values

### Attacher

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| attacher.extraArgs | object | `{"timeout":"270s"}` | Additional command-line arguments |
| attacher.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| attacher.image.repository | string | `"registry.k8s.io/sig-storage/csi-attacher"` | Image repository |
| attacher.image.tag | string | `"v4.8.1"` | Image tag |
| attacher.resources | object | `{"limits":{"memory":"100Mi"},"requests":{"cpu":"10m","memory":"25Mi"}}` | Resource requests and limits |

### Daemonset

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| daemonset.affinity | object | `{}` | Affinity for Daemonset pods. |
| daemonset.imagePullSecrets | list | `[]` | List of image pull secret names for Daemonset pods. |
| daemonset.nodeSelector | object | `{}` | Node selector for Daemonset pods. |
| daemonset.podAnnotations | object | `{}` | Additional annotations for Daemonset pods. |
| daemonset.podLabels | object | `{}` | Additional labels for Daemonset pods. |
| daemonset.podSecurityContext | object | `{}` | Security context for Daemonset pods. |
| daemonset.tolerations | list | `[]` | Tolerations for Daemonset pods. |

### Deployment

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| deployment.affinity | object | `{}` | Affinity for Deployment pods. |
| deployment.imagePullSecrets | list | `[]` | List of image pull secret names for Deployment pods. |
| deployment.nodeSelector | object | `{}` | Node selector for Deployment pods. |
| deployment.podAnnotations | object | `{}` | Additional annotations for Deployment pods. |
| deployment.podLabels | object | `{}` | Additional labels for Deployment pods. |
| deployment.podSecurityContext | object | `{}` | Security context for Deployment pods. |
| deployment.replicaCount | int | `1` | Number of Deployment pods. Setting this higher than 1 will enable leader election. |
| deployment.securityContext | object | `{"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true}` | Security context for Deployment containers. |
| deployment.tolerations | list | `[]` | Tolerations for Deployment pods. |

### Controller server

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| driver.controller.extraArgs | object | `{}` | Additional command-line arguments |
| driver.controller.extraEnv | list | `[]` | Additional environment variables |
| driver.controller.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| driver.controller.image.repository | string | `"ghcr.io/ionos-cloud/ionoscloud-blockstorage-csi-driver"` | Image repository |
| driver.controller.image.tag | string | Defaults to appVersion | Image tag |
| driver.controller.resources | object | `{"limits":{"memory":"100Mi"},"requests":{"cpu":"10m","memory":"25Mi"}}` | Resource requests and limits |

### Node server

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| driver.node.extraArgs | object | `{}` | Additional command-line arguments |
| driver.node.extraEnv | list | `[]` | Additional environment variables |
| driver.node.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| driver.node.image.repository | string | `"ghcr.io/ionos-cloud/ionoscloud-blockstorage-csi-driver"` | Image repository |
| driver.node.image.tag | string | Defaults to appVersion | Image tag |
| driver.node.resources | object | `{"limits":{"memory":"50Mi"},"requests":{"cpu":"10m","memory":"25Mi"}}` | Resource requests and limits |

### Monitoring

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| monitoring.additionalLabels | object | `{}` | Additional labels for the PodMonitor. |
| monitoring.enabled | bool | `false` | If true, render Prometheus PodMonitor resource. |
| monitoring.nameOverride | string | `""` | Specify name override for the PodMonitor. |
| monitoring.namespace | string | Release namespace | Specify namespace override for the PodMonitor. |
| monitoring.scrapeInterval | string | `"30s"` | Metrics scrape interval as duration string. |

### Provisioner

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| provisioner.extraArgs | object | `{"timeout":"930s"}` | Additional command-line arguments |
| provisioner.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| provisioner.image.repository | string | `"registry.k8s.io/sig-storage/csi-provisioner"` | Image repository |
| provisioner.image.tag | string | `"v5.2.0"` | Image tag |
| provisioner.resources | object | `{"limits":{"memory":"100Mi"},"requests":{"cpu":"10m","memory":"25Mi"}}` | Resource requests and limits |

### Node driver registrar

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| registrar.extraArgs | object | `{}` | Additional command-line arguments |
| registrar.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| registrar.image.repository | string | `"registry.k8s.io/sig-storage/csi-node-driver-registrar"` | Image repository |
| registrar.image.tag | string | `"v2.13.0"` | Image tag |
| registrar.resources | object | `{"limits":{"memory":"30Mi"},"requests":{"cpu":"10m","memory":"15Mi"}}` | Resource requests and limits |

### Resizer

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| resizer.extraArgs | object | `{"timeout":"150s"}` | Additional command-line arguments |
| resizer.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| resizer.image.repository | string | `"registry.k8s.io/sig-storage/csi-resizer"` | Image repository |
| resizer.image.tag | string | `"v1.13.2"` | Image tag |
| resizer.resources | object | `{"limits":{"memory":"100Mi"},"requests":{"cpu":"10m","memory":"25Mi"}}` | Resource requests and limits |

### Snapshot controller

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| snapshotController.extraArgs | object | `{}` | Additional command-line arguments |
| snapshotController.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| snapshotController.image.repository | string | `"registry.k8s.io/sig-storage/snapshot-controller"` | Image repository |
| snapshotController.image.tag | string | `"v8.2.1"` | Image tag |
| snapshotController.resources | object | `{"limits":{"memory":"100Mi"},"requests":{"cpu":"10m","memory":"25Mi"}}` | Resource requests and limits |

### Snapshotter

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| snapshotter.extraArgs | object | `{"timeout":"300s"}` | Additional command-line arguments |
| snapshotter.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| snapshotter.image.repository | string | `"registry.k8s.io/sig-storage/csi-snapshotter"` | Image repository |
| snapshotter.image.tag | string | `"v8.2.1"` | Image tag |
| snapshotter.resources | object | `{"limits":{"memory":"100Mi"},"requests":{"cpu":"10m","memory":"25Mi"}}` | Resource requests and limits |

### Other Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| className | string | `"ionos-cloud"` | Name of VolumeSnapshotClass. Also used as prefix for StorageClasses. |
| clusterName | string | `"k8s"` | Name used to identify managed storage resources. |
| driverName | string | `"cloud.ionos.com"` | Name of the driver in the storage class. |
| nameOverride | string | `""` | Specify a custom name override. This only influences Kubernetes resource names, not properties. |
| registry | string | Omit if empty | Specify a custom registry name that will be used as prefix for all images. Useful when pulling images from a registry mirror. |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.automount | bool | `true` | Automatically mount a ServiceAccount's API credentials? |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use. If not set and create is true, a name is generated from template |
| tokenSecretName | string | `""` | Name of the secret that contains the token used for cloud API authentication. Must contain the key "token". |

[cloud-api]: https://api.ionos.com/docs/cloud/v6/
[token-docs]: https://docs.ionos.com/cloud/getting-started/basic-tutorials/manage-authentication-tokens
[csi-spec]: https://github.com/container-storage-interface/spec
[block-storage-docs]: https://cloud.ionos.com/storage/block-storage
