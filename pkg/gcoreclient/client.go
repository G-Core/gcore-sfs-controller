package gcoreclient

import (
	"strings"

	crdv1 "github.com/G-Core/gcore-sfs-controller/api/v1"
	gcorecloud "github.com/G-Core/gcorelabscloud-go"
	cloudclient "github.com/G-Core/gcorelabscloud-go/gcore"
	"github.com/G-Core/gcorelabscloud-go/gcore/file_share/v1/file_shares"
)

const NfsProtocolName = "nfs"

type FileShareLister interface {
	ListFileShares(provisioner *crdv1.NfsProvisioner) ([]file_shares.FileShare, error)
}

type FileShareClient struct{}

func newApiTokenClient(provisioner *crdv1.NfsProvisioner, endpoint string, version string) (*gcorecloud.ServiceClient, error) {
	settings := gcorecloud.APITokenAPISettings{
		APIURL:   provisioner.Spec.APIURL,
		APIToken: provisioner.Spec.APIToken,
		Type:     "",
		Name:     endpoint,
		Region:   provisioner.Spec.RegionID,
		Project:  provisioner.Spec.ProjectID,
		Version:  version,
		Debug:    false,
	}
	return cloudclient.APITokenClientServiceWithDebug(settings.ToAPITokenOptions(), settings.ToEndpointOptions(), settings.Debug)
}

func (c FileShareClient) ListFileShares(provisioner *crdv1.NfsProvisioner) ([]file_shares.FileShare, error) {
	fileShareClient, err := newApiTokenClient(provisioner, "file_shares", "v1")
	if err != nil {
		return []file_shares.FileShare{}, err
	}
	allProjectFileShares, err := file_shares.ListAll(fileShareClient)
	if err != nil {
		return []file_shares.FileShare{}, err
	}
	nfsFileShares := []file_shares.FileShare{}
	for _, fileShare := range allProjectFileShares {
		if strings.ToLower(fileShare.Protocol) == NfsProtocolName {
			nfsFileShares = append(nfsFileShares, fileShare)
		}
	}
	return nfsFileShares, nil
}

type MockFileShareClient struct {
	FileShares []file_shares.FileShare
}

func (m MockFileShareClient) ListFileShares(provisioner *crdv1.NfsProvisioner) ([]file_shares.FileShare, error) {
	return m.FileShares, nil
}
