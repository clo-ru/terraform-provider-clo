package cloapi

import (
	"testing"
	"time"

	gen "github.com/clo-ru/cloapi-go-client/v3"
)

func strptr(s string) *string { return &s }
func boolptr(b bool) *bool    { return &b }

func TestLoadBalancerFromSchema(t *testing.T) {
	created := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	updated := time.Date(2024, 6, 7, 8, 9, 10, 0, time.UTC)

	t.Run("full", func(t *testing.T) {
		src := &gen.LBDetailResponseSchema{
			Id:                 "lb-1",
			Name:               "web",
			Project:            "proj-1",
			Status:             "ACTIVE",
			SwitchStatus:       "ON",
			Algorithm:          "ROUND_ROBIN",
			SessionPersistence: true,
			RulesCount:         3,
			Addresses:          []string{"a1", "a2"},
			CreatedIn:          created,
			UpdatedIn:          updated,
			Healthmonitor: gen.HealthmonitorDetailResponseSchema{
				Type:          "HTTP",
				Delay:         10,
				Timeout:       5,
				MaxRetries:    3,
				HttpMethod:    strptr("GET"),
				UrlPath:       strptr("/health"),
				ExpectedCodes: strptr("200"),
			},
		}
		got := loadBalancerFromSchema(src)

		if got.ID != "lb-1" || got.Name != "web" || got.Project != "proj-1" {
			t.Errorf("identity fields wrong: %+v", got)
		}
		if got.Status != "ACTIVE" || got.SwitchStatus != "ON" {
			t.Errorf("status/switch wrong: %q/%q", got.Status, got.SwitchStatus)
		}
		if got.Algorithm != "ROUND_ROBIN" || !got.SessionPersistence || got.RulesCount != 3 {
			t.Errorf("algorithm/session/rules wrong: %+v", got)
		}
		if got.CreatedIn != "2024-01-02T03:04:05Z" || got.UpdatedIn != "2024-06-07T08:09:10Z" {
			t.Errorf("timestamps wrong: %q / %q", got.CreatedIn, got.UpdatedIn)
		}
		if len(got.Addresses) != 2 || got.Addresses[0] != "a1" || got.Addresses[1] != "a2" {
			t.Errorf("addresses wrong: %v", got.Addresses)
		}
		hm := got.Healthmonitor
		if hm.Type != "HTTP" || hm.Delay != 10 || hm.Timeout != 5 || hm.MaxRetries != 3 {
			t.Errorf("healthmonitor scalars wrong: %+v", hm)
		}
		if hm.HttpMethod != "GET" || hm.UrlPath != "/health" || hm.ExpectedCodes != "200" {
			t.Errorf("healthmonitor http fields wrong: %+v", hm)
		}

		// Addresses must be a copy, not aliased to the source slice.
		src.Addresses[0] = "mutated"
		if got.Addresses[0] != "a1" {
			t.Errorf("addresses aliased source slice: %v", got.Addresses)
		}
	})

	t.Run("minimal_nil_pointers", func(t *testing.T) {
		src := &gen.LBDetailResponseSchema{
			Id:        "lb-2",
			Status:    "STOPPED",
			CreatedIn: created,
			UpdatedIn: updated,
			Healthmonitor: gen.HealthmonitorDetailResponseSchema{
				Type: "PING",
			},
		}
		got := loadBalancerFromSchema(src)

		if got.Addresses != nil {
			t.Errorf("expected nil addresses, got %v", got.Addresses)
		}
		hm := got.Healthmonitor
		if hm.HttpMethod != "" || hm.UrlPath != "" || hm.ExpectedCodes != "" {
			t.Errorf("nil healthmonitor pointers should map to empty strings: %+v", hm)
		}
	})
}

func TestRuleFromSchema(t *testing.T) {
	src := &gen.RuleDetailResponseSchema{
		Id:                   "rule-1",
		Loadbalancer:         "lb-1",
		Address:              "203.0.113.5",
		Server:               "srv-1",
		Status:               "ACTIVE",
		ExternalProtocolPort: 80,
		InternalProtocolPort: 8080,
	}
	got := ruleFromSchema(src)
	if got.ID != "rule-1" || got.Loadbalancer != "lb-1" || got.Address != "203.0.113.5" ||
		got.Server != "srv-1" || got.Status != "ACTIVE" ||
		got.ExternalProtocolPort != 80 || got.InternalProtocolPort != 8080 {
		t.Errorf("ruleFromSchema mapping wrong: %+v", got)
	}
}

