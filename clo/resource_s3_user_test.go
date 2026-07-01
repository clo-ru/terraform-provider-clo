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
	userName = "s3_user"
)

func TestAccCloS3User_basic(t *testing.T) {
	var s3User = new(cloapi.S3User)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckS3UserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloS3UserBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckS3UserExists(fmt.Sprintf("clo_storage_s3_user.%s", userName), s3User),
				),
			},
		},
	})
}

func testAccCloS3UserBasic() string {
	return fmt.Sprintf(`resource "clo_storage_s3_user" "%s"{
 		project_id="%s"
 		canonical_name="%s"
 		max_buckets=2
 		user_quota_max_size=30
	}`, userName, os.Getenv("CLO_API_PROJECT_ID"), userName)
}

func testAccCheckS3UserExists(n string, s3UserItem *cloapi.S3User) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("user with ID is not set")
		}
		cli := testAccProvider.Meta().(*providerMeta).v3
		user, e := cli.GetS3User(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		*s3UserItem = *user
		return nil
	}
}

func testAccCheckS3UserDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*providerMeta).v3
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_storage_s3_user" {
			continue
		}
		_, e := cli.GetS3User(context.Background(), rs.Primary.ID)
		if cloapi.IsNotFound(e) {
			return nil
		}
		return e
	}
	return nil
}
