/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clusters

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/ratelimiter"

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/exposeservice"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	rlimiter "github.com/instaclustr/operator/pkg/ratelimiter"
	"github.com/instaclustr/operator/pkg/scheduler"
)

// CadenceReconciler reconciles a Cadence object
type CadenceReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	Scheduler     scheduler.Interface
	EventRecorder record.EventRecorder
	RateLimiter   ratelimiter.RateLimiter
}

//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cadences,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cadences/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusters.instaclustr.com,resources=cadences/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create
//+kubebuilder:rbac:groups="",resources=endpoints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;update;patch;watch
//+kubebuilder:rbac:groups=cdi.kubevirt.io,resources=datavolumes,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachineinstances,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete;deletecollection
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete;deletecollection

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *CadenceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	c := &v1beta1.Cadence{}
	err := r.Client.Get(ctx, req.NamespacedName, c)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("Cadence resource is not found",
				"resource name", req.NamespacedName,
			)
			return ctrl.Result{}, nil
		}

		l.Error(err, "Unable to fetch Cadence resource",
			"resource name", req.NamespacedName,
		)
		return ctrl.Result{}, err
	}

	switch c.Annotations[models.ResourceStateAnnotation] {
	case models.CreatingEvent:
		return r.handleCreateCluster(ctx, c, l)
	case models.UpdatingEvent:
		return r.handleUpdateCluster(ctx, c, req, l)
	case models.DeletingEvent:
		return r.handleDeleteCluster(ctx, c, l)
	case models.GenericEvent:
		l.Info("Generic event isn't handled",
			"request", req,
			"event", c.Annotations[models.ResourceStateAnnotation],
		)

		return ctrl.Result{}, nil
	default:
		l.Info("Unknown event isn't handled",
			"request", req,
			"event", c.Annotations[models.ResourceStateAnnotation],
		)

		return ctrl.Result{}, nil
	}
}

func (r *CadenceReconciler) handleCreateCluster(
	ctx context.Context,
	c *v1beta1.Cadence,
	l logr.Logger,
) (ctrl.Result, error) {
	if c.Status.ID == "" {
		patch := c.NewPatch()

		for _, packagedProvisioning := range c.Spec.PackagedProvisioning {
			requeueNeeded, err := r.preparePackagedSolution(ctx, c, packagedProvisioning)
			if err != nil {
				l.Error(err, "Cannot prepare packaged solution for Cadence cluster",
					"cluster name", c.Spec.Name,
				)

				r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
					"Cannot prepare packaged solution for Cadence cluster. Reason: %v", err)

				return ctrl.Result{}, err
			}

			if requeueNeeded {
				l.Info("Waiting for bundled clusters to be created",
					"c cluster name", c.Spec.Name)

				r.EventRecorder.Event(c, models.Normal, "Waiting",
					"Waiting for bundled clusters to be created")

				return models.ReconcileRequeue, nil
			}
		}

		l.Info(
			"Creating Cadence cluster",
			"cluster name", c.Spec.Name,
			"data centres", c.Spec.DataCentres,
		)

		cadenceAPISpec, err := c.Spec.ToInstAPI(ctx, r.Client)
		if err != nil {
			l.Error(err, "Cannot convert Cadence cluster manifest to API spec",
				"cluster manifest", c.Spec)

			r.EventRecorder.Eventf(c, models.Warning, models.ConversionFailed,
				"Cluster convertion from the Instaclustr API to k8s resource is failed. Reason: %v", err)

			return ctrl.Result{}, err
		}

		id, err := r.API.CreateCluster(instaclustr.CadenceEndpoint, cadenceAPISpec)
		if err != nil {
			l.Error(
				err, "Cannot create Cadence cluster",
				"c manifest", c.Spec,
			)
			r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
				"Cluster creation on the Instaclustr is failed. Reason: %v", err)

			return ctrl.Result{}, err
		}

		c.Status.ID = id
		err = r.Status().Patch(ctx, c, patch)
		if err != nil {
			l.Error(err, "Cannot update Cadence cluster status",
				"cluster name", c.Spec.Name,
				"cluster status", c.Status,
			)

			r.EventRecorder.Eventf(c, models.Warning, models.PatchFailed,
				"Cluster resource status patch is failed. Reason: %v", err)

			return ctrl.Result{}, err
		}

		if c.Spec.Description != "" {
			err = r.API.UpdateDescriptionAndTwoFactorDelete(instaclustr.ClustersEndpointV1, id, c.Spec.Description, nil)
			if err != nil {
				l.Error(err, "Cannot update Cadence cluster description and TwoFactorDelete",
					"cluster name", c.Spec.Name,
					"description", c.Spec.Description,
					"twoFactorDelete", c.Spec.TwoFactorDelete,
				)

				r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
					"Cluster description and TwoFactoDelete update is failed. Reason: %v", err)
			}
		}

		c.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(c, models.DeletionFinalizer)

		err = r.Patch(ctx, c, patch)
		if err != nil {
			l.Error(err, "Cannot patch Cadence cluster",
				"cluster name", c.Spec.Name, "patch", patch)

			r.EventRecorder.Eventf(c, models.Warning, models.PatchFailed,
				"Cluster resource status patch is failed. Reason: %v", err)

			return ctrl.Result{}, err
		}

		l.Info(
			"Cadence resource has been created",
			"cluster name", c.Name,
			"cluster ID", c.Status.ID,
			"kind", c.Kind,
			"api version", c.APIVersion,
			"namespace", c.Namespace,
		)

		r.EventRecorder.Eventf(c, models.Normal, models.Created,
			"Cluster creation request is sent. Cluster ID: %s", id)
	}

	if c.Status.State != models.DeletedStatus {
		err := r.startSyncJob(c)
		if err != nil {
			l.Error(err, "Cannot start cluster status job",
				"c cluster ID", c.Status.ID,
			)

			r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
				"Cluster status check job is failed. Reason: %v", err)

			return ctrl.Result{}, err
		}

		r.EventRecorder.Event(c, models.Normal, models.Created,
			"Cluster sync job is started")
	}

	return ctrl.Result{}, nil
}

