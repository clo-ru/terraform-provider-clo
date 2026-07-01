package clo

import (
	"context"
	"fmt"
	"testing"

	"github.com/clo-ru/terraform-provider-clo/v2/internal/cloapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const keypairName = "keypair_1"

func TestAccCloKeypair_import(t *testing.T) {
	kp := new(cloapi.Keypair)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckKeypairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloKeypairImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeypairExists(fmt.Sprintf("clo_compute_keypair.%s", keypairName), kp),
				),
			},
		},
	})
}

func TestAccCloKeypair_generate(t *testing.T) {
	kp := new(cloapi.Keypair)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccCloPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckKeypairDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloKeypairGenerate(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeypairExists(fmt.Sprintf("clo_compute_keypair.%s", keypairName), kp),
					resource.TestCheckResourceAttrSet(fmt.Sprintf("clo_compute_keypair.%s", keypairName), "private_key"),
					resource.TestCheckResourceAttrSet(fmt.Sprintf("clo_compute_keypair.%s", keypairName), "public_key"),
				),
			},
		},
	})
}

func testAccCloKeypairImport() string {
	return fmt.Sprintf(`resource "clo_compute_keypair" "%s"{
			project_id = "%s"
			name       = "%s"
			public_key = "%s"
	}`, keypairName, projectID, keypairName, testPublicKey)
}

func testAccCloKeypairGenerate() string {
	return fmt.Sprintf(`resource "clo_compute_keypair" "%s"{
			project_id = "%s"
			name       = "%s"
	}`, keypairName, projectID, keypairName)
}

func testAccCheckKeypairExists(n string, item *cloapi.Keypair) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("keypair ID is not set")
		}
		cli := testAccProvider.Meta().(*providerMeta).v3
		kp, e := cli.GetKeypair(context.Background(), rs.Primary.ID)
		if e != nil {
			return e
		}
		*item = *kp
		return nil
	}
}

func testAccCheckKeypairDestroy(st *terraform.State) error {
	cli := testAccProvider.Meta().(*providerMeta).v3
	for _, rs := range st.RootModule().Resources {
		if rs.Type != "clo_compute_keypair" {
			continue
		}
		_, e := cli.GetKeypair(context.Background(), rs.Primary.ID)
		if cloapi.IsNotFound(e) {
			return nil
		}
		return e
	}
	return nil
}
