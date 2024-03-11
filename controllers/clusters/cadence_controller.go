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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
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

func (r *CadenceReconciler) createCadence(ctx context.Context, c *v1beta1.Cadence, l logr.Logger) (*models.CadenceCluster, error) {
	l.Info(
		"Creating Cadence cluster",
		"cluster name", c.Spec.Name,
		"data centres", c.Spec.DataCentres,
	)

	cadenceAPISpec, err := c.Spec.ToInstAPI(ctx, r.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to convert k8s manifest to Instaclustr model, err: %w", err)
	}

	b, err := r.API.CreateClusterRaw(instaclustr.CadenceEndpoint, cadenceAPISpec)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster, err: %w", err)
	}

	instaModel := &models.CadenceCluster{}
	err = json.Unmarshal(b, instaModel)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal body to cadence model, err: %w", err)
	}

	r.EventRecorder.Eventf(c, models.Normal, models.Created,
		"Cluster creation request is sent. Cluster ID: %s", instaModel.ID,
	)

	return instaModel, nil
}

func (r *CadenceReconciler) createAWSArchivalSecret(ctx context.Context, c *v1beta1.Cadence, awsArchival []*models.AWSArchival) (*corev1.Secret, error) {
	secret := &corev1.Secret{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      fmt.Sprintf("%s-aws-archival", c.Name),
			Namespace: c.Namespace,
		},
	}

	err := controllerutil.SetOwnerReference(c, secret, r.Scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to set owner reference for aws-archival secret, err: %w", err)
	}

	secret.StringData = make(map[string]string)

	const awsSecretAccessKey = "awsSecretAccessKey"
	const awsAccessKeyID = "awsAccessKeyId"

	for _, archival := range awsArchival {
		secret.StringData[awsSecretAccessKey] = archival.AWSSecretAccessKey
		secret.StringData[awsAccessKeyID] = archival.AWSAccessKeyID
	}

	err = r.Create(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("secret creating failed, err: %w", err)
	}

	return secret, nil
}

func (r *CadenceReconciler) createCluster(ctx context.Context, c *v1beta1.Cadence, l logr.Logger) error {
	if !c.Spec.Inherits() {
		id, err := getClusterIDByName(r.API, models.CassandraAppType, c.Spec.Name)
		if err != nil {
			return err
		}

		if id != "" && c.Spec.Inherits() {
			l.Info("Cluster with provided name already exists", "name", c.Spec.Name, "clusterID", id)
			return fmt.Errorf("cluster %s already exists, please change name property", c.Spec.Name)
		}
	}

	var instaModel *models.CadenceCluster
	var err error

	switch {
	case c.Spec.Inherits():
		l.Info("Inheriting from the cluster", "clusterID", c.Spec.InheritsFrom)
		instaModel, err = r.API.GetCadence(c.Spec.InheritsFrom)
	default:
		instaModel, err = r.createCadence(ctx, c, l)
	}
	if err != nil {
		return err
	}

	patch := c.NewPatch()

	if c.Spec.Inherits() && len(instaModel.AWSArchival) > 0 {
		secret, err := r.createAWSArchivalSecret(ctx, c, instaModel.AWSArchival)
		if err != nil {
			return fmt.Errorf("failed to create aws-archival secret, err: %w", err)
		}

		c.Spec.AWSArchival = []*v1beta1.AWSArchival{{
			ArchivalS3URI:            instaModel.AWSArchival[0].ArchivalS3URI,
			ArchivalS3Region:         instaModel.AWSArchival[0].ArchivalS3Region,
			AccessKeySecretNamespace: secret.Namespace,
			AccessKeySecretName:      secret.Name,
		}}
	}

	c.Spec.FromInstAPI(instaModel)
	c.Annotations[models.ResourceStateAnnotation] = models.SyncingEvent
	err = r.Patch(ctx, c, patch)
	if err != nil {
		return fmt.Errorf("failed to patch resource spec, err: %w", err)
	}

	// TODO check on creation req
	//if c.Spec.Description != "" {
	//	err = r.API.UpdateDescriptionAndTwoFactorDelete(instaclustr.ClustersEndpointV1, id, c.Spec.Description, nil)
	//	if err != nil {
	//		l.Error(err, "Cannot update Cadence cluster description and TwoFactorDelete",
	//			"cluster name", c.Spec.Name,
	//			"description", c.Spec.Description,
	//			"twoFactorDelete", c.Spec.TwoFactorDelete,
	//		)
	//
	//		r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
	//			"Cluster description and TwoFactoDelete update is failed. Reason: %v", err)
	//	}
	//}

	c.Status.FromInstAPI(instaModel)
	err = r.Status().Patch(ctx, c, patch)
	if err != nil {
		return fmt.Errorf("failed to patch resource status, err: %w", err)
	}

	l.Info(
		"Cadence resource has been created",
		"cluster name", c.Name,
		"clusterID", c.Status.ID,
	)

	return nil
}