func (r *CadenceReconciler) handleUpdateCluster(
	ctx context.Context,
	c *v1beta1.Cadence,
	req ctrl.Request,
	l logr.Logger,
) (ctrl.Result, error) {
	instaModel, err := r.API.GetCadence(c.Status.ID)
	if err != nil {
		l.Error(
			err, "Cannot get Cadence cluster from the Instaclustr API",
			"cluster name", c.Spec.Name,
			"cluster ID", c.Status.ID,
		)

		r.EventRecorder.Eventf(c, models.Warning, models.FetchFailed,
			"Cluster fetch from the Instaclustr API is failed. Reason: %v", err)

		return ctrl.Result{}, err
	}

	iCadence := &v1beta1.Cadence{}
	iCadence.FromInstAPI(instaModel)

	if c.Annotations[models.ExternalChangesAnnotation] == models.True ||
		r.RateLimiter.NumRequeues(req) == rlimiter.DefaultMaxTries {
		return handleExternalChanges[v1beta1.CadenceSpec](r.EventRecorder, r.Client, c, iCadence, l)
	}

	if c.Spec.ClusterSettingsNeedUpdate(&iCadence.Spec.GenericClusterSpec) {
		l.Info("Updating cluster settings",
			"instaclustr description", iCadence.Spec.Description,
			"instaclustr two factor delete", iCadence.Spec.TwoFactorDelete)

		err = r.API.UpdateClusterSettings(c.Status.ID, c.Spec.ClusterSettingsUpdateToInstAPI())
		if err != nil {
			l.Error(err, "Cannot update cluster settings",
				"cluster ID", c.Status.ID, "cluster spec", c.Spec)
			r.EventRecorder.Eventf(c, models.Warning, models.UpdateFailed,
				"Cannot update cluster settings. Reason: %v", err)

			return ctrl.Result{}, err
		}
	}

	l.Info("Update request to Instaclustr API has been sent",
		"spec data centres", c.Spec.DataCentres,
		"resize settings", c.Spec.ResizeSettings,
	)
	err = r.API.UpdateCluster(c.Status.ID, instaclustr.CadenceEndpoint, c.Spec.NewDCsUpdate())
	if err != nil {
		l.Error(err, "Cannot update Cadence cluster",
			"cluster ID", c.Status.ID,
			"update request", c.Spec.NewDCsUpdate(),
		)

		r.EventRecorder.Eventf(c, models.Warning, models.UpdateFailed,
			"Cluster update on the Instaclustr API is failed. Reason: %v", err)

		return ctrl.Result{}, err
	}

	patch := c.NewPatch()
	c.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
	err = r.Patch(ctx, c, patch)
	if err != nil {
		l.Error(err, "Cannot patch Cadence cluster",
			"cluster name", c.Spec.Name,
			"patch", patch)

		r.EventRecorder.Eventf(c, models.Warning, models.PatchFailed,
			"Cluster resource patch is failed. Reason: %v", err)

		return ctrl.Result{}, err
	}

	l.Info(
		"Cluster has been updated",
		"cluster name", c.Spec.Name,
		"cluster ID", c.Status.ID,
		"data centres", c.Spec.DataCentres,
	)

	return ctrl.Result{}, nil
}

