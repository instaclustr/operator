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
	"github.com/instaclustr/operator/pkg/instaclustr"
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

			continue
		}

		if clusterInfo.Event == models.DeletingEvent {
			l.Info("Deleting user from a cluster", "cluster ID", clusterID)

			err = r.deleteUser(ctx, newUsername, clusterInfo.DefaultSecretNamespacedName)
			if err != nil {
				l.Error(err, "Cannot delete PostgreSQL user")
				r.EventRecorder.Eventf(u, models.Warning, models.DeletingEvent,
					"Cannot delete user. Reason: %v", err)

				return models.ReconcileRequeue, nil
			}

			l.Info("User has been deleted for cluster", "username", newUsername,
				"cluster ID", clusterID)
			r.EventRecorder.Eventf(u, models.Normal, models.Deleted,
				"User has been deleted for a cluster. Cluster ID: %s, username: %s",
				clusterID, newUsername)

			delete(u.Status.ClustersInfo, clusterID)

			err = r.Status().Patch(ctx, u, patch)
			if err != nil {
				l.Error(err, "Cannot patch PostgreSQL user status")
				r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
					"Resource patch is failed. Reason: %v", err)

				return models.ReconcileRequeue, nil
			}

			continue
		}

		if clusterInfo.Event == models.ClusterDeletingEvent {
			delete(u.Status.ClustersInfo, clusterID)

			err = r.Status().Patch(ctx, u, patch)
			if err != nil {
				l.Error(err, "Cannot detach clusterID from PostgreSQL user resource",
					"cluster ID", clusterID)
				r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
					"Detaching clusterID from the PostgreSQL user resource has been failed. Reason: %v", err)
				return models.ReconcileRequeue, nil
			}

			l.Info("PostgreSQL user has been detached from the cluster", "cluster ID", clusterID)
			r.EventRecorder.Eventf(u, models.Normal, models.Deleted,
				"User has been detached from the cluster. ClusterID: %v", clusterID)
		}
	}

	if u.DeletionTimestamp != nil {
		if u.Status.ClustersInfo != nil {
			l.Error(models.ErrUserStillExist, instaclustr.MsgDeleteUser)
			r.EventRecorder.Event(u, models.Warning, models.DeletingEvent, instaclustr.MsgDeleteUser)

			return models.ExitReconcile, nil
		}

		controllerutil.RemoveFinalizer(s, u.GetDeletionFinalizer())
		err = r.Update(ctx, s)
		if err != nil {
			l.Error(err, "Cannot delete finalizer from the user's secret")
			r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
				"Deleting finalizer from the user's secret has been failed. Reason: %v", err)
			return models.ReconcileRequeue, nil
		}

		controllerutil.RemoveFinalizer(u, u.GetDeletionFinalizer())
		err = r.Patch(ctx, u, patch)
		if err != nil {
			l.Error(err, "Cannot delete finalizer from the PostgreSQL user resource")
			r.EventRecorder.Eventf(u, models.Warning, models.PatchFailed,
				"Deleting finalizer from the PostgreSQL user resource has been failed. Reason: %v", err)
			return models.ReconcileRequeue, nil
		}

		l.Info("PostgreSQL user resource has been deleted")

		return models.ExitReconcile, nil
	}

	return models.ExitReconcile, nil
}

func (r *PostgreSQLUserReconciler) createUser(
	ctx context.Context,
	clusterID string,
	newUserName string,
	newPassword string,
	defaultUserSecretNamespacedName clusterresourcesv1beta1.NamespacedName,
) error {
	defaultCreds, clusterName, err := r.getDefaultPostgreSQLUserCreds(ctx, defaultUserSecretNamespacedName)
	if err != nil {
		return err
	}

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

	// TODO: Handle scenario if there are no nodes with external IP, check private/public cluster
	for _, node := range nodeList.Items {
		for _, nodeAddress := range node.Status.Addresses {
			if nodeAddress.Type == k8sCore.NodeExternalIP {
				err := r.createPostgreSQLFirewallRule(ctx, node.Name, clusterID, defaultUserSecretNamespacedName.Namespace, nodeAddress.Address)
				if err != nil {
					return fmt.Errorf("cannot create PostgreSQL firewall rule, err: %w", err)
				}
			}
		}
	}

	createUserQuery := fmt.Sprintf(`CREATE USER "%s" WITH PASSWORD '%s'`, newUserName, newPassword)
	err = r.ExecPostgreSQLQuery(ctx, createUserQuery, defaultCreds, clusterName, defaultUserSecretNamespacedName)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgreSQLUserReconciler) deleteUser(
	ctx context.Context,
	newUserName string,
	defaultUserSecretNamespacedName clusterresourcesv1beta1.NamespacedName,
) error {
	defaultCreds, clusterName, err := r.getDefaultPostgreSQLUserCreds(ctx, defaultUserSecretNamespacedName)
	if err != nil {
		return err
	}

	deleteUserQuery := fmt.Sprintf(`DROP USER IF EXISTS "%s"`, newUserName)
	err = r.ExecPostgreSQLQuery(ctx, deleteUserQuery, defaultCreds, clusterName, defaultUserSecretNamespacedName)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgreSQLUserReconciler) ExecPostgreSQLQuery(
	ctx context.Context,
	query string,
	defaultCreds *models.Credentials,
	clusterName string,
	defaultUserSecretNamespacedName clusterresourcesv1beta1.NamespacedName,
) error {
	serviceName := fmt.Sprintf(models.ExposeServiceNameTemplate, clusterName)
	host := fmt.Sprintf("%s.%s", serviceName, defaultUserSecretNamespacedName.Namespace)

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?target_session_attrs=read-write",
		defaultCreds.Username, defaultCreds.Password, host, models.DefaultPgDbPortValue, models.DefaultPgDbNameValue)

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("cannot establish a connection with a PostgreSQL server, err: %w", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("cannot execute query in PostgreSQL, err: %w", err)
	}

	return nil
}

func (r *PostgreSQLUserReconciler) getDefaultPostgreSQLUserCreds(
	ctx context.Context,
	defaultUserSecretNamespacedName clusterresourcesv1beta1.NamespacedName,
) (*models.Credentials, string, error) {
	defaultUserSecret := &k8sCore.Secret{}

	namespacedName := types.NamespacedName{
		Namespace: defaultUserSecretNamespacedName.Namespace,
		Name:      defaultUserSecretNamespacedName.Name,
	}
	err := r.Get(ctx, namespacedName, defaultUserSecret)
	if err != nil {
		return nil, "", fmt.Errorf("cannot get default PostgreSQL user secret, user reference: %v, err: %w", defaultUserSecretNamespacedName, err)
	}

	defaultUsername, defaultPassword, err := getUserCreds(defaultUserSecret)
	if err != nil {
		return nil, "", fmt.Errorf("cannot get default PostgreSQL user credentials, user reference: %v, err: %w", defaultUserSecretNamespacedName, err)
	}

	clusterName := defaultUserSecret.Labels[models.ControlledByLabel]

	defaultPostgreSQLUserCreds := &models.Credentials{
		Username: defaultUsername,
		Password: defaultPassword,
	}

	return defaultPostgreSQLUserCreds, clusterName, nil
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
