package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

type Client interface {
	azureclient.RedisClient
}

func NewClientProvider() azureclient.ClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string) (Client, error) {

		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, &azidentity.ClientSecretCredentialOptions{})

		if err != nil {
			return nil, err
		}

		armRedisClientInstance, err := armredis.NewClient(subscriptionId, cred, nil)
		if err != nil {
			return nil, err
		}

		privateEndpointClientInstance, err := armnetwork.NewPrivateEndpointsClient(subscriptionId, cred, nil)
		if err != nil {
			return nil, err
		}

		privateDnsZoneGroupClientInstance, err := armnetwork.NewPrivateDNSZoneGroupsClient(subscriptionId, cred, nil)
		if err != nil {
			return nil, err
		}

		privateDnsZoneClientInstance, err := armprivatedns.NewPrivateZonesClient(subscriptionId, cred, nil)
		if err != nil {
			return nil, err
		}

		virtualNetworkLinkClientInstance, err := armprivatedns.NewVirtualNetworkLinksClient(subscriptionId, cred, nil)
		if err != nil {
			return nil, err
		}

		return newClient(
			azureclient.NewRedisClient(
				armRedisClientInstance,
				privateEndpointClientInstance,
				privateDnsZoneGroupClientInstance,
				privateDnsZoneClientInstance,
				virtualNetworkLinkClientInstance)), nil
	}
}

type redisInstanceClient struct {
	azureclient.RedisClient
}

func newClient(redisClient azureclient.RedisClient) Client {
	return &redisInstanceClient{
		RedisClient: redisClient,
	}
}
