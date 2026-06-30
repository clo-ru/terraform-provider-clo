package cloapi

import (
	"context"
	"fmt"

	gen "github.com/clo-ru/cloapi-go-client/v3"
)

// S3User is the provider-facing view of an object-storage user.
type S3User struct {
	ID            string
	Name          string
	CanonicalName string
	Status        string
	Tenant        string
	MaxBuckets    int
	Quotas        []S3Quota
}

func s3UserFromSchema(r *gen.S3UserSchema) S3User {
	u := S3User{
		ID:            r.Id,
		Name:          r.Name,
		CanonicalName: r.CanonicalName,
		Status:        r.Status,
	}
	if r.Tenant != nil {
		u.Tenant = *r.Tenant
	}
	if r.MaxBuckets != nil {
		u.MaxBuckets = *r.MaxBuckets
	}
	for _, q := range r.Quotas {
		sq := S3Quota{Type: q.Type}
		if q.MaxSize != nil {
			sq.MaxSize = *q.MaxSize
		}
		if q.MaxObjects != nil {
			sq.MaxObjects = *q.MaxObjects
		}
		u.Quotas = append(u.Quotas, sq)
	}
	return u
}

// S3Quota is one quota entry ("user" or "bucket").
type S3Quota struct {
	Type       string
	MaxSize    int
	MaxObjects int
}

// S3UserCreateParams describes a new object-storage user.
type S3UserCreateParams struct {
	ProjectID             string
	Name                  string
	CanonicalName         string
	DefaultBucket         bool
	MaxBuckets            int
	UserQuotaMaxSize      int
	UserQuotaMaxObjects   int
	BucketQuotaMaxSize    int
	BucketQuotaMaxObjects int
}

// S3UserQuotaParams describes a quota update.
type S3UserQuotaParams struct {
	MaxBuckets            int
	UserQuotaMaxSize      int
	UserQuotaMaxObjects   int
	BucketQuotaMaxSize    int
	BucketQuotaMaxObjects int
}

// S3Keys are an access/secret key pair. SecretKey is only returned on generation.
type S3Keys struct {
	AccessKey string
	SecretKey string
}

func intPtr(v int) *int { return &v }

// CreateS3User creates an object-storage user and returns its ID.
func (c *Client) CreateS3User(ctx context.Context, p S3UserCreateParams) (string, error) {
	body := gen.S3UserCreateJSONRequestBody{
		CanonicalName: p.CanonicalName,
		MaxBuckets:    p.MaxBuckets,
	}
	name := p.CanonicalName
	if p.Name != "" {
		name = p.Name
	}
	body.Name = &name
	if p.DefaultBucket {
		db := true
		body.DefaultBucket = &db
	}
	body.UserQuota.MaxSize = p.UserQuotaMaxSize
	if p.UserQuotaMaxObjects > 0 {
		body.UserQuota.MaxObjects = intPtr(p.UserQuotaMaxObjects)
	}
	if p.BucketQuotaMaxSize > 0 || p.BucketQuotaMaxObjects > 0 {
		body.BucketQuota = &struct {
			MaxObjects *int `json:"max_objects,omitempty"`
			MaxSize    *int `json:"max_size,omitempty"`
		}{}
		if p.BucketQuotaMaxSize > 0 {
			body.BucketQuota.MaxSize = intPtr(p.BucketQuotaMaxSize)
		}
		if p.BucketQuotaMaxObjects > 0 {
			body.BucketQuota.MaxObjects = intPtr(p.BucketQuotaMaxObjects)
		}
	}

	resp, err := c.gen.S3UserCreateWithResponse(ctx, p.ProjectID, body)
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return "", fmt.Errorf("cloapi: empty s3 user create response")
	}
	return resp.OK.Result.Id, nil
}

// GetS3User returns the user's current detail.
func (c *Client) GetS3User(ctx context.Context, id string) (*S3User, error) {
	resp, err := c.gen.S3UserDetailsWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, fmt.Errorf("cloapi: empty s3 user detail response")
	}
	u := s3UserFromSchema(resp.OK.Result)
	return &u, nil
}

// ListS3Users returns the project's object-storage users (single page, matching v2).
func (c *Client) ListS3Users(ctx context.Context, projectID string) ([]S3User, error) {
	resp, err := c.gen.S3UsersListWithResponse(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, nil
	}
	items := *resp.OK.Result
	out := make([]S3User, 0, len(items))
	for i := range items {
		out = append(out, s3UserFromSchema(&items[i]))
	}
	return out, nil
}

// UpdateS3UserName updates the user's human-readable name.
func (c *Client) UpdateS3UserName(ctx context.Context, id, name string) error {
	_, err := c.gen.S3UserUpdateWithResponse(ctx, id, gen.S3UserUpdateJSONRequestBody{Name: name})
	return err
}

// UpdateS3UserQuota updates the user's quotas.
func (c *Client) UpdateS3UserQuota(ctx context.Context, id string, p S3UserQuotaParams) error {
	body := gen.S3UserUpdateQuotaJSONRequestBody{MaxBuckets: intPtr(p.MaxBuckets)}
	body.UserQuota = &struct {
		MaxObjects *int `json:"max_objects,omitempty"`
		MaxSize    int  `json:"max_size"`
	}{MaxSize: p.UserQuotaMaxSize}
	if p.UserQuotaMaxObjects > 0 {
		body.UserQuota.MaxObjects = intPtr(p.UserQuotaMaxObjects)
	}
	body.BucketQuota = &struct {
		MaxObjects *int `json:"max_objects,omitempty"`
		MaxSize    *int `json:"max_size,omitempty"`
	}{}
	if p.BucketQuotaMaxSize > 0 {
		body.BucketQuota.MaxSize = intPtr(p.BucketQuotaMaxSize)
	}
	if p.BucketQuotaMaxObjects > 0 {
		body.BucketQuota.MaxObjects = intPtr(p.BucketQuotaMaxObjects)
	}
	_, err := c.gen.S3UserUpdateQuotaWithResponse(ctx, id, body)
	return err
}

// DeleteS3User deletes the user.
func (c *Client) DeleteS3User(ctx context.Context, id string) error {
	_, err := c.gen.S3UserDeleteWithResponse(ctx, id)
	return err
}

// GenS3UserKeys generates a new access/secret key pair for the user.
func (c *Client) GenS3UserKeys(ctx context.Context, id string) (*S3Keys, error) {
	resp, err := c.gen.S3GenUserKeysWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, fmt.Errorf("cloapi: empty s3 keys generate response")
	}
	return &S3Keys{AccessKey: resp.OK.Result.AccessKey, SecretKey: resp.OK.Result.SecretKey}, nil
}

// GetS3UserAccessKey returns the user's current access key. The secret key is not
// returned by the API on read (only on generation).
func (c *Client) GetS3UserAccessKey(ctx context.Context, id string) (string, error) {
	resp, err := c.gen.S3GetUserKeysWithResponse(ctx, id)
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil || resp.OK.Result.AccessKey == nil {
		return "", nil
	}
	return *resp.OK.Result.AccessKey, nil
}
