package iprange

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureUtil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createSubnet(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vpc != nil {
		return nil, nil
	}

	logger.Info("Creating Azure iprange vpc")

	virtualNetworkName := azureUtil.GetPredictableResourceName("vpc", string(state.ObjAsIpRange().GetUID()))
	resourceGroupName := azureUtil.GetPredictableResourceName("iprange", string(state.ObjAsIpRange().GetUID()))
	subnetName := azureUtil.GetPredictableResourceName("iprange", string(state.ObjAsIpRange().GetUID()))
	location := state.Scope().Spec.Region

	error := state.client.CreateSubnet(ctx, resourceGroupName, virtualNetworkName, subnetName, location, "10.255.4.0/22")
	if error != nil {
		logger.Error(error, "Error crating Azure subnet")
		meta.SetStatusCondition(state.ObjAsIpRange().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed creating AzureRedis subnet: %s", error),
		})
		error = state.UpdateObjStatus(ctx)
		if error != nil {
			return composed.LogErrorAndReturn(error,
				"Error updating IpRange status due failed azure ip range subnet create",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}
	// we have just created the subnet, requeue so the ip range can be loaded
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