func (r *CadenceReconciler) handleDeleteCluster(
	ctx context.Context,
	c *v1beta1.Cadence,

	l logr.Logger,
) (ctrl.Result, error) {
	_, err := r.API.GetCadence(c.Status.ID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		l.Error(
			err, "Cannot get Cadence cluster status from the Instaclustr API",
			"cluster name", c.Spec.Name,
			"cluster ID", c.Status.ID,
		)

		r.EventRecorder.Eventf(c, models.Warning, models.FetchFailed,
			"Cluster resource fetch from the Instaclustr API is failed. Reason: %v", err)

		return ctrl.Result{}, err
	}

	if !errors.Is(err, instaclustr.NotFound) {
		l.Info("Sending cluster deletion to the Instaclustr API",
			"cluster name", c.Spec.Name,
			"cluster ID", c.Status.ID)

		err = r.API.DeleteCluster(c.Status.ID, instaclustr.CadenceEndpoint)
		if err != nil {
			l.Error(err, "Cannot delete Cadence cluster",
				"cluster name", c.Spec.Name,
				"cluster status", c.Status,
			)

			r.EventRecorder.Eventf(c, models.Warning, models.DeletionFailed,
				"Cluster deletion is failed on the Instaclustr. Reason: %v", err)

			return ctrl.Result{}, err
		}

		r.EventRecorder.Event(c, models.Normal, models.DeletionStarted,
			"Cluster deletion request is sent to the Instaclustr API.")

		if c.Spec.TwoFactorDelete != nil {
			patch := c.NewPatch()

			c.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
			c.Annotations[models.ClusterDeletionAnnotation] = models.Triggered
			err = r.Patch(ctx, c, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster resource",
					"cluster name", c.Spec.Name,
					"cluster state", c.Status.State)
				r.EventRecorder.Eventf(c, models.Warning, models.PatchFailed,
					"Cluster resource patch is failed. Reason: %v",
					err)

				return ctrl.Result{}, err
			}

			l.Info(msgDeleteClusterWithTwoFactorDelete, "cluster ID", c.Status.ID)

			r.EventRecorder.Event(c, models.Normal, models.DeletionStarted,
				"Two-Factor Delete is enabled, please confirm cluster deletion via email or phone.")
			return ctrl.Result{}, nil
		}
	}

	l.Info("Cadence cluster is being deleted",
		"cluster name", c.Spec.Name,
		"cluster status", c.Status)

	for _, packagedProvisioning := range c.Spec.PackagedProvisioning {
		err = r.deletePackagedResources(ctx, c, packagedProvisioning)
		if err != nil {
			l.Error(
				err, "Cannot delete Cadence packaged resources",
				"cluster name", c.Spec.Name,
				"cluster ID", c.Status.ID,
			)

			r.EventRecorder.Eventf(c, models.Warning, models.DeletionFailed,
				"Cannot delete Cadence packaged resources. Reason: %v", err)

			return ctrl.Result{}, err
		}
	}

	r.Scheduler.RemoveJob(c.GetJobID(scheduler.SyncJob))
	patch := c.NewPatch()
	controllerutil.RemoveFinalizer(c, models.DeletionFinalizer)
	c.Annotations[models.ResourceStateAnnotation] = models.DeletedEvent

	err = r.Patch(ctx, c, patch)
	if err != nil {
		l.Error(err, "Cannot patch Cadence cluster",
			"cluster name", c.Spec.Name,
			"patch", patch,
		)
		return ctrl.Result{}, err
	}

	err = exposeservice.Delete(r.Client, c.Name, c.Namespace)
	if err != nil {
		l.Error(err, "Cannot delete Cadence cluster expose service",
			"cluster ID", c.Status.ID,
			"cluster name", c.Spec.Name,
		)

		return ctrl.Result{}, err
	}

	l.Info("Cadence cluster was deleted",
		"cluster name", c.Spec.Name,
		"cluster ID", c.Status.ID,
	)

	r.EventRecorder.Event(c, models.Normal, models.Deleted, "Cluster resource is deleted")

	return ctrl.Result{}, nil
}

