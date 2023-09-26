package reconcilerutils

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/instaclustr/operator/pkg/models"
)

const DefaultMaxReconcileRequeueRetries = 5

type reconcilerBuilder struct {
	maxRetries int

	eventRecorder record.EventRecorder
	client        client.Client
	forObject     client.Object
	scheme        *runtime.Scheme
	withEvents    bool
}

func NewReconciler(scheme *runtime.Scheme) *reconcilerBuilder {
	return &reconcilerBuilder{scheme: scheme}
}

func (bld *reconcilerBuilder) WithReconcileRequeueLimit(maxRetries int) *reconcilerBuilder {
	bld.maxRetries = maxRetries
	return bld
}

func (bld *reconcilerBuilder) For(object client.Object) *reconcilerBuilder {
	bld.forObject = object
	return bld
}

type WithEventsOptions struct {
	EventRecorder record.EventRecorder
	Client        client.Client
}

func (bld *reconcilerBuilder) WithEvents(opt *WithEventsOptions) *reconcilerBuilder {
	bld.withEvents = true
	bld.client = opt.Client
	bld.eventRecorder = opt.EventRecorder
	return bld
}

func (bld *reconcilerBuilder) Complete(r reconcile.Reconciler) reconcile.Func {
	requeueLimiter := ReconcileRequeueLimiter{
		m: make(map[types.NamespacedName]int),
	}

	if bld.maxRetries < 1 {
		bld.maxRetries = DefaultMaxReconcileRequeueRetries
	}

	gvk, err := apiutil.GVKForObject(bld.forObject, bld.scheme)
	if err != nil {
		panic(fmt.Errorf("failed to create reconciler, err: %w", err))
	}

	return func(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
		if !requeueLimiter.RequeueAllowed(req.NamespacedName) {
			requeueLimiter.Reset(req.NamespacedName)

			l := log.FromContext(ctx).WithValues("component", "reconcile-requeue-limiter")

			l.Info("Amount of reconcile requeue retries reached its maximum for the resource",
				"req", req.NamespacedName,
			)

			if bld.withEvents {
				obj := &unstructured.Unstructured{}
				obj.SetGroupVersionKind(gvk)

				err := bld.client.Get(ctx, req.NamespacedName, obj)
				if err != nil {
					l.Error(err, "Failed to get the resource", "req", req.NamespacedName)
					return models.ExitReconcile, nil
				}

				bld.eventRecorder.Eventf(obj, models.Warning, models.GenericEvent,
					"Amount of reconcile requeue retries reached its maximum: %v", requeueLimiter.maxRetries,
				)
			}

			return models.ExitReconcile, nil
		}

		result, err := r.Reconcile(ctx, req)
		if err != nil || result == models.ReconcileRequeue {
			requeueLimiter.Requeue(req.NamespacedName)
			return result, err
		}

		requeueLimiter.Reset(req.NamespacedName)

		return result, nil
	}

}

type ReconcileRequeueLimiter struct {
	maxRetries int
	m          map[types.NamespacedName]int
}

func (r *ReconcileRequeueLimiter) Requeue(key types.NamespacedName) {
	r.m[key]++
}

func (r *ReconcileRequeueLimiter) RequeueAllowed(key types.NamespacedName) bool {
	return r.m[key] < r.maxRetries
}

func (r *ReconcileRequeueLimiter) Reset(key types.NamespacedName) {
	delete(r.m, key)
}
