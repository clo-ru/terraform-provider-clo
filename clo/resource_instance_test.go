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
	imageID    = "44262267-5f2e-4802-acc1-3939f7ae7b9c"
)

func TestAccCloInstance_basic(t *testing.T) {
	var server = new(servers.Server)
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
