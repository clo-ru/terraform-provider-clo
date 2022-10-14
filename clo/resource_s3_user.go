package clo

import (
	"context"
	"fmt"
	clo_lib "github.com/clo-ru/cloapi-go-client/clo"
	clo_storage "github.com/clo-ru/cloapi-go-client/services/storage"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

const (
	s3UserActive   = "AVAILABLE"
	s3UserDelete   = "DELETE"
	s3UserCreating = "CREATING"
	s3UserDeleting = "DELETING"
)

func resourceS3User() *schema.Resource {
	return &schema.Resource{
		Description:   "Create a new user of the object storage",
		ReadContext:   resourceS3UserRead,
		CreateContext: resourceS3UserCreate,
		UpdateContext: resourceS3UserUpdate,
		DeleteContext: resourceS3UserDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"project_id": {
				Description: "ID of the project where the user should be created",
				Type:        schema.TypeString,
				Required:    true,
			},
			"canonical_name": {
				Description: "Canonical name of the user. The storage uses this name. " +
					"Should be unique in scope of the tenant",
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"default_bucket": {
				Description: "Should the default bucket be created with the user",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
			},
			"max_buckets": {
				Description: "How many buckets the user could create",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"name": {
				Description: "Human-readable name of the user",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"user_quota_max_size": {
				Description: "Total size of the objects the user can store",
				Type:        schema.TypeInt,
				Required:    true,
			},
			"user_quota_max_objects": {
				Description: "How many objects the user can create",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"bucket_quota_max_size": {
				Description: "A maximum size of a bucket",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
			},
			"bucket_quota_max_objects": {
				Description: "How many objects can be created within a bucket",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
			},
			"user_id": {
				Description: "ID of the created user",
				Type:        schema.TypeString, Computed: true},
			"status": {Type: schema.TypeString, Computed: true},
			"tenant": {
				Description: "Name of the user's tenant. " +
					"Name of the user's project uses by default",
				Type: schema.TypeString, Computed: true},
		},
	}
}

func resourceS3UserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	req := buildRequest(d)
	resp, e := req.Make(ctx, cli)
	if resp.Code == 404 {
		e = fmt.Errorf("NotFound returned")
	}
	if e != nil {
		return diag.FromErr(e)
	}
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_storage.S3UserDetailRequest{UserID: resp.Result.ID}
			resp, e := req.Make(ctx, cli)
			if e != nil {
				return resp, "", e
			} else {
				return resp, resp.Result.Status, nil
			}
		},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Target:     []string{s3UserActive},
		Pending:    []string{s3UserCreating},
	}
	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Result.ID)
	return resourceS3UserRead(ctx, d, m)
}

func buildRequest(d *schema.ResourceData) clo_storage.S3UserCreateRequest {
	cn := d.Get("canonical_name").(string)
	b := clo_storage.S3UserCreateBody{
		Name:          cn,
		CanonicalName: cn,
		MaxBuckets:    d.Get("max_buckets").(int),
		BucketQuota:   clo_storage.CreateQuotaParams{},
		UserQuota: clo_storage.CreateQuotaParams{
			MaxSize: d.Get("user_quota_max_size").(int),
		},
	}
	if v, ok := d.GetOk("name"); ok {
		b.Name = v.(string)
	}
	if v, ok := d.GetOk("bucket_quota_max_size"); ok {
		b.UserQuota.MaxObjects = v.(int)
	}
	if v, ok := d.GetOk("user_quota_max_objects"); ok {
		b.UserQuota.MaxObjects = v.(int)
	}
	if v, ok := d.GetOk("bucket_quota_max_objects"); ok {
		b.UserQuota.MaxObjects = v.(int)
	}
	req := clo_storage.S3UserCreateRequest{
		ProjectID: d.Get("project_id").(string),
		Body:      b,
	}
	return req
}

