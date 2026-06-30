package cloapi

import (
	"context"
	"errors"
	"time"

	gen "github.com/clo-ru/cloapi-go-client/v3"
)

// Server is the provider-facing view of a compute instance.
type Server struct {
	ID          string
	Name        string
	Status      string
	Project     string
	CreatedIn   string
	RescueMode  string
	GuestAgent  bool
	FlavorRam   int
	FlavorVcpus int
	ImageName   string // "<distribution> <version>", empty if no image
	RecipeName  string
	Addresses   []string
	Disks       []ServerDisk
}

// ServerDisk is one disk attached to a server.
type ServerDisk struct {
	ID          string
	StorageType string
}

func serverFromSchema(r *gen.ServerSchema) Server {
	s := Server{
		ID:         r.Id,
		Name:       r.Name,
		Status:     r.Status,
		Project:    r.Project,
		RescueMode: r.RescueMode,
		GuestAgent: r.GuestAgent,
		CreatedIn:  r.CreatedIn.Format(time.RFC3339),
	}
	if r.Flavor != nil {
		s.FlavorRam = r.Flavor.Ram
		s.FlavorVcpus = r.Flavor.Vcpus
	}
	if r.Image != nil {
		s.ImageName = r.Image.OperationSystem.Distribution + " " + r.Image.OperationSystem.Version
	}
	if r.Recipe != nil {
		s.RecipeName = r.Recipe.Name
	}
	if r.Addresses != nil {
		s.Addresses = *r.Addresses
	}
	if r.DiskData != nil {
		for _, d := range *r.DiskData {
			s.Disks = append(s.Disks, ServerDisk{ID: d.Id, StorageType: d.StorageType})
		}
	}
	return s
}

// ServerStorage describes a disk to attach to a server on create.
type ServerStorage struct {
	Bootable    bool
	StorageType string
	Size        int
}

// ServerAddress describes an address to attach to a server on create.
type ServerAddress struct {
	External       bool
	Version        int
	AddressID      string
	DdosProtection bool
	Bandwidth      int
}

// ServerLicense describes a license to apply to a server on create.
type ServerLicense struct {
	Addon string
	Name  string
}

// ServerCreateParams describes a new compute instance.
type ServerCreateParams struct {
	ProjectID   string
	Name        string
	ImageID     string
	FlavorRam   int
	FlavorVcpus int
	RecipeID    string
	Storages    []ServerStorage
	Addresses   []ServerAddress
	Keypairs    []string
	Licenses    []ServerLicense
}

// CreateServer creates an instance and returns its ID. All the awkward inline-struct
// construction the generated body requires is contained here.
func (c *Client) CreateServer(ctx context.Context, p ServerCreateParams) (string, error) {
	body := gen.ServerCreateJSONRequestBody{Name: p.Name}
	body.Flavor.Ram = p.FlavorRam
	body.Flavor.Vcpus = p.FlavorVcpus
	if p.ImageID != "" {
		body.Image = &p.ImageID
	}
	if p.RecipeID != "" {
		body.Recipe = &p.RecipeID
	}
	if len(p.Keypairs) > 0 {
		body.Keypairs = &p.Keypairs
	}

	if len(p.Storages) > 0 {
		storages := make([]struct {
			Bootable    *bool   `json:"bootable,omitempty"`
			Size        int     `json:"size"`
			StorageType *string `json:"storage_type,omitempty"`
		}, len(p.Storages))
		for i, s := range p.Storages {
			bootable, st := s.Bootable, s.StorageType
			storages[i].Bootable = &bootable
			storages[i].Size = s.Size
			storages[i].StorageType = &st
		}
		body.Storages = &storages
	}

	if len(p.Addresses) > 0 {
		addrs := make([]struct {
			AddressId        *string                                            `json:"address_id,omitempty"`
			BandwidthMaxMbps *gen.ServerCreateJSONBodyAddressesBandwidthMaxMbps `json:"bandwidth_max_mbps,omitempty"`
			DdosProtection   *bool                                              `json:"ddos_protection,omitempty"`
			External         *bool                                              `json:"external,omitempty"`
			Version          *int                                               `json:"version,omitempty"`
		}, len(p.Addresses))
		for i, a := range p.Addresses {
			external, ddos, version := a.External, a.DdosProtection, a.Version
			addrs[i].External = &external
			addrs[i].DdosProtection = &ddos
			addrs[i].Version = &version
			if a.AddressID != "" {
				id := a.AddressID
				addrs[i].AddressId = &id
			}
			if a.Bandwidth != 0 {
				bw := gen.ServerCreateJSONBodyAddressesBandwidthMaxMbps(a.Bandwidth)
				addrs[i].BandwidthMaxMbps = &bw
			}
		}
		body.Addresses = &addrs
	}

	if len(p.Licenses) > 0 {
		lic := make([]struct {
			Addon string `json:"addon"`
			Name  string `json:"name"`
		}, len(p.Licenses))
		for i, l := range p.Licenses {
			lic[i].Addon = l.Addon
			lic[i].Name = l.Name
		}
		body.Licenses = &lic
	}

	resp, err := c.gen.ServerCreateWithResponse(ctx, p.ProjectID, body)
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return "", errors.New("cloapi: empty server create response")
	}
	return resp.OK.Result.Id, nil
}

// GetServer returns the instance's current detail.
func (c *Client) GetServer(ctx context.Context, id string) (*Server, error) {
	resp, err := c.gen.ServerDetailWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, errors.New("cloapi: empty server detail response")
	}
	s := serverFromSchema(resp.OK.Result)
	return &s, nil
}

// ListServers returns the project's instances (single page, matching v2 behavior).
func (c *Client) ListServers(ctx context.Context, projectID string) ([]Server, error) {
	resp, err := c.gen.ProjectServerListWithResponse(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return nil, nil
	}
	items := *resp.OK.Result
	out := make([]Server, 0, len(items))
	for i := range items {
		out = append(out, serverFromSchema(&items[i]))
	}
	return out, nil
}

// ResizeServer changes the instance's flavor.
func (c *Client) ResizeServer(ctx context.Context, id string, ram, vcpus int) error {
	_, err := c.gen.ServerResizeWithResponse(ctx, id, gen.ServerResizeJSONRequestBody{Ram: ram, Vcpus: vcpus})
	return err
}

// ChangeServerPassword sets the instance's password.
func (c *Client) ChangeServerPassword(ctx context.Context, id, password string) error {
	_, err := c.gen.ServerChangePasswordWithResponse(ctx, id, gen.ServerChangePasswordJSONRequestBody{Password: password})
	return err
}

// DeleteServer deletes the instance, optionally deleting the given addresses and volumes.
func (c *Client) DeleteServer(ctx context.Context, id string, deleteAddresses, deleteVolumes []string) error {
	body := gen.ServerDeleteJSONRequestBody{}
	if len(deleteAddresses) > 0 {
		body.DeleteAddresses = &gen.ServerDeleteJSONBody_DeleteAddresses{}
		if err := body.DeleteAddresses.FromServerDeleteJSONBodyDeleteAddresses0(deleteAddresses); err != nil {
			return err
		}
	}
	if len(deleteVolumes) > 0 {
		body.DeleteVolumes = &gen.ServerDeleteJSONBody_DeleteVolumes{}
		if err := body.DeleteVolumes.FromServerDeleteJSONBodyDeleteVolumes0(deleteVolumes); err != nil {
			return err
		}
	}
	_, err := c.gen.ServerDeleteWithResponse(ctx, id, body)
	return err
}
