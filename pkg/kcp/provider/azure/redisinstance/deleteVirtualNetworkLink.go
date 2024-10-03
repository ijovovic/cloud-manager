package redisinstance

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureMeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteVirtualNetworkLink(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.virtualNetworkLink == nil {
		return nil, nil
	}

	if *state.virtualNetworkLink.Properties.ProvisioningState == "Deleting" {
		return nil, nil
	}

	logger.Info("Deleting Azure VirtualNetworkLink")

	resourceGroupName := state.resourceGroupName
	privateDnsZoneName := azureutil.GetDefaultPrivateDnsZoneName()
	virtualNetworkLinkname := azureutil.GetDefaultVirtualNetworkLinkName()

	err := state.client.DeleteVirtualNetworkLink(ctx, resourceGroupName, privateDnsZoneName, virtualNetworkLinkname)
	if err != nil {
		if azureMeta.IsNotFound(err) {
			return nil, nil
		}

		logger.Error(err, "Error deleting Azure VirtualNetworkLink")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed deleting Azure VirtualNetworkLink: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisInstance status due failed azure VirtualNetworkLink deleting",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
