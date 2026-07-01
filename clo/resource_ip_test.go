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
	ipName = "fip_1"
	ptr    = "serv.ru"
)

func TestAccCloIP_basic(t *testing.T) {
	var ip = new(cloapi.Address)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloIPBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIPExists(fmt.Sprintf("clo_network_ip.%s", ipName), ip),
				),
			},
		},
	})
}

func testAccCloIPBasic() string {
	return fmt.Sprintf(`resource "clo_network_ip" "fip_1" {
		project_id = "%s"
		ptr = "%s"
	}`, os.Getenv("CLO_API_PROJECT_ID"), ptr)
}

func testAccCheckIPExists(n string, ipItem *cloapi.Address) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("ip with ID is not set")
		}
		cli := testAccProvider.Meta().(*providerMeta).v3
		addr, e := cli.GetAddress(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		*ipItem = *addr
		return nil
	}
}

func testAccCheckIPDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*providerMeta).v3
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_network_ip" {
			continue
		}
		_, e := cli.GetAddress(context.Background(), rs.Primary.ID)
		if cloapi.IsNotFound(e) {
			return nil
		}
		return e
	}
	return nil
}
