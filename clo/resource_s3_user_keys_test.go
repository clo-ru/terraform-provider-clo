package clo

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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
		cli := testAccProvider.Meta().(*providerMeta).v3
		accessKey, e := cli.GetS3UserAccessKey(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		if accessKey == "" {
			return fmt.Errorf("no s3 user access key returned")
		}
		return nil
	}
}