func resourceS3UserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	uId := d.Id()
	cli := m.(*clo_lib.ApiClient)
	if d.HasChange("name") {
		_, n := d.GetChange("name")
		req := clo_storage.S3UserPatchRequest{
			UserID: uId,
			Body: clo_storage.S3UserPatchBody{
				Name: n.(string),
			},
		}
		_, e := req.Make(ctx, cli)
		if e != nil {
			return diag.FromErr(e)
		}
	}
	if d.HasChanges(
		"max_buckets",
		"user_quota.max_size",
		"user_quota.max_objects",
		"bucket_quota_max_size",
		"bucket_quota_max_objects",
	) {
		b := clo_storage.S3UserQuotaPatchBody{}
		if d.HasChange("max_buckets") {
			_, mb := d.GetChange("max_buckets")
			b.MaxBuckets = mb.(int)
		}
		if d.HasChange("user_quota_max_size") {
			_, ms := d.GetChange("user_quota.max_size")
			b.UserQuota.MaxSize = ms.(int)
		}
		if d.HasChange("user_quota_max_objects") {
			_, ms := d.GetChange("user_quota.max_objects")
			b.UserQuota.MaxObjects = ms.(int)
		}
		if d.HasChange("bucket_quota_max_size") {
			_, ms := d.GetChange("bucket_quota.max_size")
			b.BucketQuota.MaxSize = ms.(int)
		}
		if d.HasChange("bucket_quota_max_objects") {
			_, ms := d.GetChange("bucket_quota.max_objects")
			b.BucketQuota.MaxObjects = ms.(int)
		}
		req := clo_storage.S3UserQuotaPatchRequest{
			UserID: uId,
			Body:   b,
		}
		_, e := req.Make(ctx, cli)
		if e != nil {
			return diag.FromErr(e)
		}
	}
	return resourceS3UserRead(ctx, d, m)
}

func resourceS3UserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	uId := d.Id()
	cli := m.(*clo_lib.ApiClient)
	req := clo_storage.S3UserDetailRequest{UserID: uId}
	resp, e := req.Make(ctx, cli)
	if e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("user_id", resp.Result.Status); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("status", resp.Result.Status); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("tenant", resp.Result.Tenant); e != nil {
		return diag.FromErr(e)
	}
	if e := d.Set("max_buckets", resp.Result.MaxBuckets); e != nil {
		return diag.FromErr(e)
	}
	return dispatchQuotaInfo(d, resp.Result.Quotas)
}

func dispatchQuotaInfo(d *schema.ResourceData, q []clo_storage.QuotaInfo) diag.Diagnostics {
	for _, qi := range q {
		switch qi.Type {
		case "user":
			if e := d.Set("user_quota_max_size", qi.MaxSize); e != nil {
				return diag.FromErr(e)
			}
			if e := d.Set("user_quota_max_objects", qi.MaxObjects); e != nil {
				return diag.FromErr(e)
			}
		case "bucket":
			if e := d.Set("bucket_quota_max_size", qi.MaxSize); e != nil {
				return diag.FromErr(e)
			}
			if e := d.Set("bucket_quota_max_objects", qi.MaxObjects); e != nil {
				return diag.FromErr(e)
			}
		default:
			return diag.FromErr(fmt.Errorf("wrong the quota_info type, %s", qi.Type))
		}
	}
	return nil
}

func resourceS3UserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	cli := m.(*clo_lib.ApiClient)
	req := clo_storage.S3UserDeleteRequest{UserID: d.Id()}
	if e := req.Make(ctx, cli); e != nil {
		return diag.FromErr(e)
	}
	createStateConf := resource.StateChangeConf{
		Refresh: func() (result interface{}, state string, err error) {
			req := clo_storage.S3UserDetailRequest{UserID: d.Id()}
			resp, e := req.Make(ctx, cli)
			if e != nil {
				return resp, "", e
			}
			if resp.Code == 404 {
				return resp.Result, s3UserDelete, nil
			}
			return resp.Result, resp.Result.Status, nil
		},
		Target:     []string{s3UserDelete},
		Pending:    []string{s3UserDeleting},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
	}
	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
