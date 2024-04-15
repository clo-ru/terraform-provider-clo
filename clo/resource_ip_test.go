package clo

import (
	"context"
	"errors"
	"fmt"
	clo_lib "github.com/clo-ru/cloapi-go-client/v2/clo"
	cloTools "github.com/clo-ru/cloapi-go-client/v2/clo/request_tools"
	clo_ip "github.com/clo-ru/cloapi-go-client/v2/services/ip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"
)

const (
	ipName = "fip_1"
	ptr    = "serv.ru"
)

func TestAccCloIP_basic(t *testing.T) {
	var ip = new(clo_ip.Address)
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

func testAccCheckIPExists(n string, ipItem *clo_ip.Address) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("ip with ID is not set")
		}
		cli := testAccProvider.Meta().(*clo_lib.ApiClient)
		req := clo_ip.AddressDetailRequest{
			AddressID: rs.Primary.ID,
		}
		resp, e := req.Do(context.Background(), cli)
		if e != nil {
			return e
		}
		*ipItem = resp.Result
		return nil
	}
}

func testAccCheckIPDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*clo_lib.ApiClient)
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_network_ip" {
			continue
		}
		req := clo_ip.AddressDetailRequest{AddressID: rs.Primary.ID}
		_, e := req.Do(context.Background(), cli)

		apiError := cloTools.DefaultError{}
		if errors.As(e, &apiError) && apiError.Code == 404 {
			return nil
		}
		return e
	}
	return nil
}
