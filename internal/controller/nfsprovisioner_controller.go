/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strings"

	crdv1 "github.com/G-Core/gcore-sfs-controller/api/v1"
	"github.com/G-Core/gcore-sfs-controller/pkg/gcoreclient"
	"github.com/G-Core/gcorelabscloud-go/gcore/file_share/v1/file_shares"
	gohelmclient "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	kerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const RepositoryName = "nfs-subdir-external-provisioner"
const RepositoryUrl = "https://kubernetes-sigs.github.io/nfs-subdir-external-provisioner/"

const (
	FileShareIDLabelName      = "fileShareID"
	FileShareNameLabelName    = "fileShareName"
	NfsProvisionerIDLabelName = "nfsProvisionerID"
)

type StringSet map[string]bool

// NfsProvisionerReconciler reconciles a NfsProvisioner object
type NfsProvisionerReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	HelmClient      gohelmclient.Client
	FileShareClient gcoreclient.FileShareLister
}

//+kubebuilder:rbac:groups=crd.gcore-sfs-controller.io,resources=nfsprovisioners,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=crd.gcore-sfs-controller.io,resources=nfsprovisioners/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=crd.gcore-sfs-controller.io,resources=nfsprovisioners/finalizers,verbs=update
//+kubebuilder:rbac:groups=storage.k8s.io,resources=storageclasses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods;secrets;serviceaccounts;persistentvolumes;persistentvolumeclaims;events,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;create;upate;patch;delete
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=clusterroles;clusterrolebindings;roles;rolebindings,verbs=get;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=endpoints,verbs=get;list;watch;create;update;patch

func (r *NfsProvisionerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, reterr error) {
	log := log.FromContext(ctx)

	provisioner := crdv1.NfsProvisioner{}
	if err := r.Client.Get(ctx, req.NamespacedName, &provisioner); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get NfsProvisioner customer resource")
	}

	if provisioner.Spec.Paused {
		log.Info("Reconciliation is paused for this object", "provisioner", req.NamespacedName)
		return ctrl.Result{}, nil
	}
	log.Info("Start reconciling")

	if !controllerutil.ContainsFinalizer(&provisioner, crdv1.NfsProvisionerFinalizer) {
		controllerutil.AddFinalizer(&provisioner, crdv1.NfsProvisionerFinalizer)
		if err := r.Client.Update(ctx, &provisioner, &client.UpdateOptions{}); err != nil {
			log.Error(err, "Failed to update NfsProvisioner to add finalizer")
			return ctrl.Result{}, err
		}
	}
	err := r.HelmClient.AddOrUpdateChartRepo(repo.Entry{
		Name: RepositoryName,
		URL:  provisioner.Spec.HelmRepository,
	})
	if err != nil {
		log.Error(err, "add or update repo")
	}

	defer func() {
		// If object has finalizer attempt to update status.
		if controllerutil.ContainsFinalizer(&provisioner, crdv1.NfsProvisionerFinalizer) {
			if err := r.updateStatus(ctx, &provisioner); err != nil {
				log.Error(err, "Failed to update NfsProvisioner Status")
				reterr = kerrors.NewAggregate([]error{reterr, err})
			}
			// Always attempt to Update the NfsProvisioner object and status after each reconciliation.
			if err := r.Client.Status().Update(ctx, &provisioner, &client.SubResourceUpdateOptions{}); err != nil {
				log.Error(err, "Failed to update NfsProvisioner status")
				reterr = kerrors.NewAggregate([]error{reterr, err})
			}
		}
	}()

	// Handle deletion reconciliation loop.
	if !provisioner.ObjectMeta.DeletionTimestamp.IsZero() {
		result, err := r.reconcileDelete(ctx, &provisioner)
		if err != nil {
			log.Error(err, "Failed delete reconciliation NfsProvisioner")
			return result, err
		}
		return ctrl.Result{}, nil
	}
	result, err := r.reconcileNormal(ctx, &provisioner)
	if err != nil {
		log.Error(err, "Failed normal reconciliation of NfsProvisioner", "namespace", provisioner.Namespace, "name", provisioner.Name)
		return result, err
	}

	log.Info("Reconciling has completed")
	return ctrl.Result{}, nil
}

func (r *NfsProvisionerReconciler) updateStatus(ctx context.Context, provisioner *crdv1.NfsProvisioner) error {
	provisionerPodList := corev1.PodList{}
	listOptions := client.ListOptions{
		LabelSelector: labels.SelectorFromSet(
			map[string]string{NfsProvisionerIDLabelName: string(provisioner.UID)},
		),
	}
	if err := r.Client.List(ctx, &provisionerPodList, &listOptions); err != nil {
		return err
	}
	var status bool = true
	for _, pod := range provisionerPodList.Items {
		if pod.Status.Phase != corev1.PodRunning {
			status = false
			break
		}
	}
	provisioner.Status.ProvisionersReady = status
	return nil
}

