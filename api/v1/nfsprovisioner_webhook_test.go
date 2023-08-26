package v1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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
	})
})
