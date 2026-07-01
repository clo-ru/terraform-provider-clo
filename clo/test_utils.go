package clo

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/google/uuid"
)

func getTestClient() (*cloapi.Client, error) {
	return cloapi.New(os.Getenv("CLO_API_AUTH_TOKEN"), os.Getenv("CLO_API_AUTH_URL"))
}

func getTestProject() string {
	return os.Getenv("CLO_API_PROJECT_ID")
}

// getTestImageID resolves a usable OS image for the test project at runtime instead
// of hard-coding an ID (image IDs differ per environment and change over time). It
// prefers a Linux image (small, no license needed) and falls back to the first
// available one.
func getTestImageID(t *testing.T, cli *cloapi.Client) string {
	t.Helper()
	images, err := cli.ListImages(context.Background(), getTestProject())
	if err != nil {
		t.Fatalf("list images: %v", err)
	}
	if len(images) == 0 {
		t.Fatalf("no images available in project %s", getTestProject())
	}
	for _, im := range images {
		if strings.EqualFold(im.OSFamily, "linux") {
			return im.ID
		}
	}
	return images[0].ID
}

// Server
func buildTestServer(cli *cloapi.Client, t *testing.T) (string, error) {
	id, err := cli.CreateServer(context.Background(), cloapi.ServerCreateParams{
		ProjectID:   getTestProject(),
		Name:        "testServer",
		ImageID:     getTestImageID(t, cli),
		FlavorRam:   2,
		FlavorVcpus: 1,
		Storages:    []cloapi.ServerStorage{{Bootable: true, StorageType: "local", Size: 10}},
		Addresses:   []cloapi.ServerAddress{{External: false, Version: 4}},
	})
	if err != nil {
		return "", err
	}

	t.Cleanup(func() {
		t.Logf("Cleanup test server %s", id)
		if err := destroyTestServer(id, cli); err != nil {
			t.Log("Error on cleanup server")
		}
	})

	if err := waitInstanceState(context.Background(), id, cli, []string{creatingInstance}, []string{activeInstance}, 20*time.Minute); err != nil {
		return "", err
	}
	return id, nil
}

func destroyTestServer(serverId string, cli *cloapi.Client) error {
	if err := cli.DeleteServer(context.Background(), serverId, nil, nil); err != nil {
		return err
	}
	return waitInstanceDeleted(context.Background(), serverId, cli, 20*time.Minute)
}

// Volume

func buildTestVolume(cli *cloapi.Client, t *testing.T) (string, error) {
	id, err := cli.CreateVolume(context.Background(), cloapi.VolumeCreateParams{
		ProjectID:  getTestProject(),
		Name:       "test_volume",
		Size:       10,
		Autorename: true,
	})
	if err != nil {
		return "", err
	}

	t.Cleanup(func() {
		t.Logf("Cleanup test volume %s", id)
		if err := destroyTestVolume(id, cli); err != nil {
			t.Log("Error on cleanup volume")
		}
	})

	if err := waitVolumeState(context.Background(), id, cli, []string{creatingVolume}, []string{activeVolume}, 10*time.Minute); err != nil {
		return "", err
	}
	return id, nil
}

func destroyTestVolume(volumeId string, cli *cloapi.Client) error {
	if err := cli.DeleteVolume(context.Background(), volumeId); err != nil {
		return err
	}
	return waitVolumeDeleted(context.Background(), volumeId, cli, 10*time.Minute)
}

// NOTE: IP/address acceptance fixtures were intentionally removed. The service
// forbids deleting an address within 7 days of creation, so a create+destroy test
// cycle is impossible. The IP resource/data-source code itself is unchanged.

// S3 user

func buildTestS3user(cli *cloapi.Client, t *testing.T) (string, error) {
	testName := strings.ReplaceAll(uuid.NewString(), "-", "")
	id, err := cli.CreateS3User(context.Background(), cloapi.S3UserCreateParams{
		ProjectID:           getTestProject(),
		Name:                testName,
		CanonicalName:       testName,
		UserQuotaMaxSize:    10,
		UserQuotaMaxObjects: 10,
	})
	if err != nil {
		return "", err
	}

	t.Cleanup(func() {
		t.Logf("Cleanup test s3 user %s", id)
		if err := destroyTestS3User(id, cli); err != nil {
			t.Log("Error on cleanup s3 user ", err)
		}
	})

	if err := waitS3UserState(context.Background(), id, cli, []string{s3UserCreating}, []string{s3UserActive}, 10*time.Minute); err != nil {
		return "", err
	}
	return id, nil
}

func destroyTestS3User(id string, cli *cloapi.Client) error {
	if err := cli.DeleteS3User(context.Background(), id); err != nil {
		return err
	}
	return waitS3UserDeleted(context.Background(), id, cli, 10*time.Minute)
}

// keypair

const testPublicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDptRaDWlcd/LHGOTmtc3/hetEacftO2wPjLxahl/el7L32B3Nw1ATOYnsL+hHd4Acx7BVJd7j6ciwE5MYxiCvynGp3TrXVlcqUjW6SBfZNlFy2mgCIW70wfIN1usXQDuNAUbfdyF5qaQxWN38WL8CsoO3DBU2oGgeJko9CdLtEaxU7QJEfcKIq6sBeLmpB+TELJpnaACxUF7aq1V/YPx+wZFZeqnlf0V5blQ/Yo+bBncFChP8xjmmu5ckfuiNHfEWBq+RYFytWt03mC/eB0K+b8IQlcaYSh58jVExTlBmjaizqOT1j8Ahc3RewOgALez7//+c3HI+z9ryrOOymZC3B"

func buildTestKeypair(cli *cloapi.Client, t *testing.T) (string, error) {
	testName := strings.ReplaceAll(uuid.NewString(), "-", "")
	id, err := cli.ImportKeypair(context.Background(), getTestProject(), testName, testPublicKey)
	if err != nil {
		return "", err
	}

	t.Cleanup(func() {
		t.Logf("Cleanup test keypair %s", id)
		if err := destroyTestKeypair(id, cli); err != nil {
			t.Log("Error on cleanup keypair ", err)
		}
	})

	return id, nil
}

func destroyTestKeypair(id string, cli *cloapi.Client) error {
	return cli.DeleteKeypair(context.Background(), id)
}
