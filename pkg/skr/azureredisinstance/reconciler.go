package azureredisinstance

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewReconcilerFactory() skrruntime.ReconcilerFactory {
	return &reconcilerFactory{}
}

type reconcilerFactory struct {
}

func (f *reconcilerFactory) New(args skrruntime.ReconcilerArguments) reconcile.Reconciler {
	return &reconciler{
		factory: newStateFactory(
			composed.NewStateFactory(composed.NewStateClusterFromCluster(args.SkrCluster)),
			args.KymaRef,
			composed.NewStateClusterFromCluster(args.KcpCluster),
		),
	}
}

type reconciler struct {
	factory *stateFactory
}

func (r *reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	state := r.factory.NewState(request)
	action := r.newAction()

	return composed.Handle(action(ctx, state))
}

func (r *reconciler) newAction() composed.Action {
	return composed.ComposeActions(
		"azureRedisInstance",
		feature.LoadFeatureContextFromObj(&cloudresourcesv1beta1.AzureRedisInstance{}),
		composed.LoadObj,

		updateId,
		loadKcpRedisInstance,
		loadAuthSecret,

		composed.IfElse(composed.Not(composed.MarkedForDeletionPredicate),
			composed.ComposeActions(
				"azureRedisInstance-create",
				actions.AddFinalizer,
				createKcpRedisInstance,
				waitKcpStatusUpdate,
				updateStatus,
				waitSkrStatusReady,
				createAuthSecret,
			),
			composed.ComposeActions(
				"azureRedisInstance-delete",
				removeAuthSecretFinalizer,
				deleteAuthSecret,
				waitAuthSecretDeleted,
				deleteKcpRedisInstance,
				waitKcpRedisInstanceDeleted,
				actions.RemoveFinalizer,
			),
		),

		composed.StopAndForgetAction,
	)
}
