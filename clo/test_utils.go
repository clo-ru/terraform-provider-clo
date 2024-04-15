package clo

import (
	"context"
	clo_lib "github.com/clo-ru/cloapi-go-client/v2/clo"
	"github.com/clo-ru/cloapi-go-client/v2/services/disks"
	"github.com/clo-ru/cloapi-go-client/v2/services/ip"
	"github.com/clo-ru/cloapi-go-client/v2/services/servers"
	"github.com/clo-ru/cloapi-go-client/v2/services/storage"
	"github.com/google/uuid"
	"os"
	"strings"
	"testing"
	"time"
)

const testImageID = "44262267-5f2e-4802-acc1-3939f7ae7b9c"

func getTestClient() (*clo_lib.ApiClient, error) {
	authKey := os.Getenv("CLO_API_AUTH_TOKEN")
	baseUrl := os.Getenv("CLO_API_AUTH_URL")
	return clo_lib.NewDefaultClient(authKey, baseUrl)
}

func getTestProject() string {
	return os.Getenv("CLO_API_PROJECT_ID")
}

// Server
func buildTestServer(cli *clo_lib.ApiClient, t *testing.T) (string, error) {
	serverReq := servers.ServerCreateRequest{
		ProjectID: getTestProject(),
		Body: servers.ServerCreateBody{
			Name:      "testServer",
			Image:     testImageID,
			Flavor:    servers.ServerFlavorBody{Ram: 2, Vcpus: 1},
			Storages:  []servers.ServerStorageBody{{10, true, "local"}},
			Addresses: []servers.ServerAddressesBody{{External: false, Version: 4}},
		},
	}
	res, err := serverReq.Do(context.Background(), cli)
	if err != nil {
		return "", err
	}

	t.Cleanup(func() {
		t.Logf("Cleanup test server %s", res.Result.ID)
		if err := destroyTestServer(res.Result.ID, cli); err != nil {
			t.Log("Error on cleanup server")
		}
	})

	if err := waitInstanceState(context.Background(), res.Result.ID, cli, []string{creatingInstance}, []string{activeInstance}, 20*time.Minute); err != nil {
		return "", err
	}
	return res.Result.ID, nil
}

func destroyTestServer(serverId string, cli *clo_lib.ApiClient) error {
	req := servers.ServerDeleteRequest{ServerID: serverId, Body: servers.ServerDeleteBody{}}
	if err := req.Do(context.Background(), cli); err != nil {
		return err
	}

	return waitInstanceDeleted(context.Background(), serverId, cli, 20*time.Minute)
}

// Volume

func buildTestVolume(cli *clo_lib.ApiClient, t *testing.T) (string, error) {
	volumeReq := disks.VolumeCreateRequest{ProjectID: getTestProject(), Body: disks.VolumeCreateBody{
		"test_volume", 10, true}}
	volumeRes, err := volumeReq.Do(context.Background(), cli)
	if err != nil {
		return "", err
	}

	t.Cleanup(func() {
		t.Logf("Cleanup test volume %s", volumeRes.Result.ID)
		if err := destroyTestVolume(volumeRes.Result.ID, cli); err != nil {
			t.Log("Error on cleanup volume")
		}
	})

	if _, err := waitVolumeState(context.Background(), volumeRes.Result.ID, cli, []string{creatingVolume}, []string{activeVolume}, 10*time.Minute); err != nil {
		return "", err
	}
	return volumeRes.Result.ID, nil
}

func destroyTestVolume(volumeId string, cli *clo_lib.ApiClient) error {
	req := disks.VolumeDeleteRequest{VolumeID: volumeId}
	if err := req.Do(context.Background(), cli); err != nil {
		return err
	}
	return waitVolumeDeleted(context.Background(), volumeId, cli, 10*time.Minute)
}

// Address

func buildTestAddress(cli *clo_lib.ApiClient, t *testing.T) (string, error) {
	volumeReq := ip.AddressCreateRequest{ProjectID: getTestProject(), Body: ip.AddressCreateBody{}}
	res, err := volumeReq.Do(context.Background(), cli)
	if err != nil {
		return "", err
	}

	t.Cleanup(func() {
		t.Logf("Cleanup test address %s", res.Result.ID)
		if err := destroyTestAddress(res.Result.ID, cli); err != nil {
			t.Log("Error on cleanup address ", err)
		}
	})

	if _, err := waitAddressState(context.Background(), res.Result.ID, cli, []string{processingIp}, []string{detachedIp}, 10*time.Minute); err != nil {
		return "", err
	}
	return res.Result.ID, nil
}

func destroyTestAddress(id string, cli *clo_lib.ApiClient) error {
	req := ip.AddressDeleteRequest{AddressID: id}
	if err := req.Do(context.Background(), cli); err != nil {
		return err
	}
	return waitAddressDeleted(context.Background(), id, cli, 10*time.Minute)
}

// S3 user

func buildTestS3user(cli *clo_lib.ApiClient, t *testing.T) (string, error) {
	testName := strings.ReplaceAll(uuid.NewString(), "-", "")
	volumeReq := storage.S3UserCreateRequest{
		ProjectID: getTestProject(),
		Body: storage.S3UserCreateBody{
			Name:          testName,
			CanonicalName: testName,
			UserQuota:     storage.CreateQuotaParams{MaxObjects: 10, MaxSize: 10},
		},
	}
	res, err := volumeReq.Do(context.Background(), cli)
	if err != nil {
		return "", err
	}

	t.Cleanup(func() {
		t.Logf("Cleanup test s3 user %s", res.Result.ID)
		if err := destroyTestS3User(res.Result.ID, cli); err != nil {
			t.Log("Error on cleanup s3 user ", err)
		}
	})

	if _, err := waitS3UserState(context.Background(), res.Result.ID, cli, []string{s3UserCreating}, []string{s3UserActive}, 10*time.Minute); err != nil {
		return "", err
	}
	return res.Result.ID, nil
}

func destroyTestS3User(id string, cli *clo_lib.ApiClient) error {
	req := storage.S3UserDeleteRequest{UserID: id}
	if err := req.Do(context.Background(), cli); err != nil {
		return err
	}
	return waitS3UserDeleted(context.Background(), id, cli, 10*time.Minute)
}
