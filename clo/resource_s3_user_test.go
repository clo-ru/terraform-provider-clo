package clo

import (
	"context"
	"fmt"
	clo_lib "github.com/clo-ru/cloapi-go-client/clo"
	"github.com/clo-ru/cloapi-go-client/services/storage"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"net/http"
	"os"
	"testing"
)

const (
	userName = "s3_user"
)

func TestAccCloS3User_basic(t *testing.T) {
	var s3User = new(storage.ResponseItem)
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

func testAccCheckS3UserExists(n string, s3UserItem *storage.ResponseItem) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("user with ID is not set")
		}
		cli := testAccProvider.Meta().(*clo_lib.ApiClient)
		req := storage.S3UserDetailRequest{
			UserID: rs.Primary.ID,
		}
		resp, e := req.Make(context.Background(), cli)
		if e != nil {
			return e
		}
		if resp.Code != 200 {
			return fmt.Errorf("http code 200 expected, got %s", http.StatusText(resp.Code))
		}
		*s3UserItem = resp.Result
		return nil
	}
}

func testAccCheckS3UserDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*clo_lib.ApiClient)
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_storage_s3_user" {
			continue
		}
		req := storage.S3UserDetailRequest{
			UserID: rs.Primary.ID,
		}
		resp, e := req.Make(context.Background(), cli)
		if e != nil {
			return e
		}
		if resp.Code != 404 {
			return fmt.Errorf("clo s3User %s still exists", rs.Primary.ID)
		}
	}
	return nil
}
