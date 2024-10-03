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

func deletePrivateDnsZone(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.privateDnsZone == nil {
		return nil, nil
	}

	if isPrivateDnsZoneSafeToDelete(state) == false {
		return nil, nil
	}

	if *state.privateDnsZone.Properties.ProvisioningState == "Deleting" {
		return nil, nil
	}

	logger.Info("Deleting Azure PrivateDnsZone")

	resourceGroupName := state.resourceGroupName
	privateDnsZoneName := azureutil.GetDefaultPrivateDnsZoneName()

	err := state.client.DeletePrivateDnsZone(ctx, resourceGroupName, privateDnsZoneName)
	if err != nil {
		if azureMeta.IsNotFound(err) {
			return nil, nil
		}

		logger.Error(err, "Error deleting Azure PrivateDnsZone")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed deleting Azure PrivateDnsZone: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisInstance status due failed azure PrivateDnsZone deleting",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}

func isPrivateDnsZoneSafeToDelete(state *State) bool {
	// Unused privateDNsZone has single SOA record, anything on top of that indicates it should not be deleted
	return *state.privateDnsZone.Properties.NumberOfRecordSets == 1
}
