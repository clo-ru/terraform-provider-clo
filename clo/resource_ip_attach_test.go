package clo

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloAddressAttach_basic(t *testing.T) {
	cli, err := getTestClient()
	if err != nil {
		t.Error("Error get test client ", err)
	}

	addressId, err := buildTestAddress(cli, t)
	if err != nil {
		t.Error("Error while create address ", err)
	}
	serverId, err := buildTestServer(cli, t)
	if err != nil {
		t.Error("Error while create server ", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAddressAttachDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloAddressAttachBasic(serverId, addressId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAddressAttachExists("clo_network_ip_attach.test_attach", serverId),
				),
			},
		},
	})

}

func testAccCloAddressAttachBasic(serverId string, addressId string) string {
	return fmt.Sprintf(`resource "clo_network_ip_attach" "test_attach"{
			entity_name = "server"
			entity_id = "%s"
            address_id = "%s"
	}`, serverId, addressId)
}

func testAccCheckAddressAttachExists(n string, serverId string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("volume with ID is not set")
		}
		cli := testAccProvider.Meta().(*providerMeta).v3
		addr, e := cli.GetAddress(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		if addr.AttachedTo == nil || addr.AttachedTo.ID != serverId {
			return fmt.Errorf("invalid address attachment %v", addr.AttachedTo)
		}
		return nil
	}
}

func testAccCheckAddressAttachDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*providerMeta).v3
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_network_ip_attach" {
			continue
		}
		addr, e := cli.GetAddress(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		if addr.AttachedTo != nil {
			return fmt.Errorf("attachment not deleted")
		}
		return nil
	}
	return nil
}