func (r *CadenceReconciler) preparePackagedSolution(
	ctx context.Context,
	c *v1beta1.Cadence,
	packagedProvisioning *v1beta1.PackagedProvisioning,
) (bool, error) {
	if len(c.Spec.DataCentres) < 1 {
		return false, models.ErrZeroDataCentres
	}

	labelsToQuery := fmt.Sprintf("%s=%s", models.ControlledByLabel, c.Name)
	selector, err := labels.Parse(labelsToQuery)
	if err != nil {
		return false, err
	}

	cassandraList := &v1beta1.CassandraList{}
	err = r.Client.List(ctx, cassandraList, &client.ListOptions{LabelSelector: selector})
	if err != nil {
		return false, err
	}

	if len(cassandraList.Items) == 0 {
		appVersions, err := r.API.ListAppVersions(models.CassandraAppKind)
		if err != nil {
			return false, fmt.Errorf("cannot list versions for kind: %v, err: %w",
				models.CassandraAppKind, err)
		}

		cassandraVersions := getSortedAppVersions(appVersions, models.CassandraAppType)
		if len(cassandraVersions) == 0 {
			return false, fmt.Errorf("there are no versions for %v kind",
				models.CassandraAppKind)
		}

		cassandraSpec, err := r.newCassandraSpec(c, cassandraVersions[len(cassandraVersions)-1].String())
		if err != nil {
			return false, err
		}

		err = r.Client.Create(ctx, cassandraSpec)
		if err != nil {
			return false, err
		}

		return true, nil
	}

	kafkaList := &v1beta1.KafkaList{}
	osList := &v1beta1.OpenSearchList{}
	advancedVisibility := &v1beta1.AdvancedVisibility{
		TargetKafka:      &v1beta1.CadenceDependencyTarget{},
		TargetOpenSearch: &v1beta1.CadenceDependencyTarget{},
	}
	var advancedVisibilities []*v1beta1.AdvancedVisibility
	if packagedProvisioning.UseAdvancedVisibility {
		err = r.Client.List(ctx, kafkaList, &client.ListOptions{LabelSelector: selector})
		if err != nil {
			return false, err
		}
		if len(kafkaList.Items) == 0 {
			appVersions, err := r.API.ListAppVersions(models.KafkaAppKind)
			if err != nil {
				return false, fmt.Errorf("cannot list versions for kind: %v, err: %w",
					models.KafkaAppKind, err)
			}

			kafkaVersions := getSortedAppVersions(appVersions, models.KafkaAppType)
			if len(kafkaVersions) == 0 {
				return false, fmt.Errorf("there are no versions for %v kind",
					models.KafkaAppType)
			}

			kafkaSpec, err := r.newKafkaSpec(c, kafkaVersions[len(kafkaVersions)-1].String())
			if err != nil {
				return false, err
			}

			err = r.Client.Create(ctx, kafkaSpec)
			if err != nil {
				return false, err
			}

			return true, nil
		}

		if len(kafkaList.Items[0].Status.DataCentres) == 0 {
			return true, nil
		}

		advancedVisibility.TargetKafka.DependencyCDCID = kafkaList.Items[0].Status.DataCentres[0].ID
		advancedVisibility.TargetKafka.DependencyVPCType = models.VPCPeered

		err = r.Client.List(ctx, osList, &client.ListOptions{LabelSelector: selector})
		if err != nil {
			return false, err
		}
		if len(osList.Items) == 0 {
			appVersions, err := r.API.ListAppVersions(models.OpenSearchAppKind)
			if err != nil {
				return false, fmt.Errorf("cannot list versions for kind: %v, err: %w",
					models.OpenSearchAppKind, err)
			}

			openSearchVersions := getSortedAppVersions(appVersions, models.OpenSearchAppType)
			if len(openSearchVersions) == 0 {
				return false, fmt.Errorf("there are no versions for %v kind",
					models.OpenSearchAppType)
			}

			// For OpenSearch we cannot use the latest version because is not supported by Cadence. So we use the oldest one.
			osSpec, err := r.newOpenSearchSpec(c, openSearchVersions[0].String())
			if err != nil {
				return false, err
			}

			err = r.Client.Create(ctx, osSpec)
			if err != nil {
				return false, err
			}

			return true, nil
		}

		if len(osList.Items[0].Status.DataCentres) == 0 {
			return true, nil
		}

		advancedVisibility.TargetOpenSearch.DependencyCDCID = osList.Items[0].Status.DataCentres[0].ID
		advancedVisibility.TargetOpenSearch.DependencyVPCType = models.VPCPeered
		advancedVisibilities = append(advancedVisibilities, advancedVisibility)
	}

	if len(cassandraList.Items[0].Status.DataCentres) == 0 {
		return true, nil
	}

	c.Spec.StandardProvisioning = append(c.Spec.StandardProvisioning, &v1beta1.StandardProvisioning{
		AdvancedVisibility: advancedVisibilities,
		TargetCassandra: &v1beta1.CadenceDependencyTarget{
			DependencyCDCID:   cassandraList.Items[0].Status.DataCentres[0].ID,
			DependencyVPCType: models.VPCPeered,
		},
	})

	return false, nil
}

