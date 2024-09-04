package iprange

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		logger := composed.LoggerFromCtx(ctx)

		state, err := stateFactory.NewState(ctx, st.(types.State), logger)
		if err != nil {
			ipRange := st.Obj().(*v1beta1.IpRange)
			return composed.UpdateStatus(ipRange).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonGcpError,
					Message: err.Error(),
				}).
				SuccessError(composed.StopAndForget).
				SuccessLogMsg(fmt.Sprintf("Error creating new Azure IpRange state: %s", err)).
				Run(ctx, st)
		}

		return composed.ComposeActions(
			"azureIpRange",
			actions.AddFinalizer,
			loadResourceGroup,
			loadSubnet,
			composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"azure-ipRange-create",
					// allocate cidr
					// check all the overalpings possible
					createResourceGroup,
					createSubnet,
					//updateStatus,
				),
				composed.ComposeActions(
					"azure-ipRange-delete",
					//deleteRedis,
					//waitRedisDeleted,
					//deleteResourceGroup,
					//waitResourceGroupDeleted,
					actions.RemoveFinalizer,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}