func TestLoadBalancerCreateBody(t *testing.T) {
	t.Run("minimal_omits_optionals", func(t *testing.T) {
		body := loadBalancerCreateBody(LoadBalancerCreateParams{
			Name: "web",
			Healthmonitor: HealthmonitorParams{
				Type: "PING", Delay: 10, Timeout: 5, MaxRetries: 3,
			},
		})

		if body.Name != "web" {
			t.Errorf("name wrong: %q", body.Name)
		}
		if body.Algorithm != nil {
			t.Errorf("algorithm should be omitted, got %v", *body.Algorithm)
		}
		if body.SessionPersistence != nil {
			t.Errorf("session_persistence should be omitted, got %v", *body.SessionPersistence)
		}
		if body.Address != nil {
			t.Errorf("address should be omitted, got %+v", body.Address)
		}
		if string(body.Healthmonitor.Type) != "PING" || body.Healthmonitor.Delay != 10 {
			t.Errorf("healthmonitor scalars wrong: %+v", body.Healthmonitor)
		}
		if body.Healthmonitor.HttpMethod != nil || body.Healthmonitor.UrlPath != nil ||
			body.Healthmonitor.ExpectedCodes != nil {
			t.Errorf("non-HTTP healthmonitor should omit http fields: %+v", body.Healthmonitor)
		}
		// rules is always sent as an explicit empty list (the API errors on a null rules key).
		if body.Rules == nil || len(*body.Rules) != 0 {
			t.Errorf("rules should be an empty non-nil list, got %v", body.Rules)
		}
	})

	t.Run("full", func(t *testing.T) {
		body := loadBalancerCreateBody(LoadBalancerCreateParams{
			Name:               "web",
			Algorithm:          "ROUND_ROBIN",
			SessionPersistence: boolptr(true),
			AddressID:          "addr-1",
			Healthmonitor: HealthmonitorParams{
				Type: "HTTP", Delay: 80, Timeout: 15, MaxRetries: 3,
				HttpMethod: "GET", UrlPath: "/health", ExpectedCodes: "200",
			},
		})

		if body.Algorithm == nil || string(*body.Algorithm) != "ROUND_ROBIN" {
			t.Errorf("algorithm wrong: %v", body.Algorithm)
		}
		if body.SessionPersistence == nil || !*body.SessionPersistence {
			t.Errorf("session_persistence wrong: %v", body.SessionPersistence)
		}
		if body.Address == nil || body.Address.Id == nil || *body.Address.Id != "addr-1" {
			t.Errorf("address id wrong: %+v", body.Address)
		}
		if body.Address.DdosProtection != nil {
			t.Errorf("ddos_protection must be omitted for an existing address id, got %v", *body.Address.DdosProtection)
		}
		hm := body.Healthmonitor
		if hm.HttpMethod == nil || string(*hm.HttpMethod) != "GET" {
			t.Errorf("http_method wrong: %v", hm.HttpMethod)
		}
		if hm.UrlPath == nil || *hm.UrlPath != "/health" {
			t.Errorf("url_path wrong: %v", hm.UrlPath)
		}
		if hm.ExpectedCodes == nil || *hm.ExpectedCodes != "200" {
			t.Errorf("expected_codes wrong: %v", hm.ExpectedCodes)
		}
	})

	t.Run("address_ddos_only_no_id", func(t *testing.T) {
		body := loadBalancerCreateBody(LoadBalancerCreateParams{
			Name:        "web",
			AddressDdos: boolptr(false),
			Healthmonitor: HealthmonitorParams{
				Type: "TCP", Delay: 10, Timeout: 5, MaxRetries: 3,
			},
		})
		if body.Address == nil {
			t.Fatalf("address block should be present when only ddos is set")
		}
		if body.Address.Id != nil {
			t.Errorf("address id should be omitted, got %q", *body.Address.Id)
		}
		if body.Address.DdosProtection == nil || *body.Address.DdosProtection {
			t.Errorf("address ddos should be false, got %v", body.Address.DdosProtection)
		}
	})
}

func TestHealthmonitorUpdateBody(t *testing.T) {
	t.Run("non_http_omits_http_fields", func(t *testing.T) {
		body := healthmonitorUpdateBody(HealthmonitorParams{
			Type: "TCP", Delay: 10, Timeout: 5, MaxRetries: 3,
		})
		if string(body.Type) != "TCP" || body.Delay != 10 || body.Timeout != 5 || body.MaxRetries != 3 {
			t.Errorf("scalars wrong: %+v", body)
		}
		if body.HttpMethod != nil || body.UrlPath != nil || body.ExpectedCodes != nil {
			t.Errorf("http fields should be omitted: %+v", body)
		}
	})

	t.Run("http_sets_fields", func(t *testing.T) {
		body := healthmonitorUpdateBody(HealthmonitorParams{
			Type: "HTTP", Delay: 10, Timeout: 5, MaxRetries: 3,
			HttpMethod: "POST", UrlPath: "/ping", ExpectedCodes: "200-299",
		})
		if body.HttpMethod == nil || string(*body.HttpMethod) != "POST" {
			t.Errorf("http_method wrong: %v", body.HttpMethod)
		}
		if body.UrlPath == nil || *body.UrlPath != "/ping" {
			t.Errorf("url_path wrong: %v", body.UrlPath)
		}
		if body.ExpectedCodes == nil || *body.ExpectedCodes != "200-299" {
			t.Errorf("expected_codes wrong: %v", body.ExpectedCodes)
		}
	})
}
