package criprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Reconciler struct {
}

func (r *Reconciler) Run(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	state := r.newState()
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *Reconciler) newState() *State {
	return nil
}

func (r *Reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"crIpRangeMain",
		composed.LoadObj,
		validateCidr,
		copyCidrToStatus,
		preventCidrChange,
		addFinalizer,
		loadKcpIpRange,
		createKcpIpRange,
		deleteKcpIpRange,
		removeFinalizer,
		updateStatus,
		composed.StopAndForgetAction,
	)
}
