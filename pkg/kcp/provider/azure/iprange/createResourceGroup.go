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

func createResourceGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.resourceGroup != nil {
		return nil, nil
	}

	logger.Info("Creating Azure iprange resourceGroup")

	resourceGroupName := azureUtil.GetPredictableResourceName("iprange", string(state.ObjAsIpRange().GetUID()))
	location := state.Scope().Spec.Region

	error := state.client.CreateResourceGroup(ctx, resourceGroupName, location)
	if error != nil {
		if azuremeta.IsNotFound(error) {
			return nil, nil
		}

		logger.Error(error, "Error crating Azure resource group")
		meta.SetStatusCondition(state.ObjAsIpRange().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonCanNotCreateResourceGroup,
			Message: fmt.Sprintf("Failed creating Azure iprange resource group: %s", error),
		})
		error = state.UpdateObjStatus(ctx)
		if error != nil {
			return composed.LogErrorAndReturn(error,
				"Error updating iprange status due failed azure iprange resource group create",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}
	// we have just created the group, requeue so the resource group can be loaded
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
