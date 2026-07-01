package clo

import (
	"context"
	"fmt"
	"testing"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const vrouterName = "vrouter_1"

func TestAccCloVrouter_basic(t *testing.T) {
	vr := new(cloapi.Vrouter)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVrouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloVrouterBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVrouterExists(fmt.Sprintf("clo_network_vrouter.%s", vrouterName), vr),
					resource.TestCheckResourceAttr(fmt.Sprintf("clo_network_vrouter.%s", vrouterName), "enabled", "true"),
					resource.TestCheckResourceAttrSet(fmt.Sprintf("clo_network_vrouter.%s", vrouterName), "status"),
				),
			},
		},
	})
}

func testAccCloVrouterBasic() string {
	return fmt.Sprintf(`resource "clo_network_vrouter" "%s"{
			project_id = "%s"
			name       = "%s"
	}`, vrouterName, projectID, vrouterName)
}

func testAccCheckVrouterExists(n string, item *cloapi.Vrouter) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("vrouter ID is not set")
		}
		cli := testAccProvider.Meta().(*providerMeta).v3
		vr, e := cli.GetVrouter(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		*item = *vr
		return nil
	}
}

func testAccCheckVrouterDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*providerMeta).v3
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_network_vrouter" {
			continue
		}
		_, e := cli.GetVrouter(context.Background(), rs.Primary.ID)
		if cloapi.IsNotFound(e) {
			continue
		}
		if e != nil {
			return e
		}
		return fmt.Errorf("vrouter %s still exists", rs.Primary.ID)
	}
	return nil
}
