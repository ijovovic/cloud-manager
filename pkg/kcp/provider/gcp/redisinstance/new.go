package redisinstance

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/kcp/redisinstance/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func New(stateFactory StateFactory) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {

		state, err := stateFactory.NewState(ctx, st.(types.State))
		if err != nil {
			composed.LoggerFromCtx(ctx).Error(err, "Failed to bootstrap GCP RedisInstance state")
			redisInstance := st.Obj().(*v1beta1.RedisInstance)
			redisInstance.Status.State = cloudcontrolv1beta1.StateError
			return composed.UpdateStatus(redisInstance).
				SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  v1beta1.ReasonCloudProviderError,
					Message: "Failed to create RedisInstance state",
				}).
				SuccessError(composed.StopAndForget).
				SuccessLogMsg(fmt.Sprintf("Error creating new GCP RedisInstance state: %s", err)).
				Run(ctx, st)
		}

		return composed.ComposeActions(
			"redisInstance",
			actions.AddFinalizer,
			loadRedis,
			composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
				composed.ComposeActions(
					"redisInstance-create",
					createRedis,
					updateStatusId,
					addUpdatingCondition,
					waitRedisAvailable,
					modifyMemorySizeGb,
					modifyMemoryReplicaCount,
					modifyRedisConfigs,
					modifyMaintenancePolicy,
					modifyAuthEnabled,
					updateRedis,
					upgradeRedis,
					updateStatus,
				),
				composed.ComposeActions(
					"redisInstance-delete",
					removeReadyCondition,
					deleteRedis,
					waitRedisDeleted,
					actions.RemoveFinalizer,
				),
			),
			composed.StopAndForgetAction,
		)(ctx, state)
	}
}
