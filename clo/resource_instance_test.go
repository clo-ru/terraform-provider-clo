package clo

import (
	"context"
	"errors"
	"fmt"
	clo_lib "github.com/clo-ru/cloapi-go-client/v2/clo"
	cloTools "github.com/clo-ru/cloapi-go-client/v2/clo/request_tools"
	"github.com/clo-ru/cloapi-go-client/v2/services/servers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"
)

const (
	serverName = "serv"
	imageID    = "e96961dd-f038-4726-9330-ad5468ab5a3b"
)

func TestAccCloInstance_basic(t *testing.T) {
	var server = new(servers.Server)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloInstanceBasicConf(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(fmt.Sprintf("clo_compute_instance.%s", serverName), server),
				),
			},
		},
	})
}

func TestAccCloInstance_withKeypair(t *testing.T) {
	cli, err := getTestClient()
	if err != nil {
		t.Error("Error get test client ", err)
	}

	var server = new(servers.Server)
	keypair, err := buildTestKeypair(cli, t)
	if err != nil {
		t.Error("Error get test client ", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloInstanceWithKeypairConf(keypair),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(fmt.Sprintf("clo_compute_instance.%s", serverName), server),
				),
			},
		},
	})
}

func testAccCloInstanceWithKeypairConf(keypair string) string {
	return fmt.Sprintf(
		`resource "clo_compute_instance" "%s" {
  				project_id = "%s" 
  				name = "%s"
  				image_id = "%s"
  				flavor_ram = 4
  				flavor_vcpus = 2
  				block_device{
   					size = 40
   					bootable=true
   					storage_type = "volume"
  				}
  				addresses{
   					version = 4
   					external=true
   					ddos_protection=false
  				}
				keypairs = ["%s"]
	}`, serverName, os.Getenv("CLO_API_PROJECT_ID"), serverName, imageID, keypair)
}

func testAccCloInstanceBasicConf() string {
	return fmt.Sprintf(
		`resource "clo_compute_instance" "%s" {
  				project_id = "%s" 
  				name = "%s"
  				image_id = "%s"
  				flavor_ram = 4
  				flavor_vcpus = 2
  				block_device{
   					size = 40
   					bootable=true
   					storage_type = "volume"
  				}
  				addresses{
   					version = 4
   					external=true
   					ddos_protection=false
  				}
	}`, serverName, os.Getenv("CLO_API_PROJECT_ID"), serverName, imageID)
}

func testAccCheckInstanceExists(n string, serverItem *servers.Server) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("server with ID is not set")
		}
		cli := testAccProvider.Meta().(*clo_lib.ApiClient)
		req := servers.ServerDetailRequest{ServerID: rs.Primary.ID}
		resp, e := req.Do(context.Background(), cli)
		if e != nil {
			return e
		}
		*serverItem = resp.Result
		return nil
	}
}

func testAccCheckInstanceDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*clo_lib.ApiClient)
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_compute_instance" {
			continue
		}
		req := servers.ServerDetailRequest{ServerID: rs.Primary.ID}
		_, e := req.Do(context.Background(), cli)
		if e == nil {
			return fmt.Errorf("clo instance %s still exists", rs.Primary.ID)
		}

		apiError := cloTools.DefaultError{}
		if errors.As(e, &apiError) && apiError.Code == 404 {
			return nil
		}
		return e
	}
	return nil
}
