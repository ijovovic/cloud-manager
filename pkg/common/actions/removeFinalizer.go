package actions

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func RemoveFinalizer(ctx context.Context, state composed.State) (error, context.Context) {

	//KindsFromObject is not being deleted, don't remove finalizer
	if !composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	//If finalizer not already present, don't remove it .
	if !controllerutil.ContainsFinalizer(state.Obj(), cloudcontrolv1beta1.FinalizerName) {
		return nil, nil
	}

	//Remove finalizer
	controllerutil.RemoveFinalizer(state.Obj(), cloudcontrolv1beta1.FinalizerName)
	if err := state.UpdateObj(ctx); err != nil {
		return composed.LogErrorAndReturn(err, "Error removing Finalizer", composed.StopWithRequeue, ctx)
	}

	//stop reconciling loop.
	return composed.StopAndForget, nil
}
