package cloapi

import (
	"context"
	"errors"
	"time"

	gen "github.com/clo-ru/cloapi-go-client/v3"
)

// LoadBalancer is the provider-facing view of a load balancer. Status is the
// lifecycle state; SwitchStatus is the ON/OFF power state toggled by Enable/Stop.
type LoadBalancer struct {
	ID                 string
	Name               string
	Project            string
	Status             string
	SwitchStatus       string
	Algorithm          string
	SessionPersistence bool
	Addresses          []string
	RulesCount         int
	Healthmonitor      Healthmonitor
	CreatedIn          string
	UpdatedIn          string
}

// Healthmonitor is the load balancer's health-check configuration. HttpMethod,
// UrlPath and ExpectedCodes apply only when Type is HTTP.
type Healthmonitor struct {
	Type          string
	Delay         int
	Timeout       int
	MaxRetries    int
	HttpMethod    string
	UrlPath       string
	ExpectedCodes string
}

// Rule is a load balancer listener rule (external port → internal port on a server).
type Rule struct {
	ID                   string
	Loadbalancer         string
	Address              string
	Server               string
	Status               string
	ExternalProtocolPort int
	InternalProtocolPort int
}

func loadBalancerFromSchema(r *gen.LBDetailResponseSchema) LoadBalancer {
	lb := LoadBalancer{
		ID:                 r.Id,
		Name:               r.Name,
		Project:            r.Project,
		Status:             r.Status,
		SwitchStatus:       r.SwitchStatus,
		Algorithm:          r.Algorithm,
		SessionPersistence: r.SessionPersistence,
		RulesCount:         r.RulesCount,
		Healthmonitor:      healthmonitorFromSchema(r.Healthmonitor),
		CreatedIn:          r.CreatedIn.Format(time.RFC3339),
		UpdatedIn:          r.UpdatedIn.Format(time.RFC3339),
	}
	if r.Addresses != nil {
		lb.Addresses = append([]string(nil), r.Addresses...)
	}
	return lb
}

func healthmonitorFromSchema(h gen.HealthmonitorDetailResponseSchema) Healthmonitor {
	hm := Healthmonitor{
		Type:       h.Type,
		Delay:      h.Delay,
		Timeout:    h.Timeout,
		MaxRetries: h.MaxRetries,
	}
	if h.HttpMethod != nil {
		hm.HttpMethod = *h.HttpMethod
	}
	if h.UrlPath != nil {
		hm.UrlPath = *h.UrlPath
	}
	if h.ExpectedCodes != nil {
		hm.ExpectedCodes = *h.ExpectedCodes
	}
	return hm
}

func ruleFromSchema(r *gen.RuleDetailResponseSchema) Rule {
	return Rule{
		ID:                   r.Id,
		Loadbalancer:         r.Loadbalancer,
		Address:              r.Address,
		Server:               r.Server,
		Status:               r.Status,
		ExternalProtocolPort: r.ExternalProtocolPort,
		InternalProtocolPort: r.InternalProtocolPort,
	}
}

// HealthmonitorParams holds the health-check inputs for create/update.
type HealthmonitorParams struct {
	Type          string
	Delay         int
	Timeout       int
	MaxRetries    int
	HttpMethod    string // optional; HTTP only
	UrlPath       string // optional; HTTP only
	ExpectedCodes string // optional; HTTP only
}

// LoadBalancerCreateParams holds the inputs for creating a load balancer.
type LoadBalancerCreateParams struct {
	Name               string
	Algorithm          string // optional; empty → API default
	SessionPersistence *bool  // optional
	AddressID          string // optional; empty → address auto-allocated
	AddressDdos        *bool  // optional
	Healthmonitor      HealthmonitorParams
}

// RuleCreateParams holds the inputs for creating a listener rule.
type RuleCreateParams struct {
	AddressID            string
	ExternalProtocolPort int
	InternalProtocolPort int
}

