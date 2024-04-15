package clo

import (
	"context"
	"fmt"
	clo_lib "github.com/clo-ru/cloapi-go-client/v2/clo"
	"github.com/clo-ru/cloapi-go-client/v2/services/storage"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAccCloS3UserKeys_basic(t *testing.T) {
	cli, err := getTestClient()
	if err != nil {
		t.Error("Error get test client ", err)
	}

	userId, err := buildTestS3user(cli, t)
	if err != nil {
		t.Error("Error while create s3 user ", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCloS3KeysBasic(userId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3KeysExists("clo_storage_s3_user_keys.test_keys", userId),
				),
			},
		},
	})

}

func testAccCloS3KeysBasic(userId string) string {
	return fmt.Sprintf(`resource "clo_storage_s3_user_keys" "test_keys"{
			user_id = "%s"
	}`, userId)
}

func testAccCheckS3KeysExists(n string, serverId string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("volume with ID is not set")
		}
		cli := testAccProvider.Meta().(*clo_lib.ApiClient)
		req := storage.S3KeysGetRequest{UserID: rs.Primary.ID}
		resp, e := req.Do(context.Background(), cli)

		if e != nil {
			return e
		}

		if len(resp.Result) != 1 {
			return fmt.Errorf("Invalid s3 user keys %v", resp.Result)
		}

		return nil
	}
}
