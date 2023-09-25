package reconcilerutils

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/instaclustr/operator/pkg/models"
)

const MaxReconcileRequeueRetries = 5

type ReconcilerWithLimitOptions struct {
	EventRecorder record.EventRecorder
	Client        client.Client
	MaxRetries    int
	For           client.Object
	Scheme        *runtime.Scheme
}

func ReconcilerWithLimit(r reconcile.Reconciler, opts ...ReconcilerWithLimitOptions) reconcile.Func {
	limiter := &ReconcileRequeueLimiter{
		m: make(map[types.NamespacedName]int),
	}

	var (
		eventRecorder record.EventRecorder
		client        client.Client
		withEvents    bool
		gvk           schema.GroupVersionKind
	)

	if opts != nil {
		opt := opts[0]
		if opt.EventRecorder != nil && opt.Client != nil && opt.For != nil && opt.Scheme != nil {
			withEvents = true
			eventRecorder = opt.EventRecorder
			client = opt.Client

			var err error
			gvk, err = apiutil.GVKForObject(opt.For, opt.Scheme)
			if err != nil {
				panic(err)
			}
		}

		if opt.MaxRetries < 1 {
			limiter.maxRetries = MaxReconcileRequeueRetries
		} else {
			limiter.maxRetries = opt.MaxRetries
		}
	}

	return func(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
		if !limiter.RequeueAllowed(req.NamespacedName) {
			limiter.Reset(req.NamespacedName)

			l := log.FromContext(ctx).WithValues("component", "reconcile-requeue-limiter")

			l.Info("Amount of reconcile requeue retries reached its maximum for the resource",
				"req", req.NamespacedName,
			)

			if withEvents {
				obj := &unstructured.Unstructured{}
				obj.SetGroupVersionKind(gvk)

				err := client.Get(ctx, req.NamespacedName, obj)
				if err != nil {
					l.Error(err, "Failed to get the resource", "req", req.NamespacedName)
					return models.ExitReconcile, nil
				}

				eventRecorder.Eventf(obj, models.Warning, models.GenericEvent,
					"Amount of reconcile requeue retries reached its maximum: %v", limiter.maxRetries,
				)
			}

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