// loadBalancerCreateBody builds the create request body, sending optional fields
// (algorithm, address) only when set. Healthmonitor is always sent as the API requires it.
func loadBalancerCreateBody(p LoadBalancerCreateParams) gen.LoadBalancerCreateJSONRequestBody {
	body := gen.LoadBalancerCreateJSONRequestBody{
		Name:               p.Name,
		SessionPersistence: p.SessionPersistence,
	}

	body.Healthmonitor.Type = gen.LoadBalancerCreateJSONBodyHealthmonitorType(p.Healthmonitor.Type)
	body.Healthmonitor.Delay = p.Healthmonitor.Delay
	body.Healthmonitor.Timeout = p.Healthmonitor.Timeout
	body.Healthmonitor.MaxRetries = p.Healthmonitor.MaxRetries
	if p.Healthmonitor.HttpMethod != "" {
		m := gen.LoadBalancerCreateJSONBodyHealthmonitorHttpMethod(p.Healthmonitor.HttpMethod)
		body.Healthmonitor.HttpMethod = &m
	}
	if v := p.Healthmonitor.UrlPath; v != "" {
		body.Healthmonitor.UrlPath = &v
	}
	if v := p.Healthmonitor.ExpectedCodes; v != "" {
		body.Healthmonitor.ExpectedCodes = &v
	}

	if p.Algorithm != "" {
		a := gen.LoadBalancerCreateJSONBodyAlgorithm(p.Algorithm)
		body.Algorithm = &a
	}
	if p.AddressID != "" || p.AddressDdos != nil {
		body.Address = &struct {
			DdosProtection *bool   `json:"ddos_protection,omitempty"`
			Id             *string `json:"id,omitempty"`
		}{DdosProtection: p.AddressDdos}
		if v := p.AddressID; v != "" {
			body.Address.Id = &v
		}
	}
	return body
}

// CreateLoadBalancer creates a load balancer in the project and returns its ID.
func (c *Client) CreateLoadBalancer(ctx context.Context, projectID string, p LoadBalancerCreateParams) (string, error) {
	resp, err := c.gen.LoadBalancerCreateWithResponse(ctx, projectID, loadBalancerCreateBody(p))
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return "", errors.New("cloapi: empty loadbalancer create response")
	}
	return resp.OK.Result.Id, nil
}

// GetLoadBalancer returns the load balancer's current detail.
func (c *Client) GetLoadBalancer(ctx context.Context, id string) (*LoadBalancer, error) {
	resp, err := c.gen.LoadBalancerDetailWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, errors.New("cloapi: empty loadbalancer detail response")
	}
	lb := loadBalancerFromSchema(resp.OK.Result)
	return &lb, nil
}

// ListLoadBalancers returns the project's load balancers (single page, matching the other list adapters).
func (c *Client) ListLoadBalancers(ctx context.Context, projectID string) ([]LoadBalancer, error) {
	resp, err := c.gen.LoadBalancerListWithResponse(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, nil
	}
	items := *resp.OK.Result
	out := make([]LoadBalancer, 0, len(items))
	for i := range items {
		out = append(out, loadBalancerFromSchema(&items[i]))
	}
	return out, nil
}

// RenameLoadBalancer changes the load balancer's name.
func (c *Client) RenameLoadBalancer(ctx context.Context, id, name string) error {
	_, err := c.gen.LoadBalancerRenameWithResponse(ctx, id, gen.LoadBalancerRenameJSONRequestBody{Name: name})
	return err
}

// UpdateLoadBalancer updates the balancing algorithm and/or session persistence.
// An empty algorithm is left unchanged; a nil sessionPersistence is left unchanged.
func (c *Client) UpdateLoadBalancer(ctx context.Context, id, algorithm string, sessionPersistence *bool) error {
	body := gen.LoadBalancerUpdateJSONRequestBody{SessionPersistence: sessionPersistence}
	if algorithm != "" {
		a := gen.LoadBalancerUpdateJSONBodyAlgorithm(algorithm)
		body.Algorithm = &a
	}
	_, err := c.gen.LoadBalancerUpdateWithResponse(ctx, id, body)
	return err
}

