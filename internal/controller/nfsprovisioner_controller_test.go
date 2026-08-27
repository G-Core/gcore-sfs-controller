package controller

import (
	crdv1 "github.com/G-Core/gcore-sfs-controller/api/v1"
	"github.com/G-Core/gcore-sfs-controller/pkg/gcoreclient"
	"github.com/G-Core/gcorelabscloud-go/gcore/file_share/v1/file_shares"
	gohelmclient "github.com/mittwald/go-helm-client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

const testNfsProvisionerName = "test-provisioner"
const DefaultNamespace = "default"

var _ = Describe("NfsProvisioner Reconciler", func() {
	It("Calling reconcile should generate nfs storage class", func() {
		// Create NfsProvisioner CR and start its reconciliation
		provisioner := crdv1.NfsProvisioner{
			TypeMeta: metav1.TypeMeta{
				Kind:       "NfsProvisioner",
				APIVersion: crdv1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      testNfsProvisionerName,
				Namespace: DefaultNamespace,
			},
			Spec: crdv1.NfsProvisionerSpec{
				APIToken:       "faketoken",
				APIURL:         "http://127.0.0.1",
				RegionID:       2,
				ProjectID:      5,
				HelmRepository: "https://kubernetes-sigs.github.io/nfs-subdir-external-provisioner",
				ChartName:      "nfs-subdir-external-provisioner",
				ImageVersion:   "v4.0.2",
			},
		}
		err := k8sClient.Create(ctx, &provisioner)
		provisionerID := string(provisioner.UID)
		Expect(err).NotTo(HaveOccurred())

		helmClient, err := gohelmclient.NewClientFromRestConf(
			&gohelmclient.RestConfClientOptions{
				Options:    &gohelmclient.Options{},
				RestConfig: cfg,
			})
		Expect(err).NotTo(HaveOccurred())
		fileShare := file_shares.FileShare{
			Name:            "mock_file_share",
			ID:              "d918f840-29a2-4d54-a67e-5c9d4e34a408",
			Protocol:        "nfs",
			Status:          "available",
			Size:            2,
			VolumeType:      "default_share_type",
			CreatedAt:       nil,
			ConnectionPoint: "10.33.20.91:/shares/share-d994e4f4-0e01-4358-93fd-1eb4c273e505",
			CreatorTaskID:   nil,
			ProjectID:       1,
			RegionID:        1,
		}
		fileShareLiseter := gcoreclient.MockFileShareClient{
			FileShares: []file_shares.FileShare{
				fileShare,
			},
		}

		reconciler := NfsProvisionerReconciler{
			Client:          k8sClient,
			HelmClient:      helmClient,
			FileShareClient: fileShareLiseter,
		}
		_, err = reconciler.Reconcile(
			ctx,
			ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: DefaultNamespace,
					Name:      testNfsProvisionerName,
				}})
		Expect(err).NotTo(HaveOccurred())

		storageClassList := storagev1.StorageClassList{}
		err = k8sClient.List(ctx, &storageClassList)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(storageClassList.Items)).To(Equal(1))
		storageClass := storageClassList.Items[0]
		Expect(storageClass.Name).To(Equal("nfs-" + fileShare.ID))
		Expect(storageClass.Labels["nfsProvisionerID"]).To(Equal(provisionerID))
		Expect(storageClass.Labels["fileShareName"]).To(Equal(fileShare.Name))
		Expect(storageClass.Labels["fileShareID"]).To(Equal(fileShare.ID))

		// Remove provisioner and check that resources are deleted after reconciliation
		err = k8sClient.Delete(ctx, &provisioner)
		Expect(err).NotTo(HaveOccurred())

		_, err = reconciler.Reconcile(
			ctx,
			ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: DefaultNamespace,
					Name:      testNfsProvisionerName,
				}})
		Expect(err).NotTo(HaveOccurred())

		storageClassList = storagev1.StorageClassList{}
		err = k8sClient.List(ctx, &storageClassList)
		Expect(err).NotTo(HaveOccurred())
		Expect(len(storageClassList.Items)).To(Equal(0))

	})

	It("File share name with Helm --set syntax should not inject chart values", func() {
		provisioner := crdv1.NfsProvisioner{
			TypeMeta: metav1.TypeMeta{
				Kind:       "NfsProvisioner",
				APIVersion: crdv1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      testNfsProvisionerName + "-inject",
				Namespace: DefaultNamespace,
			},
			Spec: crdv1.NfsProvisionerSpec{
				APIToken:       "faketoken",
				APIURL:         "http://127.0.0.1",
				RegionID:       2,
				ProjectID:      5,
				HelmRepository: "https://kubernetes-sigs.github.io/nfs-subdir-external-provisioner",
				ChartName:      "nfs-subdir-external-provisioner",
				ImageVersion:   "v4.0.2",
			},
		}
		err := k8sClient.Create(ctx, &provisioner)
		Expect(err).NotTo(HaveOccurred())

		helmClient, err := gohelmclient.NewClientFromRestConf(
			&gohelmclient.RestConfClientOptions{
				Options:    &gohelmclient.Options{},
				RestConfig: cfg,
			})
		Expect(err).NotTo(HaveOccurred())
		fileShare := file_shares.FileShare{
			Name:            "legit,image.repository=evil/malicious,image.tag=latest",
			ID:              "0e2a9d4c-11a2-4f54-b67e-6c9d4e34a409",
			Protocol:        "nfs",
			Status:          "available",
			Size:            2,
			VolumeType:      "default_share_type",
			ConnectionPoint: "10.33.20.91:/shares/share-0e2a9d4c-11a2-4f54-b67e-6c9d4e34a409",
			ProjectID:       1,
			RegionID:        1,
		}
		fileShareLister := gcoreclient.MockFileShareClient{
			FileShares: []file_shares.FileShare{fileShare},
		}

		reconciler := NfsProvisionerReconciler{
			Client:          k8sClient,
			HelmClient:      helmClient,
			FileShareClient: fileShareLister,
		}
		_, err = reconciler.Reconcile(
			ctx,
			ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: DefaultNamespace,
					Name:      provisioner.Name,
				}})
		Expect(err).NotTo(HaveOccurred())

		// The user-supplied release values must not contain keys injected via
		// commas/equals in the file share name.
		releaseValues, err := helmClient.GetReleaseValues("nfsprovisioner-"+fileShare.ID, false)
		Expect(err).NotTo(HaveOccurred())
		imageValues, ok := releaseValues["image"].(map[string]interface{})
		Expect(ok).To(BeTrue())
		Expect(imageValues).NotTo(HaveKey("repository"))
		Expect(imageValues["tag"]).To(Equal("v4.0.2"))

		// The file share name must be sanitized into a valid label value.
		storageClass := storagev1.StorageClass{}
		err = k8sClient.Get(ctx, types.NamespacedName{Name: "nfs-" + fileShare.ID}, &storageClass)
		Expect(err).NotTo(HaveOccurred())
		Expect(storageClass.Labels["fileShareName"]).To(
			Equal("legit-image.repository-evil-malicious-image.tag-latest"))

		// Cleanup
		err = k8sClient.Delete(ctx, &provisioner)
		Expect(err).NotTo(HaveOccurred())
		_, err = reconciler.Reconcile(
			ctx,
			ctrl.Request{
				NamespacedName: types.NamespacedName{
					Namespace: DefaultNamespace,
					Name:      provisioner.Name,
				}})
		Expect(err).NotTo(HaveOccurred())
	})
})
