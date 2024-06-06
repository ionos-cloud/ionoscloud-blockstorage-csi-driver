# ionoscloud-blockstorage-csi-driver

![image-version] ![chart-version]

<p align="center">
  <img src="https://raw.githubusercontent.com/container-storage-interface/spec/master/logo.png" width="200">
  <img src="./docs/assets/images/LOGO_IONOS_Blue_RGB.png" width="200">
</p>

Container Storage Interface (CSI) plugin for IONOS Cloud [block storage][1].

The repository currently only contains the Helm chart.
The actual driver will be open sourced later.

Select release candidate images will be pushed to GitHub Container Registry.
Only these images should be used together with the helm chart.

## Features

* **Static Provisioning**: Associate an externally-created volume with a [Persistent Volume][2] (PV) for consumption
  within Kubernetes.
* **Dynamic Provisioning**: Automatically create volumes and associated [Persistent Volumes][2] from
  [PersistentVolumeClaims][3] (PVC). Parameters can be passed in via a [StorageClass][4] for fine-grained control over
  volume creation.
* **Volume Resizing**: Expand a volume without downtime by specifying a new size in the [PersistentVolumeClaims][5].
* **Volume Snapshots**: Create and restore [snapshots][6] taken from a volume in Kubernetes.

Configurable storage class parameters:

* `type`: Available type names are `HDD`, `SSD`, `SSD Standard` or `SSD Premium`. `SSD` is an alias for `SSD Standard`.
* `availabilityZone`: Defaults to `AUTO` if not set.
* `fstype`: Currently only `ext2`, `ext3` and `ext4` (default) are supported.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for more information.

[1]: https://cloud.ionos.com/storage/block-storage
[2]: https://kubernetes.io/docs/concepts/storage/persistent-volumes
[3]: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#dynamic
[4]: https://kubernetes.io/docs/concepts/storage/storage-classes/#the-storageclass-resource
[5]: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#expanding-persistent-volumes-claims
[6]: https://kubernetes.io/docs/concepts/storage/volume-snapshots

[image-version]: <https://ghcr-badge.egpl.dev/ionos-cloud/ionoscloud-blockstorage-csi-driver/latest_tag?label=app version>
[chart-version]: <https://ghcr-badge.egpl.dev/ionos-cloud/helm-charts/ionoscloud-blockstorage-csi-driver/latest_tag?label=chart version>
