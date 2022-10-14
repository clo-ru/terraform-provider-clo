package clo

import (
	"context"
	"fmt"
	clo_lib "github.com/clo-ru/cloapi-go-client/clo"
	"github.com/clo-ru/cloapi-go-client/services/disks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"net/http"
	"os"
	"testing"
)

const (
	volumeName = "volume_1"
)

func TestAccCloVolume_basic(t *testing.T) {
	var volume = new(disks.VolumeDetail)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVolumeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloVolumeBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeExists(fmt.Sprintf("clo_disks_volume.%s", volumeName), volume),
				),
			},
		},
	})
}

func testAccCloVolumeBasic() string {
	return fmt.Sprintf(`resource "clo_disks_volume" "%s"{
			project_id = "%s"
			name = "%s"
			size = 30
	}`, volumeName, os.Getenv("CLO_API_PROJECT_ID"), volumeName)
}

func testAccCheckVolumeExists(n string, volumeItem *disks.VolumeDetail) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("volume with ID is not set")
		}
		cli := testAccProvider.Meta().(*clo_lib.ApiClient)
		req := disks.VolumeDetailRequest{
			VolumeID: rs.Primary.ID,
		}
		resp, e := req.Make(context.Background(), cli)
		if e != nil {
			return e
		}
		if resp.Code != 200 {
			return fmt.Errorf("http code 200 expected, got %s", http.StatusText(resp.Code))
		}
		*volumeItem = resp.Result
		return nil
	}
}

func testAccCheckVolumeDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*clo_lib.ApiClient)
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_disks_volume" {
			continue
		}
		req := disks.VolumeDetailRequest{
			VolumeID: rs.Primary.ID,
		}
		resp, e := req.Make(context.Background(), cli)
		if e != nil {
			return e
		}
		if resp.Code != 404 {
			return fmt.Errorf("clo volume %s still exists", rs.Primary.ID)
		}
	}
	return nil
}
