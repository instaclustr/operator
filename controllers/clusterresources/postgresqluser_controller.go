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

package clusterresources

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	k8sCore "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	"github.com/instaclustr/operator/pkg/exposeservice"
	"github.com/instaclustr/operator/pkg/models"
)

// PostgreSQLUserReconciler reconciles a PostgreSQLUser object
type PostgreSQLUserReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=postgresqlusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=postgresqlusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=clusterresources.instaclustr.com,resources=postgresqlusers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *PostgreSQLUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	u := &clusterresourcesv1beta1.PostgreSQLUser{}
	err := r.Get(ctx, req.NamespacedName, u)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			l.Info("PostgreSQL user resource is not found", "request", req)

			return models.ExitReconcile, nil
		}
		l.Error(err, "Cannot fetch PostgreSQL user resource", "request", req)

		return models.ReconcileRequeue, nil
	}

	s := &k8sCore.Secret{}
	err = r.Get(ctx, types.NamespacedName{
		Namespace: u.Spec.SecretRef.Namespace,
		Name:      u.Spec.SecretRef.Name,
	}, s)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			l.Info("PostgreSQL user secret is not found", "request", req)
			r.EventRecorder.Event(u, models.Warning, models.NotFound,
				"Secret is not found, please create a new secret or set an actual reference")
			return models.ReconcileRequeue, nil
		}

		l.Error(err, "Cannot get PostgreSQL user secret", "user", u.Name)

		return models.ReconcileRequeue, nil
	}

	newUsername, newPassword, err := getUserCreds(s)
	if err != nil {
		l.Error(err, "Cannot get the PostgreSQL user credentials from the secret",
			"secret name", s.Name,
			"secret namespace", s.Namespace)
		r.EventRecorder.Eventf(u, models.Warning, models.CreatingEvent,
			"Cannot get the PostgreSQL user credentials from the secret. Reason: %v", err)

		return models.ReconcileRequeue, nil
	}

	if controllerutil.AddFinalizer(s, u.GetDeletionFinalizer()) {
		err = r.Update(ctx, s)
		if err != nil {
			l.Error(err, "Cannot update PostgreSQL user's secret with deletion finalizer",
				"secret name", s.Name, "secret namespace", s.Namespace)
			r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
				"Update secret with deletion finalizer has been failed. Reason: %v", err)
			return models.ReconcileRequeue, nil
		}
	}

	patch := u.NewPatch()
	if controllerutil.AddFinalizer(u, u.GetDeletionFinalizer()) {
		err = r.Patch(ctx, u, patch)
		if err != nil {
			l.Error(err, "Cannot patch PostgreSQL user with deletion finalizer")
			r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
				"Patching PostgreSQL user with deletion finalizer has been failed. Reason: %v", err)
			return models.ReconcileRequeue, nil
		}
	}

	for clusterID, clusterInfo := range u.Status.ClustersInfo {
		if clusterInfo.Event == models.CreatingEvent {
			l.Info("Creating user", "user", u, "cluster ID", clusterID)

			err = r.createUser(ctx, clusterID, newUsername, newPassword, clusterInfo.DefaultSecretNamespacedName)
			if err != nil {
				if errors.Is(err, models.ErrExposeServiceNotCreatedYet) ||
					errors.Is(err, models.ErrExposeServiceEndpointsNotCreatedYet) {
					l.Info("Expose service or expose service endpoints are not created yet",
						"username", newUsername)

					return models.ReconcileRequeue, nil
				}

				l.Error(err, "Cannot create a user for the PostgreSQL cluster",
					"cluster ID", clusterID,
					"username", newUsername)
				r.EventRecorder.Eventf(u, models.Warning, models.CreatingEvent,
					"Cannot create user. Reason: %v", err)

				return models.ReconcileRequeue, nil
			}

			u.Status.ClustersInfo[clusterID] = clusterresourcesv1beta1.ClusterInfo{
				DefaultSecretNamespacedName: clusterInfo.DefaultSecretNamespacedName,
				Event:                       models.Created,
			}

			err = r.Status().Patch(ctx, u, patch)
			if err != nil {
				l.Error(err, "Cannot patch PostgreSQL user status",
					"cluster ID", clusterID,
					"username", newUsername)
				r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
					"Resource patch is failed. Reason: %v", err)

				return models.ReconcileRequeue, nil
			}

			l.Info("User has been created", "username", newUsername)

			r.EventRecorder.Eventf(u, models.Normal, models.Created,
				"User has been created for a cluster. Cluster ID: %s, username: %s",
				clusterID, newUsername)

			// TODO: Add deletion user finalizers

			continue
		}

		// TODO: implement user deletion logic on this event
	}

	// TODO: add logic for Deletion case

	return models.ExitReconcile, nil
}

