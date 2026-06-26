package cloapi

import (
	"context"
	"fmt"

	gen "github.com/clo-ru/cloapi-go-client/v3"
)

// ImportKeypair imports an existing public key and returns the keypair ID.
func (c *Client) ImportKeypair(ctx context.Context, projectID, name, publicKey string) (string, error) {
	resp, err := c.gen.ImportKeypairWithResponse(ctx, projectID, gen.ImportKeypairJSONRequestBody{
		Name:      name,
		PublicKey: publicKey,
	})
	if err != nil {
		return "", err
	}
	if resp.OK == nil || resp.OK.Result == nil {
		return "", fmt.Errorf("cloapi: empty keypair import response")
	}
	return resp.OK.Result.Id, nil
}

// DeleteKeypair deletes a keypair.
func (c *Client) DeleteKeypair(ctx context.Context, id string) error {
	_, err := c.gen.KeypairDeleteWithResponse(ctx, id)
	return err
}
