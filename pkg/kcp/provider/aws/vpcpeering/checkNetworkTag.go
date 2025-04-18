package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	peeringconfig "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/config"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func checkNetworkTag(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	obj := state.ObjAsVpcPeering()

	if state.remoteVpcPeering != nil {
		return nil, nil
	}

	var kv []any

	for _, t := range state.remoteVpc.Tags {
		kv = append(kv, *t.Key, *t.Value)
	}

	// If VpcNetwork is found but tags don't match user can recover by adding tag to remote VPC network so, we are
	// adding stop with requeue delay of one minute.
	hasShootTag := awsutil.HasEc2Tag(state.remoteVpc.Tags, peeringconfig.VpcPeeringConfig.NetworkTag)
	if !hasShootTag {
		hasShootTag = awsutil.HasEc2Tag(state.remoteVpc.Tags, state.Scope().Spec.ShootName)
	}

	if !hasShootTag {

		logger.Info("Loaded remote VPC Network have no matching tags", kv...)

		return composed.PatchStatus(obj).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonFailedLoadingRemoteVpcNetwork,
				Message: "Loaded remote Vpc network has no matching tags",
			}).
			ErrorLogMessage("Error updating VpcPeering status due to remote vpc network tag mismatch").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	return nil, nil
}
