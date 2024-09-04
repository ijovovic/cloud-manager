package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	armResources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"k8s.io/utils/ptr"
)

type Client interface {
	GetSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string) (*armnetwork.SubnetsClientGetResponse, error)
	CreateSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName, location, cidr string) error
	DeleteSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string) error
	GetVpc(ctx context.Context, resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetworksClientGetResponse, error)
	GetResourceGroup(ctx context.Context, resourceGroupName string) (*armResources.ResourceGroupsClientGetResponse, error)
	CreateResourceGroup(ctx context.Context, resourceGroupName, location string) error
	DeleteResourceGroup(ctx context.Context, resourceGroupName string) error
}

func NewClientProvider() azureClient.SkrClientProvider[Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string) (Client, error) {

		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, &azidentity.ClientSecretCredentialOptions{})

		if err != nil {
			return nil, err
		}

		virtualNetworksClientInstance, err := armnetwork.NewVirtualNetworksClient(subscriptionId, cred, nil)

		if err != nil {
			return nil, err
		}

		subnetClientInstance, err := armnetwork.NewSubnetsClient(subscriptionId, cred, nil)

		if err != nil {
			return nil, err
		}

		resourceGroupClientInstance, err := armResources.NewResourceGroupsClient(subscriptionId, cred, nil)

		if err != nil {
			return nil, err
		}

		return newClient(virtualNetworksClientInstance, subnetClientInstance, resourceGroupClientInstance), nil
	}
}

type networkClient struct {
	VirtualNetworkClient *armnetwork.VirtualNetworksClient
	SubnetClient         *armnetwork.SubnetsClient
	ResourceGroupClient  *armResources.ResourceGroupsClient
}

func newClient(virtualNetworkClient *armnetwork.VirtualNetworksClient, subnetClient *armnetwork.SubnetsClient, resourceGroupClient *armResources.ResourceGroupsClient) Client {
	return &networkClient{
		VirtualNetworkClient: virtualNetworkClient,
		SubnetClient:         subnetClient,
		ResourceGroupClient:  resourceGroupClient,
	}
}

func (c *networkClient) GetSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string) (*armnetwork.SubnetsClientGetResponse, error) {
	logger := composed.LoggerFromCtx(ctx)

	subnetClientGetResponse, error := c.SubnetClient.Get(ctx, resourceGroupName, virtualNetworkName, subnetName, nil)

	if error != nil {
		logger.Error(error, "Failed to get Azure Redis subnet")
		return nil, error
	}

	return &subnetClientGetResponse, nil
}

func (c *networkClient) CreateSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName, location, cidr string) error {
	logger := composed.LoggerFromCtx(ctx)

	virtualNetwork := armnetwork.VirtualNetwork{
		Location: ptr.To(location),
		Properties: &armnetwork.VirtualNetworkPropertiesFormat{
			AddressSpace: &armnetwork.AddressSpace{
				AddressPrefixes: []*string{
					to.Ptr(cidr)},
			},
			Subnets: []*armnetwork.Subnet{
				{
					Name: to.Ptr(subnetName),
					Properties: &armnetwork.SubnetPropertiesFormat{
						AddressPrefix: to.Ptr(cidr),
					},
				}},
		},
	}
	_, error := c.VirtualNetworkClient.BeginCreateOrUpdate(ctx, resourceGroupName, virtualNetworkName, virtualNetwork, nil)

	if error != nil {
		logger.Error(error, "Failed to update Azure Redis virtual network")
		return error
	}

	return nil
}

func (c *networkClient) GetVpc(ctx context.Context, resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetworksClientGetResponse, error) {
	logger := composed.LoggerFromCtx(ctx)

	vpcClientGetResponse, error := c.VirtualNetworkClient.Get(ctx, resourceGroupName, virtualNetworkName, nil)

	if error != nil {
		logger.Error(error, "Failed to get Azure Redis VPC")
		return nil, error
	}

	return &vpcClientGetResponse, nil
}

func (c *networkClient) GetResourceGroup(ctx context.Context, resourceGroupName string) (*armResources.ResourceGroupsClientGetResponse, error) {
	logger := composed.LoggerFromCtx(ctx)

	resourceGroupClientGetResponse, error := c.ResourceGroupClient.Get(ctx, resourceGroupName, nil)

	if error != nil {
		logger.Error(error, "Failed to get Azure Redis ResourceGroup")
		return nil, error
	}

	return &resourceGroupClientGetResponse, nil
}

func (c *networkClient) CreateResourceGroup(ctx context.Context, name string, location string) error {
	logger := composed.LoggerFromCtx(ctx)

	resourceGroup := armResources.ResourceGroup{Location: to.Ptr(location)}
	_, error := c.ResourceGroupClient.CreateOrUpdate(ctx, name, resourceGroup, nil)

	if error != nil {
		logger.Error(error, "Failed to create Azure Redis resource group")
		return error
	}

	return nil
}

func (c *networkClient) DeleteResourceGroup(ctx context.Context, name string) error {
	logger := composed.LoggerFromCtx(ctx)

	_, error := c.ResourceGroupClient.BeginDelete(ctx, name, nil)

	if error != nil {
		logger.Error(error, "Failed to delete Azure Redis iprange resource group")
		return error
	}

	return nil
}

func (c *networkClient) DeleteSubnet(ctx context.Context, resourceGroupName, virtualNetworkName, subnetName string) error {
	logger := composed.LoggerFromCtx(ctx)

	_, error := c.SubnetClient.BeginDelete(ctx, resourceGroupName, virtualNetworkName, subnetName, nil)

	if error != nil {
		logger.Error(error, "Failed to delete Azure Redis iprange subnet")
		return error
	}

	return nil
}
