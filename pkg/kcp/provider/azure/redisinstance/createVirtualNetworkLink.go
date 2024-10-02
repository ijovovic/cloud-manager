package redisinstance

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createVirtualNetworkLink(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.virtualNetworkLink != nil {
		return nil, nil
	}

	logger.Info("Creating Azure Private VirtualNetworkLink")
	resourceGroupName := state.resourceGroupName
	privateDnsZoneName := "privatelink.redis.cache.windows.net"
	virtualNetworkLinkName := "kyma-network-link"
	kymaNetworkName := state.Scope().Spec.Scope.Azure.VpcNetwork

	virtualNetworkLink := armprivatedns.VirtualNetworkLink{
		Location: ptr.To("global"),
		Properties: &armprivatedns.VirtualNetworkLinkProperties{
			VirtualNetwork: &armprivatedns.SubResource{
				ID: ptr.To(azureutil.NewVirtualNetworkResourceId(state.Scope().Spec.Scope.Azure.SubscriptionId,
					state.Scope().Spec.Scope.Azure.VpcNetwork, kymaNetworkName).String()),
			},
			RegistrationEnabled: ptr.To(false),
		},
	}
	err := state.client.CreateVirtualNetworkLink(ctx, resourceGroupName, privateDnsZoneName, virtualNetworkLinkName, virtualNetworkLink)

	if err != nil {
		logger.Error(err, "Error creating Azure PrivateVirtualNetworkLink")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed creating Azure PrivateVirtualNetworkLink: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisInstance status due failed azure PrivateVirtualNetworkLink creation",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