// healthmonitorUpdateBody builds the health-monitor update body, sending the
// HTTP-only fields (http_method, url_path, expected_codes) only when set.
func healthmonitorUpdateBody(p HealthmonitorParams) gen.LoadBalancerUpdateHealthmonitorJSONRequestBody {
	body := gen.LoadBalancerUpdateHealthmonitorJSONRequestBody{
		Type:       gen.LoadBalancerUpdateHealthmonitorJSONBodyType(p.Type),
		Delay:      p.Delay,
		Timeout:    p.Timeout,
		MaxRetries: p.MaxRetries,
	}
	if p.HttpMethod != "" {
		m := gen.LoadBalancerUpdateHealthmonitorJSONBodyHttpMethod(p.HttpMethod)
		body.HttpMethod = &m
	}
	if v := p.UrlPath; v != "" {
		body.UrlPath = &v
	}
	if v := p.ExpectedCodes; v != "" {
		body.ExpectedCodes = &v
	}
	return body
}

// UpdateHealthmonitor replaces the load balancer's health-check configuration.
func (c *Client) UpdateHealthmonitor(ctx context.Context, id string, p HealthmonitorParams) error {
	_, err := c.gen.LoadBalancerUpdateHealthmonitorWithResponse(ctx, id, healthmonitorUpdateBody(p))
	return err
}

// EnableLoadBalancer powers the load balancer on.
func (c *Client) EnableLoadBalancer(ctx context.Context, id string) error {
	_, err := c.gen.LoadBalancerEnableWithResponse(ctx, id)
	return err
}

// StopLoadBalancer powers the load balancer off.
func (c *Client) StopLoadBalancer(ctx context.Context, id string) error {
	_, err := c.gen.LoadBalancerStopWithResponse(ctx, id)
	return err
}

// DeleteLoadBalancer deletes a load balancer.
func (c *Client) DeleteLoadBalancer(ctx context.Context, id string) error {
	_, err := c.gen.LoadBalancerDeleteWithResponse(ctx, id)
	return err
}

// CreateRule creates a listener rule on the load balancer and returns its ID.
func (c *Client) CreateRule(ctx context.Context, loadBalancerID string, p RuleCreateParams) (string, error) {
	resp, err := c.gen.RuleCreateWithResponse(ctx, loadBalancerID, gen.RuleCreateJSONRequestBody{
		AddressId:            p.AddressID,
		ExternalProtocolPort: p.ExternalProtocolPort,
		InternalProtocolPort: p.InternalProtocolPort,
	})
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return "", errors.New("cloapi: empty rule create response")
	}
	return resp.OK.Result.Id, nil
}

// GetRule returns the listener rule's current detail.
func (c *Client) GetRule(ctx context.Context, id string) (*Rule, error) {
	resp, err := c.gen.RuleDetailWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, errors.New("cloapi: empty rule detail response")
	}
	r := ruleFromSchema(resp.OK.Result)
	return &r, nil
}

// ListRules returns the load balancer's listener rules (single page, matching the other list adapters).
func (c *Client) ListRules(ctx context.Context, loadBalancerID string) ([]Rule, error) {
	resp, err := c.gen.RuleListWithResponse(ctx, loadBalancerID)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, nil
	}
	items := *resp.OK.Result
	out := make([]Rule, 0, len(items))
	for i := range items {
		out = append(out, ruleFromSchema(&items[i]))
	}
	return out, nil
}

// DeleteRule deletes a listener rule.
func (c *Client) DeleteRule(ctx context.Context, id string) error {
	_, err := c.gen.RuleDeleteWithResponse(ctx, id)
	return err
}
