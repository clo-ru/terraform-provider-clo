package clo

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

// skipIfNotAcc skips a test before it makes any real API calls unless acceptance
// testing is enabled. resource.Test() already gates on TF_ACC, but tests that build
// fixtures (servers, volumes, …) before calling it need this guard too, so plain
// `go test ./...` stays green without credentials.
func skipIfNotAcc(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance test skipped; set TF_ACC=1 (and the CLO_API_* env) to run")
	}
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