func (r *CadenceReconciler) newCassandraSpec(c *v1beta1.Cadence, latestCassandraVersion string) (*v1beta1.Cassandra, error) {
	typeMeta := v1.TypeMeta{
		Kind:       models.CassandraKind,
		APIVersion: models.ClustersV1beta1APIVersion,
	}

	metadata := v1.ObjectMeta{
		Name:        models.CassandraChildPrefix + c.Name,
		Labels:      map[string]string{models.ControlledByLabel: c.Name},
		Annotations: map[string]string{models.ResourceStateAnnotation: models.CreatingEvent},
		Namespace:   c.ObjectMeta.Namespace,
		Finalizers:  []string{},
	}

	slaTier := c.Spec.SLATier
	privateNetwork := c.Spec.PrivateNetwork
	pciCompliance := c.Spec.PCICompliance

	var twoFactorDelete []*v1beta1.TwoFactorDelete
	if len(c.Spec.TwoFactorDelete) > 0 {
		twoFactorDelete = []*v1beta1.TwoFactorDelete{
			{
				Email: c.Spec.TwoFactorDelete[0].Email,
				Phone: c.Spec.TwoFactorDelete[0].Phone,
			},
		}
	}
	var cassNodeSize, network string
	var cassNodesNumber, cassReplicationFactor int
	var cassPrivateIPBroadcastForDiscovery, cassPasswordAndUserAuth bool
	for _, pp := range c.Spec.PackagedProvisioning {
		cassNodeSize = pp.BundledCassandraSpec.NodeSize
		network = pp.BundledCassandraSpec.Network
		cassNodesNumber = pp.BundledCassandraSpec.NodesNumber
		cassReplicationFactor = pp.BundledCassandraSpec.ReplicationFactor
		cassPrivateIPBroadcastForDiscovery = pp.BundledCassandraSpec.PrivateIPBroadcastForDiscovery
		cassPasswordAndUserAuth = pp.BundledCassandraSpec.PasswordAndUserAuth
	}

	dcName := models.CassandraChildDCName
	dcRegion := c.Spec.DataCentres[0].Region
	cloudProvider := c.Spec.DataCentres[0].CloudProvider
	providerAccountName := c.Spec.DataCentres[0].ProviderAccountName

	cassandraDataCentres := []*v1beta1.CassandraDataCentre{
		{
			GenericDataCentreSpec: v1beta1.GenericDataCentreSpec{
				Name:                dcName,
				Region:              dcRegion,
				CloudProvider:       cloudProvider,
				ProviderAccountName: providerAccountName,
				Network:             network,
			},
			NodeSize:                       cassNodeSize,
			NodesNumber:                    cassNodesNumber,
			ReplicationFactor:              cassReplicationFactor,
			PrivateIPBroadcastForDiscovery: cassPrivateIPBroadcastForDiscovery,
		},
	}

	spec := v1beta1.CassandraSpec{
		GenericClusterSpec: v1beta1.GenericClusterSpec{
			Name:            models.CassandraChildPrefix + c.Name,
			Version:         latestCassandraVersion,
			SLATier:         slaTier,
			PrivateNetwork:  privateNetwork,
			TwoFactorDelete: twoFactorDelete,
		},
		DataCentres:         cassandraDataCentres,
		PasswordAndUserAuth: cassPasswordAndUserAuth,
		PCICompliance:       pciCompliance,
		BundledUseOnly:      true,
	}

	return &v1beta1.Cassandra{
		TypeMeta:   typeMeta,
		ObjectMeta: metadata,
		Spec:       spec,
	}, nil
}

func (r *CadenceReconciler) startSyncJob(c *v1beta1.Cadence) error {
	job := r.newSyncJob(c)

	err := r.Scheduler.ScheduleJob(c.GetJobID(scheduler.SyncJob), scheduler.ClusterStatusInterval, job)
	if err != nil {
		return err
	}

	return nil
}

