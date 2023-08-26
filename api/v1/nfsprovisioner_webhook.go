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

package v1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	DefaultNfsProvisionerImageVersion = "v4.0.2"
	DefaultApiUrl                     = "https://api.gcore.com/cloud"
	DefaultHelmRepository             = "https://kubernetes-sigs.github.io/nfs-subdir-external-provisioner"
	DefaultHelmChartName              = "nfs-subdir-external-provisioner"
)

// log is for logging in this package.
var nfsprovisionerlog = logf.Log.WithName("nfsprovisioner-resource")

func (r *NfsProvisioner) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-crd-gcore-sfs-controller-io-v1-nfsprovisioner,mutating=true,failurePolicy=fail,sideEffects=None,groups=crd.gcore-sfs-controller.io,resources=nfsprovisioners,verbs=create;update,versions=v1,name=mnfsprovisioner.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &NfsProvisioner{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *NfsProvisioner) Default() {

	if r.Spec.APIURL == "" {
		r.Spec.APIURL = DefaultApiUrl
	}
	nfsprovisionerlog.Info("default", "apiURL", r.Spec.APIURL)
	if r.Spec.HelmRepository == "" {
		r.Spec.HelmRepository = DefaultHelmRepository
	}
	nfsprovisionerlog.Info("default", "helmRepository", r.Spec.HelmRepository)
	if r.Spec.ChartName == "" {
		r.Spec.ChartName = DefaultHelmChartName
	}
	nfsprovisionerlog.Info("default", "chartName", r.Spec.ChartName)
	if r.Spec.ImageVersion == "" {
		r.Spec.ImageVersion = DefaultNfsProvisionerImageVersion
	}
	nfsprovisionerlog.Info("default", "imageVersion", r.Spec.ImageVersion)
}

//+kubebuilder:webhook:path=/validate-crd-gcore-sfs-controller-io-v1-nfsprovisioner,mutating=false,failurePolicy=fail,sideEffects=None,groups=crd.gcore-sfs-controller.io,resources=nfsprovisioners,verbs=create;update,versions=v1,name=vnfsprovisioner.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &NfsProvisioner{}

func ValidateNfsProvisioner(r *NfsProvisioner) error {
	var allErrs field.ErrorList
	if r.Spec.RegionID <= 0 {
		regionErr := field.Invalid(field.NewPath("spec").Child("region"), r.Spec.RegionID, "must be positive")
		allErrs = append(allErrs, regionErr)
	}
	if r.Spec.ProjectID <= 0 {
		projectErr := field.Invalid(field.NewPath("spec").Child("project"), r.Spec.RegionID, "must be positive")
		allErrs = append(allErrs, projectErr)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(
		schema.GroupKind{Group: "crd.gcore-sfs-controller.io", Kind: "NfsProvisioner"},
		r.Name, allErrs)
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *NfsProvisioner) ValidateCreate() (admission.Warnings, error) {
	return nil, ValidateNfsProvisioner(r)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *NfsProvisioner) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	return nil, ValidateNfsProvisioner(r)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *NfsProvisioner) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}
