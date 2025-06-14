package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"k8s.io/utils/ptr"
)

type VirtualNetworkLinkClient interface {
	CreateVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName, vnetId string) error
	GetVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName string) (*armprivatedns.VirtualNetworkLink, error)
	DeleteVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName string) error
}

func NewVirtualNetworkLinkClient(svc *armprivatedns.VirtualNetworkLinksClient) VirtualNetworkLinkClient {
	return &virtualNetworkLinkClient{svc: svc}
}

var _ VirtualNetworkLinkClient = &virtualNetworkLinkClient{}

type virtualNetworkLinkClient struct {
	svc *armprivatedns.VirtualNetworkLinksClient
}

func (c *virtualNetworkLinkClient) CreateVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName, vnetId string) error {

	parameters := armprivatedns.VirtualNetworkLink{
		Location: ptr.To("global"),
		Properties: &armprivatedns.VirtualNetworkLinkProperties{
			VirtualNetwork: &armprivatedns.SubResource{
				ID: ptr.To(vnetId),
			},
			RegistrationEnabled: ptr.To(false),
		},
	}
	_, err := c.svc.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		privateZoneName,
		virtualNetworkLinkName,
		parameters,
		nil)
	if err != nil {
		return err
	}
	return nil
}
func (c *virtualNetworkLinkClient) GetVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName string) (*armprivatedns.VirtualNetworkLink, error) {
	virtualNetworkLinkClientGetResponse, err := c.svc.Get(
		ctx,
		resourceGroupName,
		privateZoneName,
		virtualNetworkLinkName,
		nil)
	if err != nil {
		return nil, err
	}
	return &virtualNetworkLinkClientGetResponse.VirtualNetworkLink, nil
}

func (c *virtualNetworkLinkClient) DeleteVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName string) error {
	_, err := c.svc.BeginDelete(
		ctx,
		resourceGroupName,
		privateZoneName,
		virtualNetworkLinkName,
		nil)
	if err != nil {
		return err
	}
	return nil
}