func (r *CadenceReconciler) newSyncJob(c *v1beta1.Cadence) scheduler.Job {
	l := log.Log.WithValues("syncJob", c.GetJobID(scheduler.SyncJob), "clusterID", c.Status.ID)
	return func() error {
		namespacedName := client.ObjectKeyFromObject(c)
		err := r.Get(context.Background(), namespacedName, c)
		if k8serrors.IsNotFound(err) {
			l.Info("Resource is not found in the k8s cluster. Closing Instaclustr status sync.",
				"namespaced name", namespacedName)
			r.Scheduler.RemoveJob(c.GetJobID(scheduler.SyncJob))
			return nil
		}
		if err != nil {
			l.Error(err, "Cannot get Cadence custom resource",
				"resource name", c.Name,
			)
			return err
		}

		instaModel, err := r.API.GetCadence(c.Status.ID)
		if err != nil {
			if errors.Is(err, instaclustr.NotFound) {
				if c.DeletionTimestamp != nil {
					_, err = r.handleDeleteCluster(context.Background(), c, l)
					return err
				}

				return r.handleExternalDelete(context.Background(), c)
			}

			l.Error(err, "Cannot get Cadence cluster from the Instaclustr API",
				"clusterID", c.Status.ID,
			)
			return err
		}

		iCadence := &v1beta1.Cadence{}
		iCadence.FromInstAPI(instaModel)

		if !c.Status.Equals(&iCadence.Status) {
			l.Info("Updating Cadence cluster status")

			areDCsEqual := c.Status.DCsEqual(iCadence.Status.DataCentres)

			patch := c.NewPatch()
			c.Status.FromInstAPI(instaModel)
			err = r.Status().Patch(context.Background(), c, patch)
			if err != nil {
				l.Error(err, "Cannot patch Cadence cluster",
					"cluster name", c.Spec.Name,
					"status", c.Status.State,
				)
				return err
			}

			if !areDCsEqual {
				var nodes []*v1beta1.Node

				for _, dc := range iCadence.Status.DataCentres {
					nodes = append(nodes, dc.Nodes...)
				}

				err = exposeservice.Create(r.Client,
					c.Name,
					c.Namespace,
					c.Spec.PrivateNetwork,
					nodes,
					models.CadenceConnectionPort)
				if err != nil {
					return err
				}
			}
		}

		if c.Annotations[models.ExternalChangesAnnotation] == models.True && c.IsSpecEqual(iCadence.Spec) {
			err = reconcileExternalChanges(r.Client, r.EventRecorder, c)
			if err != nil {
				return err
			}
		} else if c.Status.CurrentClusterOperationStatus == models.NoOperation &&
			c.Annotations[models.ResourceStateAnnotation] != models.UpdatingEvent &&
			!c.IsSpecEqual(iCadence.Spec) {
			l.Info(msgExternalChanges,
				"instaclustr data", iCadence.Spec.DataCentres,
				"k8s resource spec", c.Spec.DataCentres)

			patch := c.NewPatch()
			c.Annotations[models.ExternalChangesAnnotation] = models.True

			err = r.Patch(context.Background(), c, patch)
			if err != nil {
				l.Error(err, "Cannot patch cluster cluster",
					"cluster name", c.Spec.Name, "cluster state", c.Status.State)
				return err
			}

			msgDiffSpecs, err := createSpecDifferenceMessage(c.Spec, iCadence.Spec)
			if err != nil {
				l.Error(err, "Cannot create specification difference message",
					"instaclustr data", iCadence.Spec, "k8s resource spec", c.Spec)
				return err
			}
			r.EventRecorder.Eventf(c, models.Warning, models.ExternalChanges, msgDiffSpecs)
		}

		//TODO: change all context.Background() and context.TODO() to ctx from Reconcile
		err = r.reconcileMaintenanceEvents(context.Background(), c)
		if err != nil {
			l.Error(err, "Cannot reconcile cluster maintenance events",
				"cluster name", c.Spec.Name,
				"cluster ID", c.Status.ID,
			)
			return err
		}

		if c.Status.State == models.RunningStatus && c.Status.CurrentClusterOperationStatus == models.OperationInProgress {
			patch := c.NewPatch()
			for _, dc := range c.Status.DataCentres {
				resizeOperations, err := r.API.GetResizeOperationsByClusterDataCentreID(dc.ID)
				if err != nil {
					l.Error(err, "Cannot get data centre resize operations",
						"cluster name", c.Spec.Name,
						"cluster ID", c.Status.ID,
						"data centre ID", dc.ID,
					)

					return err
				}

				dc.ResizeOperations = resizeOperations
				err = r.Status().Patch(context.Background(), c, patch)
				if err != nil {
					l.Error(err, "Cannot patch data centre resize operations",
						"cluster name", c.Spec.Name,
						"cluster ID", c.Status.ID,
						"data centre ID", dc.ID,
					)

					return err
				}
			}
		}

		return nil
	}
}

