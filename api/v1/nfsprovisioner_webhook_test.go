package v1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NfsProvisioner webhooks", func() {
	It("Check NfsProvisioner webhook negative regionID", func() {
		provisioner := NfsProvisioner{
			TypeMeta: metav1.TypeMeta{
				Kind:       "NfsProvisioner",
				APIVersion: GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "provisioner1",
				Namespace: "default",
			},
			Spec: NfsProvisionerSpec{
				APIToken:  "faketoken",
				RegionID:  -2,
				ProjectID: 1,
			},
		}
		err := k8sClient.Create(ctx, &provisioner)
		Expect(err).To(MatchError(ContainSubstring("must be positive")))

	})
	It("Check NfsProvisioner webhook negative projectID", func() {
		provisioner := NfsProvisioner{
			TypeMeta: metav1.TypeMeta{
				Kind:       "NfsProvisioner",
				APIVersion: GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "provisioner1",
				Namespace: "default",
			},
			Spec: NfsProvisionerSpec{
				APIToken:  "faketoken",
				RegionID:  1,
				ProjectID: -1,
			},
		}
		err := k8sClient.Create(ctx, &provisioner)
		Expect(err).To(MatchError(ContainSubstring("must be positive")))

	})
	It("Check NfsProvisioner webhook check defaults", func() {
		provisioner := NfsProvisioner{
			TypeMeta: metav1.TypeMeta{
				Kind:       "NfsProvisioner",
				APIVersion: GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "provisioner1",
				Namespace: "default",
			},
			Spec: NfsProvisionerSpec{
				APITokenSecretRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "gcore-api-token"},
				},
				RegionID:  1,
				ProjectID: 1,
			},
		}
		err := k8sClient.Create(ctx, &provisioner)
		Expect(err).NotTo(HaveOccurred())
		Expect(provisioner.Spec.APIURL).To(Equal(DefaultApiUrl))
		Expect(provisioner.Spec.HelmRepository).To(Equal(DefaultHelmRepository))
		Expect(provisioner.Spec.ChartName).To(Equal(DefaultHelmChartName))
		Expect(provisioner.Spec.ImageVersion).To(Equal(DefaultNfsProvisionerImageVersion))
		Expect(provisioner.Spec.APITokenSecretRef.Key).To(Equal(DefaultAPITokenSecretKey))
	})
	It("Check NfsProvisioner webhook rejects missing API token configuration", func() {
		provisioner := NfsProvisioner{
			TypeMeta: metav1.TypeMeta{
				Kind:       "NfsProvisioner",
				APIVersion: GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "provisioner-no-token",
				Namespace: "default",
			},
			Spec: NfsProvisionerSpec{
				RegionID:  1,
				ProjectID: 1,
			},
		}
		err := k8sClient.Create(ctx, &provisioner)
		Expect(err).To(MatchError(ContainSubstring("one of apiToken or apiTokenSecretRef must be set")))
	})
	It("Check NfsProvisioner webhook rejects both apiToken and apiTokenSecretRef", func() {
		provisioner := NfsProvisioner{
			TypeMeta: metav1.TypeMeta{
				Kind:       "NfsProvisioner",
				APIVersion: GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "provisioner-both-tokens",
				Namespace: "default",
			},
			Spec: NfsProvisionerSpec{
				APIToken: "faketoken",
				APITokenSecretRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "gcore-api-token"},
				},
				RegionID:  1,
				ProjectID: 1,
			},
		}
		err := k8sClient.Create(ctx, &provisioner)
		Expect(err).To(MatchError(ContainSubstring("mutually exclusive")))
	})
	It("Check NfsProvisioner webhook rejects apiTokenSecretRef without a name", func() {
		provisioner := NfsProvisioner{
			TypeMeta: metav1.TypeMeta{
				Kind:       "NfsProvisioner",
				APIVersion: GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "provisioner-unnamed-secret",
				Namespace: "default",
			},
			Spec: NfsProvisionerSpec{
				APITokenSecretRef: &corev1.SecretKeySelector{},
				RegionID:          1,
				ProjectID:         1,
			},
		}
		err := k8sClient.Create(ctx, &provisioner)
		Expect(err).To(MatchError(ContainSubstring("secret name must be set")))
	})
})
