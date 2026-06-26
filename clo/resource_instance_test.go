package clo

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	serverName = "serv"
	imageID    = "e96961dd-f038-4726-9330-ad5468ab5a3b"
)

func TestAccCloInstance_basic(t *testing.T) {
	var server = new(cloapi.Server)

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

	var server = new(cloapi.Server)
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

func testAccCheckInstanceExists(n string, serverItem *cloapi.Server) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("server with ID is not set")
		}
		cli := testAccProvider.Meta().(*providerMeta).v3
		srv, e := cli.GetServer(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		*serverItem = *srv
		return nil
	}
}

func testAccCheckInstanceDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*providerMeta).v3
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_compute_instance" {
			continue
		}
		_, e := cli.GetServer(context.Background(), rs.Primary.ID)
		if e == nil {
			return fmt.Errorf("clo instance %s still exists", rs.Primary.ID)
		}
		if cloapi.IsNotFound(e) {
			return nil
		}
		return e
	}
	return nil
}