func (r *CadenceReconciler) newKafkaSpec(c *v1beta1.Cadence, latestKafkaVersion string) (*v1beta1.Kafka, error) {
	typeMeta := v1.TypeMeta{
		Kind:       models.KafkaKind,
		APIVersion: models.ClustersV1beta1APIVersion,
	}

	metadata := v1.ObjectMeta{
		Name:        models.KafkaChildPrefix + c.Name,
		Labels:      map[string]string{models.ControlledByLabel: c.Name},
		Annotations: map[string]string{models.ResourceStateAnnotation: models.CreatingEvent},
		Namespace:   c.ObjectMeta.Namespace,
		Finalizers:  []string{},
	}

	if len(c.Spec.DataCentres) == 0 {
		return nil, models.ErrZeroDataCentres
	}

	var kafkaTFD []*v1beta1.TwoFactorDelete
	for _, cadenceTFD := range c.Spec.TwoFactorDelete {
		twoFactorDelete := &v1beta1.TwoFactorDelete{
			Email: cadenceTFD.Email,
			Phone: cadenceTFD.Phone,
		}
		kafkaTFD = append(kafkaTFD, twoFactorDelete)
	}
	bundledKafkaSpec := c.Spec.PackagedProvisioning[0].BundledKafkaSpec

	kafkaNetwork := bundledKafkaSpec.Network
	kafkaNodeSize := bundledKafkaSpec.NodeSize
	kafkaNodesNumber := bundledKafkaSpec.NodesNumber
	dcName := models.KafkaChildDCName
	dcRegion := c.Spec.DataCentres[0].Region
	cloudProvider := c.Spec.DataCentres[0].CloudProvider
	providerAccountName := c.Spec.DataCentres[0].ProviderAccountName
	kafkaDataCentres := []*v1beta1.KafkaDataCentre{
		{
			GenericDataCentreSpec: v1beta1.GenericDataCentreSpec{
				Name:                dcName,
				Region:              dcRegion,
				CloudProvider:       cloudProvider,
				ProviderAccountName: providerAccountName,
				Network:             kafkaNetwork,
			},
			NodeSize:    kafkaNodeSize,
			NodesNumber: kafkaNodesNumber,
		},
	}

	slaTier := c.Spec.SLATier
	privateClusterNetwork := c.Spec.PrivateNetwork
	pciCompliance := c.Spec.PCICompliance
	clientEncryption := c.Spec.DataCentres[0].ClientEncryption
	spec := v1beta1.KafkaSpec{
		GenericClusterSpec: v1beta1.GenericClusterSpec{
			Name:            models.KafkaChildPrefix + c.Name,
			Version:         latestKafkaVersion,
			SLATier:         slaTier,
			PrivateNetwork:  privateClusterNetwork,
			TwoFactorDelete: kafkaTFD,
		},
		DataCentres:               kafkaDataCentres,
		ReplicationFactor:         bundledKafkaSpec.ReplicationFactor,
		PartitionsNumber:          bundledKafkaSpec.PartitionsNumber,
		AllowDeleteTopics:         true,
		AutoCreateTopics:          true,
		ClientToClusterEncryption: clientEncryption,
		BundledUseOnly:            true,
		PCICompliance:             pciCompliance,
	}

	return &v1beta1.Kafka{
		TypeMeta:   typeMeta,
		ObjectMeta: metadata,
		Spec:       spec,
	}, nil
}

func (r *CadenceReconciler) newOpenSearchSpec(c *v1beta1.Cadence, oldestOpenSearchVersion string) (*v1beta1.OpenSearch, error) {
	typeMeta := v1.TypeMeta{
		Kind:       models.OpenSearchKind,
		APIVersion: models.ClustersV1beta1APIVersion,
	}

	metadata := v1.ObjectMeta{
		Name:        models.OpenSearchChildPrefix + c.Name,
		Labels:      map[string]string{models.ControlledByLabel: c.Name},
		Annotations: map[string]string{models.ResourceStateAnnotation: models.CreatingEvent},
		Namespace:   c.ObjectMeta.Namespace,
		Finalizers:  []string{},
	}

	if len(c.Spec.DataCentres) < 1 {
		return nil, models.ErrZeroDataCentres
	}

	bundledOpenSearchSpec := c.Spec.PackagedProvisioning[0].BundledOpenSearchSpec

	managerNodes := []*v1beta1.ClusterManagerNodes{{
		NodeSize:         bundledOpenSearchSpec.NodeSize,
		DedicatedManager: false,
	}}

	oNumberOfRacks := bundledOpenSearchSpec.NumberOfRacks
	slaTier := c.Spec.SLATier
	privateClusterNetwork := c.Spec.PrivateNetwork
	pciCompliance := c.Spec.PCICompliance

	var twoFactorDelete []*v1beta1.TwoFactorDelete
	if len(c.Spec.TwoFactorDelete) > 0 {
		twoFactorDelete = []*v1beta1.TwoFactorDelete{
			{
				Email: c.Spec.TwoFactorDelete[0].Email,
				Phone: c.Spec.TwoFactorDelete[0].Phone,
			},
		}
	}

	osNetwork := bundledOpenSearchSpec.Network
	dcName := models.OpenSearchChildDCName
	dcRegion := c.Spec.DataCentres[0].Region
	cloudProvider := c.Spec.DataCentres[0].CloudProvider
	providerAccountName := c.Spec.DataCentres[0].ProviderAccountName

	osDataCentres := []*v1beta1.OpenSearchDataCentre{
		{
			GenericDataCentreSpec: v1beta1.GenericDataCentreSpec{
				Name:                dcName,
				Region:              dcRegion,
				CloudProvider:       cloudProvider,
				ProviderAccountName: providerAccountName,
				Network:             osNetwork,
			},
			NumberOfRacks: oNumberOfRacks,
		},
	}
	spec := v1beta1.OpenSearchSpec{
		GenericClusterSpec: v1beta1.GenericClusterSpec{
			Name:            models.OpenSearchChildPrefix + c.Name,
			Version:         oldestOpenSearchVersion,
			SLATier:         slaTier,
			PrivateNetwork:  privateClusterNetwork,
			TwoFactorDelete: twoFactorDelete,
		},
		DataCentres:         osDataCentres,
		ClusterManagerNodes: managerNodes,
		BundledUseOnly:      true,
		PCICompliance:       pciCompliance,
	}

	return &v1beta1.OpenSearch{
		TypeMeta:   typeMeta,
		ObjectMeta: metadata,
		Spec:       spec,
	}, nil
}

