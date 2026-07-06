package clo

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const instancePowerName = "power_1"

func TestAccCloInstancePower_basic(t *testing.T) {
	skipIfNotAcc(t)
	cli, err := getTestClient()
	if err != nil {
		t.Fatal("Error get test client ", err)
	}
	serverID, err := buildTestServer(cli, t)
	if err != nil {
		t.Fatal("Error while create server ", err)
	}

	addr := fmt.Sprintf("clo_compute_instance_power.%s", instancePowerName)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				// A freshly built server is running; power it off.
				Config: testAccCloInstancePowerConfig(serverID, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "enabled", "false"),
					resource.TestCheckResourceAttr(addr, "switch_status", "OFF"),
				),
			},
			{
				// Power it back on.
				Config: testAccCloInstancePowerConfig(serverID, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "enabled", "true"),
					resource.TestCheckResourceAttr(addr, "switch_status", "ON"),
				),
			},
		},
	})
}

func testAccCloInstancePowerConfig(instanceID string, enabled bool) string {
	return fmt.Sprintf(`resource "clo_compute_instance_power" "%s" {
	instance_id = "%s"
	enabled     = %t
}`, instancePowerName, instanceID, enabled)
}
