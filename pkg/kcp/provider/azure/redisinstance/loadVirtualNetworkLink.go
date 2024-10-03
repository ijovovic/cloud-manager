package redisinstance

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadVirtualNetworkLink(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.virtualNetworkLink != nil {
		logger.Info("Azure Private VirtualNetworkLink already loaded")
		return nil, nil
	}

	logger.Info("Loading Azure Private VirtualNetworkLink")

	resourceGroupName := state.resourceGroupName
	privateZoneName := azureutil.GetDefaultPrivateDnsZoneName()
	virtualNetworkLinkName := azureutil.GetDefaultVirtualNetworkLinkName()

	privateVirtualNetworkLinkInstance, err := state.client.GetVirtualNetworkLink(ctx, resourceGroupName, privateZoneName, virtualNetworkLinkName)
	if err != nil {
		if azuremeta.IsNotFound(err) {
			logger.Info("Azure Private VirtualNetworkLink instance not found")
			return nil, nil
		}

		logger.Error(err, "Error loading Azure Private VirtualNetworkLink")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed loading AzureRedis: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisInstance status due failed azure Private VirtualNetworkLink loading",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}
	state.virtualNetworkLink = privateVirtualNetworkLinkInstance

	return nil, nil
}
