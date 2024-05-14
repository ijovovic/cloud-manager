package iprange

import (
	"context"
	"fmt"
	"github.com/3th1nk/cidr"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func preventCidrOverlap(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	// WARNING! Overlap is symmetrical!!!
	// If new iprange overlaps with some old iprange, then the old one also overlaps with the new one
	// Desired behavior is that old range remains Ready, and only new one to end up with warning
	// Also, a valid use case would be that user deletes the old range and then the new range
	// should be able to step out of the overlap warning and become ready

	allIpRanges := &cloudresourcesv1beta1.IpRangeList{}
	err := state.Cluster().K8sClient().List(ctx, allIpRanges)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing all SKR IpRanges to check CIDR overlap", composed.StopWithRequeue, ctx)
	}

	myCidr, err := cidr.Parse(state.ObjAsIpRange().Spec.Cidr)
	if err != nil {
		if err != nil {
			state.ObjAsIpRange().Status.State = cloudresourcesv1beta1.StateError
			return composed.UpdateStatus(state.ObjAsIpRange()).
				SetCondition(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonInvalidCidr,
					Message: fmt.Sprintf("CIDR %s has invalid syntax", state.ObjAsIpRange().Spec.Cidr),
				}).
				RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
				ErrorLogMessage("Error updating IpRange status with invalid CIDR syntax").
				SuccessLogMsg("Forgetting IpRange with invalid Cidr syntax").
				Run(ctx, state)
		}
	}

	myDate := state.ObjAsIpRange().CreationTimestamp

	for _, ipRange := range allIpRanges.Items {
		if ipRange.Name == state.ObjAsIpRange().Name &&
			ipRange.Namespace == state.ObjAsIpRange().Namespace {
			// skip me (the reconciled IpRange) - I always overlap with myself
			continue
		}

		hisCidr, err := cidr.Parse(ipRange.Spec.Cidr)
		if err != nil {
			continue
		}

		if util.CidrEquals(myCidr.CIDR(), hisCidr.CIDR()) ||
			util.CidrOverlap(myCidr.CIDR(), hisCidr.CIDR()) {

			logger = logger.WithValues(
				"cidr", state.ObjAsIpRange().Spec.Cidr,
				"overlappingCidr", ipRange.Spec.Cidr,
				"overlappingIpRange", fmt.Sprintf("%s/%s", ipRange.Namespace, ipRange.Name),
			)

			hisDate := ipRange.CreationTimestamp

			// put error on NEWER range only
			// pass-on and do not modify the OLDER range
			if myDate.Before(&hisDate) {
				// Im older, skip me
				logger.Info("Letting older range with overlapping CIDR pass")
				return nil, nil
			}

			// Im newer, set me overlap error

			ctx = composed.LoggerIntoCtx(ctx, logger)

			state.ObjAsIpRange().Status.State = cloudresourcesv1beta1.StateError
			return composed.UpdateStatus(state.ObjAsIpRange()).
				SetCondition(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonCidrOverlap,
					Message: fmt.Sprintf("CIDR overlaps with %s/%s", ipRange.Namespace, ipRange.Name),
				}).
				RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
				ErrorLogMessage("Error updating IpRange status with CIDR overlap error").
				SuccessLogMsg("Forgetting IpRange with Cidr overlap").
				SuccessError(composed.StopAndForget).
				Run(ctx, state)
		}
	}

	return nil, nil
}