func (r *CadenceReconciler) deletePackagedResources(
	ctx context.Context,
	c *v1beta1.Cadence,
	packagedProvisioning *v1beta1.PackagedProvisioning,
) error {
	labelsToQuery := fmt.Sprintf("%s=%s", models.ControlledByLabel, c.Name)
	selector, err := labels.Parse(labelsToQuery)
	if err != nil {
		return err
	}

	cassandraList := &v1beta1.CassandraList{}
	err = r.Client.List(ctx, cassandraList, &client.ListOptions{LabelSelector: selector})
	if err != nil {
		return err
	}

	if len(cassandraList.Items) != 0 {
		for _, cassandraCluster := range cassandraList.Items {
			err = r.Client.Delete(ctx, &cassandraCluster)
			if err != nil {
				return err
			}
		}
	}

	if packagedProvisioning.UseAdvancedVisibility {
		kafkaList := &v1beta1.KafkaList{}
		err = r.Client.List(ctx, kafkaList, &client.ListOptions{LabelSelector: selector})
		if err != nil {
			return err
		}
		if len(kafkaList.Items) != 0 {
			for _, kafkaCluster := range kafkaList.Items {
				err = r.Client.Delete(ctx, &kafkaCluster)
				if err != nil {
					return err
				}
			}
		}

		osList := &v1beta1.OpenSearchList{}
		err = r.Client.List(ctx, osList, &client.ListOptions{LabelSelector: selector})
		if err != nil {
			return err
		}
		if len(osList.Items) != 0 {
			for _, osCluster := range osList.Items {
				err = r.Client.Delete(ctx, &osCluster)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CadenceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{RateLimiter: r.RateLimiter}).
		For(&v1beta1.Cadence{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				if deleting := confirmDeletion(event.Object); deleting {
					return true
				}

				event.Object.GetAnnotations()[models.ResourceStateAnnotation] = models.CreatingEvent

				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				if event.ObjectNew.GetAnnotations()[models.ResourceStateAnnotation] == models.DeletedEvent {
					return false
				}
				if deleting := confirmDeletion(event.ObjectNew); deleting {
					return true
				}

				oldObj := event.ObjectOld.(*v1beta1.Cadence)
				newObj := event.ObjectNew.(*v1beta1.Cadence)

				if newObj.Status.ID == "" {
					newObj.Annotations[models.ResourceStateAnnotation] = models.CreatingEvent
					return true
				}

				if oldObj.Generation == newObj.Generation {
					return false
				}

				newObj.Annotations[models.ResourceStateAnnotation] = models.UpdatingEvent
				event.ObjectNew.SetAnnotations(newObj.Annotations)

				return true
			},
			GenericFunc: func(genericEvent event.GenericEvent) bool {
				annotations := genericEvent.Object.GetAnnotations()
				annotations[models.ResourceStateAnnotation] = models.GenericEvent
				genericEvent.Object.SetAnnotations(annotations)
				return true
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				return false
			},
		})).
		Complete(r)
}

func (r *CadenceReconciler) reconcileMaintenanceEvents(ctx context.Context, c *v1beta1.Cadence) error {
	l := log.FromContext(ctx)

	iMEStatuses, err := r.API.FetchMaintenanceEventStatuses(c.Status.ID)
	if err != nil {
		return err
	}

	if !c.Status.MaintenanceEventsEqual(iMEStatuses) {
		patch := c.NewPatch()
		c.Status.MaintenanceEvents = iMEStatuses
		err = r.Status().Patch(ctx, c, patch)
		if err != nil {
			return err
		}

		l.Info("Cluster maintenance events were reconciled",
			"cluster ID", c.Status.ID,
			"events", c.Status.MaintenanceEvents,
		)
	}

	return nil
}

func (r *CadenceReconciler) handleExternalDelete(ctx context.Context, c *v1beta1.Cadence) error {
	l := log.FromContext(ctx)

	patch := c.NewPatch()
	c.Status.State = models.DeletedStatus
	err := r.Status().Patch(ctx, c, patch)
	if err != nil {
		return err
	}

	l.Info(instaclustr.MsgInstaclustrResourceNotFound)
	r.EventRecorder.Eventf(c, models.Warning, models.ExternalDeleted, instaclustr.MsgInstaclustrResourceNotFound)

	r.Scheduler.RemoveJob(c.GetJobID(scheduler.BackupsChecker))
	r.Scheduler.RemoveJob(c.GetJobID(scheduler.SyncJob))

	return nil
}
