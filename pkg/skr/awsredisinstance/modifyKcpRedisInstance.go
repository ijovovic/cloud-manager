package awsredisinstance

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func modifyKcpRedisInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	awsRedisInstance := state.ObjAsAwsRedisInstance()

	if state.KcpRedisInstance == nil {
		logger.Error(fmt.Errorf("kcpRedisInstance not found"), "KcpRedisInstance not found")
		return composed.StopWithRequeue, nil
	}

	if !areMapsDifferent(state.KcpRedisInstance.Spec.Instance.Aws.Parameters, awsRedisInstance.Spec.Parameters) {
		return nil, nil
	}

	state.KcpRedisInstance.Spec.Instance.Aws.Parameters = awsRedisInstance.Spec.Parameters
	err := state.KcpCluster.K8sClient().Update(ctx, state.KcpRedisInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating KCP RedisInstance", composed.StopWithRequeue, ctx)
	}

	return nil, nil
}
