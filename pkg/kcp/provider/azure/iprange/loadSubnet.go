package iprange

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureUtil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadSubnet(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vpc != nil {
		logger.Info("Azure ipRange vpcName already loaded")
		return nil, nil
	}

	logger.Info("Loading Azure ipRange vpcName")

	virtualNetworkName := azureUtil.GetPredictableResourceName("vpc", string(state.ObjAsIpRange().GetUID()))
	resourceGroupName := azureUtil.GetPredictableResourceName("iprange", string(state.ObjAsIpRange().GetUID()))
	virtualNetworksClientGetResponse, error := state.client.GetVpc(ctx, resourceGroupName, virtualNetworkName)
	if error != nil {
		if azuremeta.IsNotFound(error) {
			return nil, nil
		}

		logger.Error(error, "Error loading Azure resource group")
		meta.SetStatusCondition(state.ObjAsIpRange().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed loading Azureiprange vpc: %s", error),
		})
		error = state.UpdateObjStatus(ctx)
		if error != nil {
			return composed.LogErrorAndReturn(error,
				"Error updating ipRange status due failed azure ip range vpc loading",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	state.vpc = &virtualNetworksClientGetResponse.VirtualNetwork

	return nil, nil
}
