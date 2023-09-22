package reconcilerutils

import (
	"context"
	"github.com/instaclustr/operator/pkg/models"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const MaxReconcileRequeueRetries = 5

func ReconcilerWithLimit(r reconcile.Reconciler) reconcile.Func {
	limiter := &ReconcileRequeueLimiter{
		maxRetries: MaxReconcileRequeueRetries,
		m:          make(map[types.NamespacedName]int),
	}

	return func(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
		if !limiter.RequeueAllowed(req.NamespacedName) {
			l := log.FromContext(ctx)
			l.Info("Amount of reconcile requeue retries reached its maximum for the resource",
				"req", req.NamespacedName,
			)

			limiter.Reset(req.NamespacedName)

			return models.ExitReconcile, nil
		}

		result, err := r.Reconcile(ctx, req)
		if err != nil || result == models.ReconcileRequeue {
			limiter.Requeue(req.NamespacedName)
			return result, err
		}

		limiter.Reset(req.NamespacedName)

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
