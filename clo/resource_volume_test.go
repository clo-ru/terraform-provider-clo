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
	volumeName = "volume_1"
)

func TestAccCloVolume_basic(t *testing.T) {
	var volume = new(cloapi.Volume)
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
			size = 10
	}`, volumeName, os.Getenv("CLO_API_PROJECT_ID"), volumeName)
}

func testAccCheckVolumeExists(n string, volumeItem *cloapi.Volume) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("volume with ID is not set")
		}
		cli := testAccProvider.Meta().(*providerMeta).v3
		vol, e := cli.GetVolume(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		*volumeItem = *vol
		return nil
	}
}

func testAccCheckVolumeDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*providerMeta).v3
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_disks_volume" {
			continue
		}
		_, e := cli.GetVolume(context.Background(), rs.Primary.ID)
		if cloapi.IsNotFound(e) {
			return nil
		}
		return e
	}
	return nil
}
