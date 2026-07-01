package clo

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloVolumeAttach_basic(t *testing.T) {
	skipIfNotAcc(t)
	cli, err := getTestClient()
	if err != nil {
		t.Error("Error get test client ", err)
	}

	volumeId, err := buildTestVolume(cli, t)
	if err != nil {
		t.Error("Error while create volume ", err)
	}
	serverId, err := buildTestServer(cli, t)
	if err != nil {
		t.Error("Error while create server ", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVolumeAttachDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloVolumeAttachBasic(volumeId, serverId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVolumeAttachExists("clo_disks_volume_attach.test_attach", serverId),
				),
			},
		},
	})

}

func testAccCloVolumeAttachBasic(volumeId, serverId string) string {
	return fmt.Sprintf(`resource "clo_disks_volume_attach" "test_attach"{
			volume_id = "%s"
			instance_id = "%s"
	}`, volumeId, serverId)
}

func testAccCheckVolumeAttachExists(n string, serverId string) resource.TestCheckFunc {
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
		if vol.Attachment == nil || vol.Attachment.ID != serverId {
			return fmt.Errorf("invalid volume attachment %v", vol.Attachment)
		}
		return nil
	}
}

func testAccCheckVolumeAttachDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*providerMeta).v3
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_disks_volume_attach" {
			continue
		}
		vol, e := cli.GetVolume(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		if vol.Attachment != nil {
			return fmt.Errorf("attachment not deleted")
		}
		return nil
	}
	return nil
}