func (r *CadenceReconciler) handleCreateCluster(
	ctx context.Context,
	c *v1beta1.Cadence,
	l logr.Logger,
) (ctrl.Result, error) {
	if c.Status.ID == "" {
		for _, pp := range c.Spec.PackagedProvisioning {
			requeueNeeded, err := r.reconcilePackagedProvisioning(ctx, c, pp.UseAdvancedVisibility)
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

		err := r.createCluster(ctx, c, l)
		if err != nil {
			r.EventRecorder.Eventf(c, models.Warning, models.CreationFailed,
				"Failed to create cluster. Reason: %v", err,
			)
			return ctrl.Result{}, err
		}
	}

	if c.Status.State != models.DeletedStatus {
		patch := c.NewPatch()
		c.Annotations[models.ResourceStateAnnotation] = models.CreatedEvent
		controllerutil.AddFinalizer(c, models.DeletionFinalizer)
		err := r.Patch(ctx, c, patch)
		if err != nil {
			l.Error(err, "Cannot patch Cadence cluster",
				"cluster name", c.Spec.Name, "patch", patch)

			r.EventRecorder.Eventf(c, models.Warning, models.PatchFailed,
				"Cluster resource status patch is failed. Reason: %v", err)

			return ctrl.Result{}, err
		}

		err = r.startSyncJob(c)
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

func (r *CadenceReconciler) reconcilePackagedProvisioning(
	ctx context.Context,
	c *v1beta1.Cadence,
	useAdvancedVisibility bool,
) (bool, error) {
	for _, dc := range c.Spec.DataCentres {
		availableCIDRs, err := calculateAvailableNetworks(dc.Network)
		if err != nil {
			return false, err
		}

		labelsToQuery := fmt.Sprintf("%s=%s", models.ControlledByLabel, c.Name)
		selector, err := labels.Parse(labelsToQuery)
		if err != nil {
			return false, err
		}

		cassandraList := &v1beta1.CassandraList{}
		err = r.Client.List(ctx, cassandraList, &client.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			return false, err
		}

		if len(cassandraList.Items) == 0 {
			version, err := r.reconcileAppVersion(models.CassandraAppKind, models.CassandraAppType)
			if err != nil {
				return false, err
			}

			cassandraSpec, err := r.newCassandraSpec(c, version, availableCIDRs)
			if err != nil {
				return false, err
			}

			err = r.Client.Create(ctx, cassandraSpec)
			if err != nil {
				return false, err
			}

			return true, nil
		}

		var advancedVisibility []*v1beta1.AdvancedVisibility
		var clusterRefs []*v1beta1.Reference

		if useAdvancedVisibility {
			av := &v1beta1.AdvancedVisibility{
				TargetKafka:      &v1beta1.CadenceDependencyTarget{},
				TargetOpenSearch: &v1beta1.CadenceDependencyTarget{},
			}

			kafkaList := &v1beta1.KafkaList{}
			err = r.Client.List(ctx, kafkaList, &client.ListOptions{LabelSelector: selector})
			if err != nil {
				return false, err
			}

			if len(kafkaList.Items) == 0 {
				version, err := r.reconcileAppVersion(models.KafkaAppKind, models.KafkaAppType)
				if err != nil {
					return false, err
				}

				kafkaSpec, err := r.newKafkaSpec(c, version, availableCIDRs)
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

			av.TargetKafka.DependencyCDCID = kafkaList.Items[0].Status.DataCentres[0].ID
			av.TargetKafka.DependencyVPCType = models.VPCPeered

			ref := &v1beta1.Reference{
				Name:      kafkaList.Items[0].Name,
				Namespace: kafkaList.Items[0].Namespace,
			}
			clusterRefs = append(clusterRefs, ref)

			osList := &v1beta1.OpenSearchList{}
			err = r.Client.List(ctx, osList, &client.ListOptions{LabelSelector: selector})
			if err != nil {
				return false, err
			}

			if len(osList.Items) == 0 {
				version, err := r.reconcileAppVersion(models.OpenSearchAppKind, models.OpenSearchAppType)
				if err != nil {
					return false, err
				}

				osSpec, err := r.newOpenSearchSpec(c, version, availableCIDRs)
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

			ref = &v1beta1.Reference{
				Name:      osList.Items[0].Name,
				Namespace: osList.Items[0].Namespace,
			}
			clusterRefs = append(clusterRefs, ref)

			av.TargetOpenSearch.DependencyCDCID = osList.Items[0].Status.DataCentres[0].ID
			av.TargetOpenSearch.DependencyVPCType = models.VPCPeered
			advancedVisibility = append(advancedVisibility, av)
		}

		if len(cassandraList.Items[0].Status.DataCentres) == 0 {
			return true, nil
		}

		ref := &v1beta1.Reference{
			Name:      cassandraList.Items[0].Name,
			Namespace: cassandraList.Items[0].Namespace,
		}
		clusterRefs = append(clusterRefs, ref)
		c.Status.PackagedProvisioningClusterRefs = clusterRefs

		c.Spec.StandardProvisioning = append(c.Spec.StandardProvisioning, &v1beta1.StandardProvisioning{
			AdvancedVisibility: advancedVisibility,
			TargetCassandra: &v1beta1.CadenceDependencyTarget{
				DependencyCDCID:   cassandraList.Items[0].Status.DataCentres[0].ID,
				DependencyVPCType: models.VPCPeered,
			},
		})

	}

	return false, nil
}

func (r *CadenceReconciler) newCassandraSpec(
	c *v1beta1.Cadence,
	version string,
	availableCIDRs []string,
) (*v1beta1.Cassandra, error) {
	typeMeta := r.reconcileTypeMeta(models.CassandraAppKind)
	metadata := r.reconcileMetadata(c, models.CassandraAppKind)
	dcs := r.reconcileAppDCs(c, availableCIDRs, models.CassandraAppKind)

	cassandraDCs, ok := dcs.([]*v1beta1.CassandraDataCentre)
	if !ok {
		return nil, fmt.Errorf("resource is not of type %T, got %T", []*v1beta1.CassandraDataCentre(nil), cassandraDCs)
	}

	spec := r.reconcileAppSpec(c, models.CassandraAppKind, version)
	cassandraSpec, ok := spec.(v1beta1.CassandraSpec)
	if !ok {
		return nil, fmt.Errorf("resource is not of type %T, got %T", v1beta1.CassandraSpec{}, spec)
	}

	cassandraSpec.DataCentres = cassandraDCs

	return &v1beta1.Cassandra{
		TypeMeta:   typeMeta,
		ObjectMeta: metadata,
		Spec:       cassandraSpec,
	}, nil
}

func (r *CadenceReconciler) reconcileTypeMeta(app string) v1.TypeMeta {
	typeMeta := v1.TypeMeta{
		APIVersion: models.ClustersV1beta1APIVersion,
	}

	switch app {
	case models.CassandraAppKind:
		typeMeta.Kind = models.CassandraKind
	case models.KafkaAppKind:
		typeMeta.Kind = models.KafkaKind
	case models.OpenSearchAppKind:
		typeMeta.Kind = models.OpenSearchKind
	}

	return typeMeta
}

func (r *CadenceReconciler) reconcileMetadata(c *v1beta1.Cadence, app string) v1.ObjectMeta {
	metadata := v1.ObjectMeta{
		Labels:      map[string]string{models.ControlledByLabel: c.Name},
		Annotations: map[string]string{models.ResourceStateAnnotation: models.CreatingEvent},
		Namespace:   c.ObjectMeta.Namespace,
		Finalizers:  []string{},
	}

	switch app {
	case models.CassandraAppKind:
		metadata.Name = models.CassandraChildPrefix + c.Name
	case models.KafkaAppKind:
		metadata.Name = models.KafkaChildPrefix + c.Name
	case models.OpenSearchAppKind:
		metadata.Name = models.OpenSearchChildPrefix + c.Name
	}

	return metadata
}

func (r *CadenceReconciler) reconcileAppDCs(c *v1beta1.Cadence, availableCIDRs []string, app string) any {
	network := r.reconcileNetwork(availableCIDRs, app)

	genericDataCentreSpec := v1beta1.GenericDataCentreSpec{
		Region:              c.Spec.DataCentres[0].Region,
		CloudProvider:       c.Spec.DataCentres[0].CloudProvider,
		ProviderAccountName: c.Spec.DataCentres[0].ProviderAccountName,
		Network:             network,
	}

	switch app {
	case models.CassandraAppKind:
		var privateIPBroadcast bool
		if c.Spec.PrivateNetwork {
			privateIPBroadcast = true
		}

		genericDataCentreSpec.Name = models.CassandraChildDCName
		cassandraDCs := []*v1beta1.CassandraDataCentre{
			{
				GenericDataCentreSpec: genericDataCentreSpec,
				NodeSize: c.Spec.CalculateNodeSize(
					c.Spec.DataCentres[0].CloudProvider,
					c.Spec.PackagedProvisioning[0].SolutionSize,
					models.CassandraAppKind,
				),
				NodesNumber:                    3,
				ReplicationFactor:              3,
				PrivateIPBroadcastForDiscovery: privateIPBroadcast,
			},
		}

		return cassandraDCs

	case models.KafkaAppKind:
		genericDataCentreSpec.Name = models.KafkaChildDCName
		kafkaDCs := []*v1beta1.KafkaDataCentre{
			{
				GenericDataCentreSpec: genericDataCentreSpec,
				NodeSize: c.Spec.CalculateNodeSize(
					c.Spec.DataCentres[0].CloudProvider,
					c.Spec.PackagedProvisioning[0].SolutionSize,
					models.KafkaAppKind,
				),
				NodesNumber: 3,
			},
		}
		return kafkaDCs

	case models.OpenSearchAppKind:
		genericDataCentreSpec.Name = models.OpenSearchChildDCName
		openSearchDCs := []*v1beta1.OpenSearchDataCentre{
			{
				GenericDataCentreSpec: genericDataCentreSpec,
				NumberOfRacks:         3,
			},
		}
		return openSearchDCs
	}

	return nil
}

func (r *CadenceReconciler) reconcileAppSpec(c *v1beta1.Cadence, app, version string) any {
	var twoFactorDelete []*v1beta1.TwoFactorDelete
	for _, cadenceTFD := range c.Spec.TwoFactorDelete {
		tfd := &v1beta1.TwoFactorDelete{
			Email: cadenceTFD.Email,
			Phone: cadenceTFD.Phone,
		}
		twoFactorDelete = append(twoFactorDelete, tfd)
	}

	genericClusterSpec := v1beta1.GenericClusterSpec{
		Version:         version,
		SLATier:         c.Spec.SLATier,
		PrivateNetwork:  c.Spec.PrivateNetwork,
		TwoFactorDelete: twoFactorDelete,
	}

	switch app {
	case models.CassandraAppKind:
		genericClusterSpec.Name = models.CassandraChildPrefix + c.Name
		spec := v1beta1.CassandraSpec{
			GenericClusterSpec:  genericClusterSpec,
			PasswordAndUserAuth: true,
			PCICompliance:       c.Spec.PCICompliance,
			BundledUseOnly:      true,
		}
		return spec

	case models.KafkaAppKind:
		genericClusterSpec.Name = models.KafkaChildPrefix + c.Name
		spec := v1beta1.KafkaSpec{
			GenericClusterSpec:        genericClusterSpec,
			PCICompliance:             c.Spec.PCICompliance,
			BundledUseOnly:            true,
			ReplicationFactor:         3,
			PartitionsNumber:          3,
			AllowDeleteTopics:         true,
			AutoCreateTopics:          true,
			ClientToClusterEncryption: c.Spec.DataCentres[0].ClientEncryption,
		}

		return spec

	case models.OpenSearchAppKind:
		managerNodes := []*v1beta1.ClusterManagerNodes{{
			NodeSize: c.Spec.CalculateNodeSize(
				c.Spec.DataCentres[0].CloudProvider,
				c.Spec.PackagedProvisioning[0].SolutionSize,
				models.OpenSearchAppKind,
			),
			DedicatedManager: false,
		}}

		genericClusterSpec.Name = models.OpenSearchChildPrefix + c.Name
		spec := v1beta1.OpenSearchSpec{
			GenericClusterSpec:  genericClusterSpec,
			PCICompliance:       c.Spec.PCICompliance,
			BundledUseOnly:      true,
			ClusterManagerNodes: managerNodes,
		}

		return spec
	}

	return nil
}

func (r *CadenceReconciler) reconcileNetwork(availableCIDRs []string, app string) string {
	switch app {
	case models.CassandraAppKind:
		return availableCIDRs[1]
	case models.KafkaAppKind:
		return availableCIDRs[2]
	case models.OpenSearchAppKind:
		return availableCIDRs[3]
	}

	return ""
}

func (r *CadenceReconciler) reconcileAppVersion(appKind, appType string) (string, error) {
	appVersions, err := r.API.ListAppVersions(appKind)
	if err != nil {
		return "", err
	}

	sortedAppVersions := getSortedAppVersions(appVersions, appType)
	if len(sortedAppVersions) == 0 {
		return "", err
	}

	switch appType {
	case models.CassandraAppType, models.KafkaAppType:
		return sortedAppVersions[len(sortedAppVersions)-1].String(), nil

	case models.OpenSearchAppType:
		return sortedAppVersions[0].String(), nil
	}

	return "", nil
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

func (r *CadenceReconciler) newKafkaSpec(
	c *v1beta1.Cadence,
	version string,
	availableCIDRs []string,
) (*v1beta1.Kafka, error) {
	typeMeta := r.reconcileTypeMeta(models.KafkaAppKind)
	metadata := r.reconcileMetadata(c, models.KafkaAppKind)
	dcs := r.reconcileAppDCs(c, availableCIDRs, models.KafkaAppKind)

	kafkaDCs, ok := dcs.([]*v1beta1.KafkaDataCentre)
	if !ok {
		return nil, fmt.Errorf("resource is not of type %T, got %T", []*v1beta1.KafkaDataCentre(nil), kafkaDCs)
	}

	spec := r.reconcileAppSpec(c, models.KafkaAppKind, version)
	kafkaSpec, ok := spec.(v1beta1.KafkaSpec)
	if !ok {
		return nil, fmt.Errorf("resource is not of type %T, got %T", v1beta1.KafkaSpec{}, spec)
	}

	kafkaSpec.DataCentres = kafkaDCs

	return &v1beta1.Kafka{
		TypeMeta:   typeMeta,
		ObjectMeta: metadata,
		Spec:       kafkaSpec,
	}, nil
}

func (r *CadenceReconciler) newOpenSearchSpec(
	c *v1beta1.Cadence,
	version string,
	availableCIDRs []string,
) (*v1beta1.OpenSearch, error) {
	typeMeta := r.reconcileTypeMeta(models.OpenSearchAppKind)
	metadata := r.reconcileMetadata(c, models.OpenSearchAppKind)
	dcs := r.reconcileAppDCs(c, availableCIDRs, models.OpenSearchAppKind)

	osDCs, ok := dcs.([]*v1beta1.OpenSearchDataCentre)
	if !ok {
		return nil, fmt.Errorf("resource is not of type %T, got %T", []*v1beta1.OpenSearchDataCentre(nil), osDCs)
	}

	spec := r.reconcileAppSpec(c, models.OpenSearchAppKind, version)
	osSpec, ok := spec.(v1beta1.OpenSearchSpec)
	if !ok {
		return nil, fmt.Errorf("resource is not of type %T, got %T", v1beta1.OpenSearchSpec{}, spec)
	}

	osSpec.DataCentres = osDCs

	return &v1beta1.OpenSearch{
		TypeMeta:   typeMeta,
		ObjectMeta: metadata,
		Spec:       osSpec,
	}, nil
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