func (r *PostgreSQLUserReconciler) createUser(
	ctx context.Context,
	clusterID string,
	newUserName string,
	newPassword string,
	defaultUserSecretNamespacedName clusterresourcesv1beta1.NamespacedName,
) error {
	defaultUserSecret := &k8sCore.Secret{}

	namespacedName := types.NamespacedName{
		Namespace: defaultUserSecretNamespacedName.Namespace,
		Name:      defaultUserSecretNamespacedName.Name,
	}
	err := r.Get(ctx, namespacedName, defaultUserSecret)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return fmt.Errorf("cannot get default PostgreSQL user secret, user reference: %v, err: %w", defaultUserSecretNamespacedName, err)
		}

		return err
	}

	defaultUsername, defaultPassword, err := getUserCreds(defaultUserSecret)
	if err != nil {
		return fmt.Errorf("cannot get default PostgreSQL user credentials, user reference: %v, err: %w", defaultUserSecretNamespacedName, err)
	}

	clusterName := defaultUserSecret.Labels[models.ControlledByLabel]
	exposeServiceList, err := exposeservice.GetExposeService(r.Client, clusterName, defaultUserSecretNamespacedName.Namespace)
	if err != nil {
		return fmt.Errorf("cannot list expose services for cluster: %s, err: %w", clusterID, err)
	}

	if len(exposeServiceList.Items) == 0 {
		return models.ErrExposeServiceNotCreatedYet
	}

	exposeServiceEndpoints, err := exposeservice.GetExposeServiceEndpoints(r.Client, clusterName, defaultUserSecretNamespacedName.Namespace)
	if err != nil {
		return fmt.Errorf("cannot list expose service endpoints for cluster: %s, err: %w", clusterID, err)
	}

	if len(exposeServiceEndpoints.Items) == 0 {
		return models.ErrExposeServiceEndpointsNotCreatedYet
	}

	nodeList := &k8sCore.NodeList{}

	err = r.List(ctx, nodeList, &client.ListOptions{})
	if err != nil {
		return fmt.Errorf("cannot list nodes, err: %w", err)
	}

	// TODO: Handle scenario if there are no nodes with external IP

	for _, node := range nodeList.Items {
		for _, nodeAddress := range node.Status.Addresses {
			if nodeAddress.Type == k8sCore.NodeExternalIP {
				err := r.createPostgreSQLFirewallRule(ctx, node.Name, clusterID, defaultUserSecretNamespacedName.Namespace, nodeAddress.Address)
				if err != nil {
					return fmt.Errorf("cannot create postgreSQL firewall rule, err: %w", err)
				}
			}
		}
	}

	serviceName := fmt.Sprintf(models.ExposeServiceNameTemplate, clusterName)
	host := fmt.Sprintf("%s.%s", serviceName, exposeServiceList.Items[0].Namespace)

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?target_session_attrs=read-write",
		defaultUsername, defaultPassword, host, models.DefaultPgDbPortValue, models.DefaultPgDbNameValue)

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("cannot establish a connection with a PostgreSQL server, err: %w", err)
	}
	defer conn.Close(ctx)

	createUserQuery := fmt.Sprintf(`CREATE USER "%s" WITH PASSWORD '%s'`, newUserName, newPassword)
	_, err = conn.Exec(ctx, createUserQuery)
	if err != nil {
		return fmt.Errorf("cannot execute creation user query in postgresql, err: %w", err)
	}

	return nil
}

func (r *PostgreSQLUserReconciler) createPostgreSQLFirewallRule(
	ctx context.Context,
	nodeName string,
	clusterID string,
	ns string,
	nodeAddress string,
) error {
	firewallRuleName := fmt.Sprintf("%s-%s-%s", models.ClusterNetworkFirewallRulePrefix, clusterID, nodeName)

	exists, err := r.firewallRuleExists(ctx, firewallRuleName, ns)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	firewallRule := &clusterresourcesv1beta1.ClusterNetworkFirewallRule{
		TypeMeta: ctrl.TypeMeta{
			Kind:       models.ClusterNetworkFirewallRuleKind,
			APIVersion: models.ClusterresourcesV1beta1APIVersion,
		},
		ObjectMeta: ctrl.ObjectMeta{
			Name:        firewallRuleName,
			Namespace:   ns,
			Annotations: map[string]string{models.ResourceStateAnnotation: models.CreatingEvent},
			Labels:      map[string]string{models.ClusterIDLabel: clusterID},
			Finalizers:  []string{models.DeletionFinalizer},
		},
		Spec: clusterresourcesv1beta1.ClusterNetworkFirewallRuleSpec{
			FirewallRuleSpec: clusterresourcesv1beta1.FirewallRuleSpec{
				ClusterID: clusterID,
				Type:      models.PgAppType,
			},
			Network: fmt.Sprintf("%s/%s", nodeAddress, "32"),
		},
		Status: clusterresourcesv1beta1.ClusterNetworkFirewallRuleStatus{},
	}

	err = r.Create(ctx, firewallRule)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgreSQLUserReconciler) firewallRuleExists(ctx context.Context, firewallRuleName string, ns string) (bool, error) {
	clusterNetworkFirewallRule := &clusterresourcesv1beta1.ClusterNetworkFirewallRule{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: ns,
		Name:      firewallRuleName,
	}, clusterNetworkFirewallRule)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgreSQLUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterresourcesv1beta1.PostgreSQLUser{}).
		Complete(r)
}
