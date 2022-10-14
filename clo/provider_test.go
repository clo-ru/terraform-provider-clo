package clo

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"os"
	"testing"
)

var (
	authKey   string
	baseUrl   string
	projectID string
)

var (
	testAccProvider  *schema.Provider
	testAccProviders map[string]func() (*schema.Provider, error)
)

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]func() (*schema.Provider, error){
		"clo": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}
	projectID = os.Getenv("CLO_API_PROJECT_ID")
	authKey = os.Getenv("CLO_API_AUTH_TOKEN")
	baseUrl = os.Getenv("CLO_API_AUTH_URL")
}

func testAccCloPreCheck(t *testing.T) {
	if _, b := os.LookupEnv("CLO_API_PROJECT_ID"); !b {
		t.Fatal("CLO_API_PROJECT_ID env should be provided")
	}
	if _, b := os.LookupEnv("CLO_API_AUTH_URL"); !b {
		t.Fatal("CLO_API_AUTH_URL env should be provided")
	}
	if _, b := os.LookupEnv("CLO_API_AUTH_TOKEN"); !b {
		t.Fatal("CLO_API_AUTH_TOKEN env should be provided")
	}
}
