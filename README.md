# gcore-sfs-controller
This controller watches for nfs file shares in Gcore Cloud project and deploy storage classes and provisioner controller for each of them

## Description
Gcore SFS Controller allows us to integrate our File Share Servers with the Kubernetes cluster automatically.
When you create File Share and give access to your Kubernetes cluster, the controller configures a storage class for your k8s cluster and installs the nfs provisioner.

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

## Instalation
### Install Gcore SFS Controller:

```sh
kubectl apply -f example/deploy/gcore-sfs-controller-install.yaml
```
### Install CRD
1. You must fill in the next values:
    ```yaml
    spec:
    apiToken: <put your api token here>
    region: <put your region id here>
    project: <put your project id here>
    ```
    `apiToken`: Create API token in [CLOUD UI](https://gcore.com/docs/account-settings/create-use-or-delete-a-permanent-api-token).


    `region`: You can get a region id from our [API](https://api.gcore.com/docs/cloud#tag/Regions/operation/RegionHandler.get): You will get a list of regions from the "v1/regions" handler, and then you can find the needed region by the "display_name" field.

2. Install CRD

    ```sh
    kubectl apply -f example/deploy/nfsprovisioner.yaml
    ```

### Create File Share

Now, you can create a [File Share](https://gcore.com/docs/cloud/file-shares/configure-file-shares) and the controller configures a storage class and installs the nfs-provisioner automatically.

## Development
### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/gcore-sfs-controller:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/gcore-sfs-controller:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller from the cluster:

```sh
make undeploy
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)
