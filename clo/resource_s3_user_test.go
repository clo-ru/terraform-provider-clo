package clo

import (
	"context"
	"errors"
	"fmt"
	clo_lib "github.com/clo-ru/cloapi-go-client/v2/clo"
	cloTools "github.com/clo-ru/cloapi-go-client/v2/clo/request_tools"
	"github.com/clo-ru/cloapi-go-client/v2/services/storage"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"
)

const (
	userName = "s3_user"
)

func TestAccCloS3User_basic(t *testing.T) {
	var s3User = new(storage.S3User)
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

func testAccCheckS3UserExists(n string, s3UserItem *storage.S3User) resource.TestCheckFunc {
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
		resp, e := req.Do(context.Background(), cli)
		if e != nil {
			return e
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
		_, e := req.Do(context.Background(), cli)
		apiError := cloTools.DefaultError{}
		if errors.As(e, &apiError) && apiError.Code == 404 {
			return nil
		}
		return e
	}
	return nil
}
