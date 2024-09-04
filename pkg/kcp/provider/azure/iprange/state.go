package iprange

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	armResources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/go-logr/logr"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azureiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/iprange/client"
)

type State struct {
	iprangetypes.State

	client         azureiprangeclient.Client
	provider       azureclient.SkrClientProvider[azureiprangeclient.Client]
	clientId       string
	clientSecret   string
	subscriptionId string
	tenantId       string

	resourceGroup *armResources.ResourceGroup
	vpc           *armnetwork.VirtualNetwork
}

type StateFactory interface {
	NewState(ctx context.Context, state iprangetypes.State, logger logr.Logger) (*State, error)
}

type stateFactory struct {
	skrProvider azureclient.SkrClientProvider[azureiprangeclient.Client]
}

func NewStateFactory(skrProvider azureclient.SkrClientProvider[azureiprangeclient.Client]) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
	}
}

func (f *stateFactory) NewState(ctx context.Context, iprangeState iprangetypes.State, logger logr.Logger) (*State, error) {

	clientId := azureconfig.AzureConfig.DefaultCreds.ClientId
	clientSecret := azureconfig.AzureConfig.DefaultCreds.ClientSecret
	subscriptionId := iprangeState.Scope().Spec.Scope.Azure.SubscriptionId
	tenantId := iprangeState.Scope().Spec.Scope.Azure.TenantId

	c, err := f.skrProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)

	if err != nil {
		return nil, err
	}

	return newState(iprangeState, c, f.skrProvider, clientId, clientSecret, subscriptionId, tenantId), nil
}

func newState(state iprangetypes.State,
	client azureiprangeclient.Client,
	provider azureclient.SkrClientProvider[azureiprangeclient.Client],
	clientId string,
	clientSecret string,
	subscriptionId string,
	tenantId string) *State {
	return &State{
		State:          state,
		client:         client,
		provider:       provider,
		clientId:       clientId,
		clientSecret:   clientSecret,
		subscriptionId: subscriptionId,
		tenantId:       tenantId,
	}
}
