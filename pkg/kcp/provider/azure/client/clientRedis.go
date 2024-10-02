package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"k8s.io/utils/ptr"
)

type RedisClient interface {
	CreateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armredis.CreateParameters) error
	UpdateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armredis.UpdateParameters) error
	GetRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) (*armredis.ResourceInfo, error)
	DeleteRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) error
	GetRedisInstanceAccessKeys(ctx context.Context, resourceGroupName, redisInstanceName string) ([]string, error)
	CreatePrivateEndPoint(ctx context.Context, resourceGroupName, privateEndPointName string, parameters armnetwork.PrivateEndpoint) error
	GetPrivateEndPoint(ctx context.Context, resourceGroupName, privateEndPointName string) (*armnetwork.PrivateEndpoint, error)
	DeletePrivateEndPoint(ctx context.Context, resourceGroupName, privateEndPointName string) error
	CreatePrivateDnsZoneGroup(ctx context.Context, resourceGroupName, privateEndPointName, privateDnsZoneGroupName string, parameters armnetwork.PrivateDNSZoneGroup) error
	DeletePrivateDnsZoneGroup(ctx context.Context, resourceGroupName, privateEndPointName, privateDnsZoneGroupName string) error
	GetPrivateDnsZoneGroup(ctx context.Context, resourceGroupName, privateEndPointName, privateDnsZoneGroupName string) (*armnetwork.PrivateDNSZoneGroup, error)
	CreatePrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneName string, parameters armprivatedns.PrivateZone) error
	GetPrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneGroupName string) (*armprivatedns.PrivateZone, error)
	CreateVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName string, parameters armprivatedns.VirtualNetworkLink) error
	GetVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName string) (*armprivatedns.VirtualNetworkLink, error)
}

func NewRedisClient(svc *armredis.Client, pepClient *armnetwork.PrivateEndpointsClient, dnsZoneGroupClient *armnetwork.PrivateDNSZoneGroupsClient,
	dnsZoneClient *armprivatedns.PrivateZonesClient, virtualNetworkLinkClient *armprivatedns.VirtualNetworkLinksClient) RedisClient {
	return &redisClient{
		svc:                      svc,
		pepClient:                pepClient,
		dnsZoneGroupClient:       dnsZoneGroupClient,
		dnsZoneClient:            dnsZoneClient,
		virtualNetworkLinkClient: virtualNetworkLinkClient,
	}
}

var _ RedisClient = &redisClient{}

type redisClient struct {
	svc                      *armredis.Client
	pepClient                *armnetwork.PrivateEndpointsClient
	dnsZoneGroupClient       *armnetwork.PrivateDNSZoneGroupsClient
	dnsZoneClient            *armprivatedns.PrivateZonesClient
	virtualNetworkLinkClient *armprivatedns.VirtualNetworkLinksClient
}

func (c *redisClient) CreateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armredis.CreateParameters) error {
	_, err := c.svc.BeginCreate(
		ctx,
		resourceGroupName,
		redisInstanceName,
		parameters,
		nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *redisClient) GetRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) (*armredis.ResourceInfo, error) {
	clientGetResponse, err := c.svc.Get(ctx, resourceGroupName, redisInstanceName, nil)
	if err != nil {
		return nil, err
	}
	return &clientGetResponse.ResourceInfo, nil
}

func (c *redisClient) DeleteRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) error {
	_, err := c.svc.BeginDelete(ctx, resourceGroupName, redisInstanceName, nil)
	if err != nil {
		return err
	}
	return nil
}

func (c *redisClient) GetRedisInstanceAccessKeys(ctx context.Context, resourceGroupName, redisInstanceName string) ([]string, error) {
	redisAccessKeys, err := c.svc.ListKeys(ctx, resourceGroupName, redisInstanceName, nil)

	if err != nil {
		return nil, err
	}
	return []string{ptr.Deref(redisAccessKeys.PrimaryKey, "")}, nil
}

func (c *redisClient) UpdateRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string, parameters armredis.UpdateParameters) error {
	_, err := c.svc.Update(
		ctx,
		resourceGroupName,
		redisInstanceName,
		parameters,
		nil)

	if err != nil {
		return err
	}

	return nil
}

func (c *redisClient) CreatePrivateEndPoint(ctx context.Context, resourceGroupName, privateEndPointName string, parameters armnetwork.PrivateEndpoint) error {
	_, err := c.pepClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		privateEndPointName,
		parameters,
		nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *redisClient) GetPrivateEndPoint(ctx context.Context, resourceGroupName, privateEndPointName string) (*armnetwork.PrivateEndpoint, error) {
	privateEndpointsClientGetResponse, err := c.pepClient.Get(
		ctx,
		resourceGroupName,
		privateEndPointName,
		nil)
	if err != nil {
		return nil, err
	}

	return &privateEndpointsClientGetResponse.PrivateEndpoint, nil
}

func (c *redisClient) DeletePrivateEndPoint(ctx context.Context, resourceGroupName, privateEndPointName string) error {
	_, err := c.pepClient.BeginDelete(
		ctx,
		resourceGroupName,
		privateEndPointName,
		nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *redisClient) CreatePrivateDnsZoneGroup(ctx context.Context, resourceGroupName, privateEndPointName, privateDnsZoneGroupName string, parameters armnetwork.PrivateDNSZoneGroup) error {
	_, err := c.dnsZoneGroupClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		privateEndPointName,
		privateDnsZoneGroupName,
		parameters,
		nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *redisClient) GetPrivateDnsZoneGroup(ctx context.Context, resourceGroupName, privateEndPointName, privateDnsZoneGroupName string) (*armnetwork.PrivateDNSZoneGroup, error) {
	privateDnsZoneGroupClientGetResponse, err := c.dnsZoneGroupClient.Get(
		ctx,
		resourceGroupName,
		privateEndPointName,
		privateDnsZoneGroupName,
		nil)
	if err != nil {
		return nil, err
	}

	return &privateDnsZoneGroupClientGetResponse.PrivateDNSZoneGroup, nil
}

func (c *redisClient) DeletePrivateDnsZoneGroup(ctx context.Context, resourceGroupName, privateEndPointName, privateDnsZoneGroupName string) error {
	_, err := c.dnsZoneGroupClient.BeginDelete(
		ctx,
		resourceGroupName,
		privateEndPointName,
		privateDnsZoneGroupName,
		nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *redisClient) CreatePrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneName string, parameters armprivatedns.PrivateZone) error {
	_, err := c.dnsZoneClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		privateDnsZoneName,
		parameters,
		nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *redisClient) GetPrivateDnsZone(ctx context.Context, resourceGroupName, privateDnsZoneGroupName string) (*armprivatedns.PrivateZone, error) {
	privateDnsZoneClientGetResponse, err := c.dnsZoneClient.Get(
		ctx,
		resourceGroupName,
		privateDnsZoneGroupName,
		nil)
	if err != nil {
		return nil, err
	}

	return &privateDnsZoneClientGetResponse.PrivateZone, nil
}

func (c *redisClient) CreateVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName string, parameters armprivatedns.VirtualNetworkLink) error {
	_, err := c.virtualNetworkLinkClient.BeginCreateOrUpdate(
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

func (c *redisClient) GetVirtualNetworkLink(ctx context.Context, resourceGroupName, privateZoneName, virtualNetworkLinkName string) (*armprivatedns.VirtualNetworkLink, error) {
	virtualNetworkLinkClientGetResponse, err := c.virtualNetworkLinkClient.Get(
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