func (r *NfsProvisionerReconciler) reconcileNormal(ctx context.Context, provisioner *crdv1.NfsProvisioner) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	allFileShares, err := r.FileShareClient.ListFileShares(provisioner)
	if err != nil {
		log.Error(err, "get file shares in the project", "regionID", provisioner.Spec.RegionID, "projectID", provisioner.Spec.ProjectID)
		return ctrl.Result{}, err
	}
	currentReleaseNameSet, err := r.getCurrentReleaseNameSet(ctx, provisioner)
	if err != nil {
		log.Error(err, "failed get provisioner helm releases")
		return ctrl.Result{}, err
	}
	createReleaseNameSet := make(map[string]bool)
	for _, fileShare := range allFileShares {
		// ConnectionPoint == "" if file share is creating
		if fileShare.ConnectionPoint != "" {
			releaseName, err := r.deployNfsProvisioner(ctx, provisioner, &fileShare)
			if err != nil {
				log.Error(err, "failed deploy chart", "namespace", provisioner.Namespace, "chartName", provisioner.Spec.ChartName)
				return ctrl.Result{}, err
			}
			createReleaseNameSet[releaseName] = true
		}
	}
	for currentReleaseName := range currentReleaseNameSet {
		if _, found := createReleaseNameSet[currentReleaseName]; !found {
			if err = r.HelmClient.UninstallReleaseByName(currentReleaseName); err != nil {
				log.Error(err, "failed uninstall chart", "namespace", provisioner.Namespace, "release", currentReleaseName)
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

func (r NfsProvisionerReconciler) getReleaseName(fileShareID string) string {
	return fmt.Sprintf("nfsprovisioner-%s", fileShareID)
}

func (r *NfsProvisionerReconciler) getCurrentReleaseNameSet(ctx context.Context, provisioner *crdv1.NfsProvisioner) (StringSet, error) {
	fileShareStorageClasseList := storagev1.StorageClassList{}
	listOptions := client.ListOptions{
		LabelSelector: labels.SelectorFromSet(
			map[string]string{NfsProvisionerIDLabelName: string(provisioner.UID)},
		),
	}
	if err := r.Client.List(ctx, &fileShareStorageClasseList, &listOptions); err != nil {
		return StringSet{}, err
	}
	currentReleaseNameSet := StringSet{}
	for _, storageClass := range fileShareStorageClasseList.Items {
		currentReleaseNameSet[r.getReleaseName(storageClass.Labels[FileShareIDLabelName])] = true
	}
	return currentReleaseNameSet, nil
}

func (r *NfsProvisionerReconciler) deployNfsProvisioner(ctx context.Context, provisioner *crdv1.NfsProvisioner, fileShare *file_shares.FileShare) (string, error) {
	nfsServer, nfsPath, err := r.getNfsServerAndPath(fileShare)
	if err != nil {
		return "", err
	}
	release, err := r.HelmClient.InstallOrUpgradeChart(ctx, &gohelmclient.ChartSpec{
		ReleaseName: r.getReleaseName(fileShare.ID),
		ChartName:   fmt.Sprintf("%s/%s", RepositoryName, provisioner.Spec.ChartName),
		Namespace:   provisioner.Namespace,
		ValuesOptions: values.Options{
			Values: []string{
				fmt.Sprintf("nfs.server=%s", nfsServer),
				fmt.Sprintf("nfs.path=%s", nfsPath),
				fmt.Sprintf("storageClass.name=nfs-%s", fileShare.ID),
				"storageClass.accessModes=ReadWriteMany",
				"storageClass.defaultClass=false",
				"nfs.mountOptions={soft}", // Options allow unmount volume when file share was deleted
				fmt.Sprintf("image.tag=%s", provisioner.Spec.ImageVersion),
				fmt.Sprintf("labels.%s=%s", NfsProvisionerIDLabelName, provisioner.UID),
				fmt.Sprintf("labels.%s=%s", FileShareIDLabelName, fileShare.ID),
				fmt.Sprintf("labels.%s=%s", FileShareNameLabelName, fileShare.Name),
			},
		}},
		nil)
	if err != nil {
		return "", err
	}
	return release.Name, nil
}

func (r *NfsProvisionerReconciler) reconcileDelete(ctx context.Context, provisioner *crdv1.NfsProvisioner) (ctrl.Result, error) {
	currentReleaseNameSet, err := r.getCurrentReleaseNameSet(ctx, provisioner)
	if err != nil {
		return ctrl.Result{}, err
	}

	for releaseName := range currentReleaseNameSet {
		err := r.HelmClient.UninstallReleaseByName(releaseName)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	if controllerutil.ContainsFinalizer(provisioner, crdv1.NfsProvisionerFinalizer) {
		controllerutil.RemoveFinalizer(provisioner, crdv1.NfsProvisionerFinalizer)
		if err := r.Client.Update(ctx, provisioner, &client.UpdateOptions{}); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *NfsProvisionerReconciler) getNfsServerAndPath(fileShare *file_shares.FileShare) (string, string, error) {
	// Connection point  "10.33.20.241:/shares/share-e1dca5e4-257d-47c2-82ac-980fa43e0da9"
	ServerAndPath := strings.Split(fileShare.ConnectionPoint, ":")
	//nolint: gomnd
	if len(ServerAndPath) != 2 {
		return "", "", fmt.Errorf("incorrect file share connection point %v", fileShare.ConnectionPoint)
	}
	return ServerAndPath[0], ServerAndPath[1], nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NfsProvisionerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&crdv1.NfsProvisioner{}).
		Complete(r)
}
