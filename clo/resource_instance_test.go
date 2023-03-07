package clo

import (
	"context"
	"fmt"
	clo_lib "github.com/clo-ru/cloapi-go-client/clo"
	"github.com/clo-ru/cloapi-go-client/services/servers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"net/http"
	"os"
	"testing"
)

const (
	serverName = "serv"
)

func TestAccCloInstance_basic(t *testing.T) {
	var server = new(servers.ServerDetailItem)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloInstanceBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(fmt.Sprintf("clo_compute_instance.%s", serverName), server),
				),
			},
		},
	})
}

func testAccCloInstanceBasic() string {
	return fmt.Sprintf(`
			data "clo_project_image" "ubuntu"{
				project_id = "%s" 
 				name = "Ubuntu 20"
			}
			resource "clo_compute_instance" "%s" {
  				project_id = "%s" 
  				name = "%s"
  				image_id = data.clo_project_image.ubuntu.image_id
  				flavor_ram = 4
  				flavor_vcpus = 2
  				block_device{
   					size = 40
   					bootable=true
   					storage_type = "volume"
  				}
  				addresses{
   						version = 4
   						external=false
   						ddos_protection=false
  				}
	}`, os.Getenv("CLO_API_PROJECT_ID"), serverName, os.Getenv("CLO_API_PROJECT_ID"), serverName)
}

func testAccCheckInstanceExists(n string, serverItem *servers.ServerDetailItem) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("server with ID is not set")
		}
		cli := testAccProvider.Meta().(*clo_lib.ApiClient)
		req := servers.ServerDetailRequest{
			ServerID: rs.Primary.ID,
		}
		resp, e := req.Make(context.Background(), cli)
		if e != nil {
			return e
		}
		if resp.Code != 200 {
			return fmt.Errorf("http code 200 expected, got %s", http.StatusText(resp.Code))
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
		req := servers.ServerDetailRequest{
			ServerID: rs.Primary.ID,
		}
		resp, e := req.Make(context.Background(), cli)
		if e != nil {
			return e
		}
		if resp.Code != 404 {
			return fmt.Errorf("clo instance %s still exists", rs.Primary.ID)
		}
	}
	return nil
}
