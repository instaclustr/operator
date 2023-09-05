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

package kafkamanagement

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"strings"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/instaclustr/operator/apis/kafkamanagement/v1beta1"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
)

const (
	kafkaUserField = ".spec.kafkaUserSecretName"
)

// KafkaUserReconciler reconciles a KafkaUser object
type KafkaUserReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=kafkausers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=kafkausers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=kafkausers/finalizers,verbs=update
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=secrets,verbs=get;list;watch;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *KafkaUserReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	user := &v1beta1.KafkaUser{}
	err := r.Client.Get(ctx, req.NamespacedName, user)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("Kafka user resource is not found", "request", req)
			return models.ExitReconcile, nil
		}
		l.Error(err, "Unable to fetch Kafka user", "request", req)
		r.EventRecorder.Eventf(
			user, models.Warning, models.FetchFailed,
			"Fetch user is failed. Reason: %v",
			err,
		)

		return models.ReconcileRequeue, nil
	}

	secret := &v1.Secret{}
	kafkaUserSecretNamespacedName := types.NamespacedName{
		Name:      user.Spec.SecretRef.Name,
		Namespace: user.Spec.SecretRef.Namespace,
	}
	err = r.Client.Get(ctx, kafkaUserSecretNamespacedName, secret)
	if err != nil {
		l.Error(err, "Cannot get secret for kafka user.", "request", req)
		r.EventRecorder.Eventf(user, models.Warning, models.FetchFailed,
			"Fetch user credentials secret is failed. Reason: %v", err)

		return models.ReconcileRequeue, nil
	}

	username, password, err := r.getKafkaUserCredsFromSecret(user.Spec)
	if err != nil {
		l.Error(err, "Cannot get Kafka user creds from secret", "kafka user spec", user.Spec)
		r.EventRecorder.Eventf(user, models.Warning, models.FetchFailed,
			"Fetch user credentials from secret is failed. Reason: %v", err)

		return models.ReconcileRequeue, nil
	}

	if controllerutil.AddFinalizer(secret, user.GetDeletionFinalizer()) {
		err = r.Update(ctx, secret)
		if err != nil {
			l.Error(err, "Cannot update Kafka user secret",
				"secret name", secret.Name,
				"secret namespace", secret.Namespace)
			r.EventRecorder.Eventf(user, models.Warning, models.UpdatedEvent,
				"Cannot assign Kafka user to a k8s secret. Reason: %v", err)

			return models.ReconcileRequeue, nil
		}
	}

	patch := user.NewPatch()
	if controllerutil.AddFinalizer(user, user.GetDeletionFinalizer()) {
		err = r.Patch(ctx, user, patch)
		if err != nil {
			l.Error(err, "Cannot patch Kafka user with finalizer")
			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Patching Kafka user with finalizer has been failed. Reason: %v", err)

			return models.ReconcileRequeue, nil
		}
	}

	for clusterID, clusterEvent := range user.Status.ClustersEvents {
		if clusterEvent == models.CreatingEvent {
			l.Info(
				"Creating Kafka user",
				"initial permissions", user.Spec.InitialPermissions,
				"kafka user options", user.Spec.Options,
			)
			iKafkaUser := user.Spec.ToInstAPI(clusterID, username, password)
			_, err = r.API.CreateKafkaUser(instaclustr.KafkaUserEndpoint, iKafkaUser)
			if err != nil {
				l.Error(
					err, "Cannot create Kafka User",
					"kafka user resource spec", user.Spec,
				)
				r.EventRecorder.Eventf(
					user, models.Warning, models.CreationFailed,
					"Resource creation on the Instaclustr is failed. Reason: %v",
					err,
				)

				return models.ReconcileRequeue, nil
			}

			r.EventRecorder.Eventf(
				user, models.Normal, models.Created,
				"Resource creation request is sent. user ID: %s",
				user.GetID(clusterID, username),
			)

			annots := user.GetAnnotations()
			if annots == nil {
				user.SetAnnotations(make(map[string]string))
			}

			user.Status.ClustersEvents[clusterID] = models.CreatedEvent
			err = r.Status().Patch(ctx, user, patch)
			if err != nil {
				l.Error(err, "Cannot patch Kafka user resource status",
					"initial permissions", user.Spec.InitialPermissions,
					"kafka user options", user.Spec.Options,
					"kafka user metadata", user.ObjectMeta,
				)
				r.EventRecorder.Eventf(
					user, models.Warning, models.PatchFailed,
					"Resource status patch is failed. Reason: %v",
					err,
				)

				return models.ReconcileRequeue, nil
			}

			l.Info(
				"Kafka user was created",
				"initial permissions", user.Spec.InitialPermissions,
				"kafka user options", user.Spec.Options,
			)

			continue
		}

		if clusterEvent == models.DeletingEvent {
			userID := user.GetID(clusterID, username)
			err = r.API.DeleteKafkaUser(userID, instaclustr.KafkaUserEndpoint)
			if err != nil {
				l.Error(err, "cannot delete Kafka user",
					"initial permissions", user.Spec.InitialPermissions,
					"kafka user options", user.Spec.Options,
					"kafka user metadata", user.ObjectMeta,
				)
				r.EventRecorder.Eventf(
					user, models.Warning, models.DeletionFailed,
					"Resource deletion on the Instaclustr is failed. Reason: %v",
					err,
				)

				return models.ReconcileRequeue, nil
			}

			r.EventRecorder.Eventf(
				user, models.Normal, models.DeletionStarted,
				"Resource deletion request is sent to the Instaclustr API.",
			)

			patch = user.NewPatch()
			delete(user.Status.ClustersEvents, clusterID)
			err = r.Status().Patch(ctx, user, patch)
			if err != nil {
				l.Error(err, "Cannot patch Kafka user resource status",
					"initial permissions", user.Spec.InitialPermissions,
					"kafka user options", user.Spec.Options,
					"kafka user metadata", user.ObjectMeta,
				)
				r.EventRecorder.Eventf(
					user, models.Warning, models.PatchFailed,
					"Resource status patch is failed. Reason: %v",
					err,
				)

				return models.ReconcileRequeue, nil
			}

			l.Info("Kafka user has been deleted",
				"initial permissions", user.Spec.InitialPermissions,
				"kafka user options", user.Spec.Options,
			)

			r.EventRecorder.Eventf(user, models.Normal, models.Deleted,
				"user has been deleted for a cluster, username: %s, clusterID: %s.",
				username, clusterID)

			continue
		}

		if clusterEvent == models.ClusterDeletingEvent {
			delete(user.Status.ClustersEvents, clusterID)
			err = r.Status().Patch(ctx, user, patch)
			if err != nil {
				l.Error(err, "Cannot patch Kafka user resource status",
					"initial permissions", user.Spec.InitialPermissions,
					"kafka user options", user.Spec.Options,
					"kafka user metadata", user.ObjectMeta,
				)
				r.EventRecorder.Eventf(
					user, models.Warning, models.PatchFailed,
					"Resource status patch is failed. Reason: %v",
					err,
				)

				return models.ReconcileRequeue, nil
			}

			l.Info("Kafka user has been detached",
				"initial permissions", user.Spec.InitialPermissions,
				"kafka user options", user.Spec.Options,
			)
			r.EventRecorder.Eventf(
				user, models.Normal, models.Deleted,
				"User is detached from cluster",
			)

			continue

		}

		iKafkaUser := user.Spec.ToInstAPI(clusterID, username, password)
		userID := user.GetID(clusterID, username)
		err = r.API.UpdateKafkaUser(userID, iKafkaUser)
		if err != nil {
			l.Error(err, "Cannot update Kafka user",
				"initial permissions", user.Spec.InitialPermissions,
				"kafka user options", user.Spec.Options,
			)
			r.EventRecorder.Eventf(
				user, models.Warning, models.UpdateFailed,
				"Resource update on the Instaclustr API is failed. Reason: %v",
				err,
			)

			return models.ReconcileRequeue, nil
		}

		user.Annotations[models.ResourceStateAnnotation] = models.UpdatedEvent
		err = r.Patch(ctx, user, patch)
		if err != nil {
			l.Error(err, "Cannot patch Kafka user resource metadata",
				"initial permissions", user.Spec.InitialPermissions,
				"kafka user options", user.Spec.Options,
				"kafka user metadata", user.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v",
				err,
			)

			return models.ReconcileRequeue, nil
		}

		user.Status.ClustersEvents[clusterID] = models.UpdatedEvent
		err = r.Status().Patch(ctx, user, patch)
		if err != nil {
			l.Error(err, "Cannot patch Kafka user resource status",
				"initial permissions", user.Spec.InitialPermissions,
				"kafka user options", user.Spec.Options,
				"kafka user metadata", user.ObjectMeta,
			)
			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Resource status patch is failed. Reason: %v",
				err,
			)

			return models.ReconcileRequeue, nil
		}

		l.Info("Kafka user resource has been updated",
			"initial permissions", user.Spec.InitialPermissions,
			"kafka user options", user.Spec.Options,
		)

		continue
	}

	if user.DeletionTimestamp != nil {
		for clusterID, clusterEvent := range user.Status.ClustersEvents {
			if clusterEvent == models.Created || clusterEvent == models.CreatingEvent || clusterEvent == models.UpdatingEvent || clusterEvent == models.UpdatedEvent {
				l.Error(models.ErrUserStillExist, instaclustr.MsgDeleteUser,
					"username", username, "cluster ID", clusterID)
				r.EventRecorder.Event(user, models.Warning, models.DeletingEvent, instaclustr.MsgDeleteUser)

				return models.ReconcileRequeue, nil
			}
		}

		controllerutil.RemoveFinalizer(user, user.GetDeletionFinalizer())
		err = r.Patch(ctx, user, patch)
		if err != nil {
			l.Error(err, "Cannot delete finalizer from the Kafka user resource")
			r.EventRecorder.Eventf(
				user, models.Warning, models.PatchFailed,
				"Deleting finalizer from the Kafka user resource has been failed. Reason: %v", err,
			)

			return models.ReconcileRequeue, nil
		}

		controllerutil.RemoveFinalizer(secret, user.GetDeletionFinalizer())
		err = r.Update(ctx, secret)
		if err != nil {
			l.Error(err, "Cannot remove finalizer from secret", "secret name", secret.Name)

			r.EventRecorder.Eventf(user, models.Warning, models.PatchFailed,
				"Resource patch is failed. Reason: %v", err)

			return models.ReconcileRequeue, nil
		}

		l.Info("Kafka user resource has been deleted", "username", username)
	}

	return models.ExitReconcile, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KafkaUserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&v1beta1.KafkaUser{},
		kafkaUserField,
		func(rawObj client.Object,
		) []string {

			kafkaUser := rawObj.(*v1beta1.KafkaUser)
			if kafkaUser.Spec.SecretRef.Name == "" {
				return nil
			}

			return []string{kafkaUser.Spec.SecretRef.Name}
		}); err != nil {

		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.KafkaUser{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: func(event event.UpdateEvent) bool {
				newObj := event.ObjectNew.(*v1beta1.KafkaUser)
				oldObj := event.ObjectOld.(*v1beta1.KafkaUser)
				if newObj.Generation != event.ObjectOld.GetGeneration() {
					r.handleCertificateEvent(newObj, oldObj.Spec.CertificateRequests)
				}

				return true
			},
		})).Owns(&v1.Secret{}).
		Watches(
			&source.Kind{Type: &v1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.findSecretObjects),
		).
		Complete(r)
}

func (r *KafkaUserReconciler) findSecretObjects(secret client.Object) []reconcile.Request {
	kafkaUserList := &v1beta1.KafkaUserList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(kafkaUserField, secret.GetName()),
		Namespace:     secret.GetNamespace(),
	}
	err := r.List(context.TODO(), kafkaUserList, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, len(kafkaUserList.Items))
	for i, item := range kafkaUserList.Items {
		patch := item.NewPatch()
		item.GetAnnotations()[models.ResourceStateAnnotation] = models.SecretEvent
		err = r.Patch(context.TODO(), &item, patch)
		if err != nil {
			return []reconcile.Request{}
		}
		requests[i] = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      item.GetName(),
				Namespace: item.GetNamespace(),
			},
		}
	}
	return requests
}

func (r *KafkaUserReconciler) getKafkaUserCredsFromSecret(
	kafkaUserSpec v1beta1.KafkaUserSpec,
) (string, string, error) {
	kafkaUserSecret := &v1.Secret{}
	kafkaUserSecretNamespacedName := types.NamespacedName{
		Name:      kafkaUserSpec.SecretRef.Name,
		Namespace: kafkaUserSpec.SecretRef.Namespace,
	}

	err := r.Get(context.TODO(), kafkaUserSecretNamespacedName, kafkaUserSecret)
	if err != nil {
		return "", "", err
	}

	username := kafkaUserSecret.Data[models.Username]
	password := kafkaUserSecret.Data[models.Password]

	if len(username) == 0 || len(password) == 0 {
		return "", "", models.ErrMissingSecretKeys
	}

	return string(username[:len(username)-1]), string(password[:len(password)-1]), nil
}

func (r *KafkaUserReconciler) getKafkaUserCertIDFromSecret(
	ctx context.Context,
	certRequest *v1beta1.CertificateRequest,
) (string, error) {
	kafkaUserCertSecret := &v1.Secret{}
	kafkaUserCertSecretNamespacedName := types.NamespacedName{
		Name:      certRequest.SecretName,
		Namespace: certRequest.SecretNamespace,
	}

	err := r.Get(ctx, kafkaUserCertSecretNamespacedName, kafkaUserCertSecret)
	if err != nil {
		return "", err
	}

	certID := kafkaUserCertSecret.Data["id"]

	if len(certID) == 0 {
		return "", models.ErrMissingSecretKeys
	}

	return string(certID), nil
}

func (r *KafkaUserReconciler) GenerateCSR(certRequest *v1beta1.CertificateRequest) (string, error) {
	keyBytes, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}

	subj := pkix.Name{
		CommonName:         certRequest.CommonName,
		Country:            []string{certRequest.Country},
		Organization:       []string{certRequest.Organization},
		OrganizationalUnit: []string{certRequest.OrganizationalUnit},
	}

	template := x509.CertificateRequest{
		Subject:            subj,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, keyBytes)
	if err != nil {
		return "", err
	}
	strBuf := strings.Builder{}
	err = pem.Encode(&strBuf, &pem.Block{Type: "NEW CERTIFICATE REQUEST", Bytes: csrBytes})
	if err != nil {
		return "", err
	}

	return strBuf.String(), nil
}

func (r *KafkaUserReconciler) UpdateCertSecret(ctx context.Context, secret *v1.Secret, certResp *v1beta1.Certificate) error {
	secret.StringData = make(map[string]string)

	secret.StringData["id"] = certResp.ID
	secret.StringData["expiryDate"] = certResp.ExpiryDate
	secret.StringData["signedCertificate"] = certResp.SignedCertificate

	err := r.Update(ctx, secret)
	if err != nil {
		return err
	}

	return nil
}

func (r *KafkaUserReconciler) handleCertificateEvent(
	newObj *v1beta1.KafkaUser,
	oldCerts []*v1beta1.CertificateRequest,
) {
	ctx := context.TODO()
	l := log.FromContext(ctx)

	for _, oldCert := range oldCerts {
		var exist bool
		for _, newCert := range newObj.Spec.CertificateRequests {
			if *oldCert == *newCert {
				exist = true
				break
			}
		}

		if exist {
			continue
		}

		err := r.handleDeleteCertificate(ctx, newObj, l, oldCert)
		if err != nil {
			l.Error(err, "Cannot delete Kafka user mTLS certificate", "user", oldCert)
			r.EventRecorder.Eventf(newObj, models.Warning, models.CreatingEvent,
				"Cannot delete mTlS certificate. Reason: %v", err)
		}
	}

	for _, newCert := range newObj.Spec.CertificateRequests {
		var exist bool
		for _, oldCert := range oldCerts {
			if *newCert == *oldCert {
				exist = true
				break
			}
		}

		if exist {
			if newCert.AutoRenew {
				err := r.handleRenewCertificate(ctx, newObj, newCert, l)
				if err != nil {
					l.Error(err, "Cannot renew Kafka user mTLS certificate", "cert", newCert)
					r.EventRecorder.Eventf(newObj, models.Warning, models.CreatingEvent,
						"Cannot renew user mTLS certificate. Reason: %v", err)
				}
			}
			continue
		}

		err := r.handleCreateCertificate(ctx, newObj, l, newCert)
		if err != nil {
			l.Error(err, "Cannot create Kafka user mTLS certificate", "cert", newCert)
			r.EventRecorder.Eventf(newObj, models.Warning, models.CreatingEvent,
				"Cannot create user mTLS certificate. Reason: %v", err)
		}

		oldCerts = append(oldCerts, newCert)
	}
}

func (r *KafkaUserReconciler) handleCreateCertificate(ctx context.Context, user *v1beta1.KafkaUser, l logr.Logger, certRequest *v1beta1.CertificateRequest) error {
	username, _, err := r.getKafkaUserCredsFromSecret(user.Spec)
	if err != nil {
		l.Error(
			err, "Cannot get Kafka user creds from secret",
			"kafka user spec", user.Spec,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.FetchFailed,
			"Fetch user credentials from secret is failed. Reason: %v",
			err,
		)

		return err
	}

	var isCSRGenerated bool

	if certRequest.CSR == "" {
		certRequest.CSR, err = r.GenerateCSR(certRequest)
		if err != nil {
			l.Error(err, "Cannot generate CSR for Kafka user certificate creation",
				"user", user.Name,
			)
			r.EventRecorder.Eventf(
				user, models.Warning, models.GenerateFailed,
				"Generate CSR is failed. Reason: %v",
				err,
			)

			return err
		}
		isCSRGenerated = true
	}

	certResponse, err := r.API.CreateKafkaUserCertificate(certRequest.ToInstAPI(username))
	if err != nil {
		l.Error(err, "Cannot create Kafka user mTLS certificate",
			"user", user.Name,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.CreationFailed,
			"Certificate creation is failed. Reason: %v",
			err,
		)

		return err
	}

	newSecret := user.NewCertificateSecret(certRequest.SecretName, certRequest.SecretNamespace)
	err = r.Client.Create(ctx, newSecret)
	if err != nil {
		l.Error(err, "Cannot create Kafka user Cert Secret.",
			"secret name", certRequest.SecretName,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.CreationFailed,
			"create user Cert Secret is failed. Reason: %v",
			err,
		)

		return err

	}

	controllerutil.AddFinalizer(newSecret, models.DeletionFinalizer)

	err = r.UpdateCertSecret(ctx, newSecret, certResponse)
	if err != nil {
		l.Error(err, "Cannot update certificate secret",
			"user", user.Name,
			"secret", newSecret.Name,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.UpdateFailed,
			"Certificate secret update is failed. Reason: %v",
			err,
		)

		return err
	}

	l.Info("Kafka user mTLS certificate has been created",
		"User ID", user.GetID(certRequest.ClusterID, username),
	)

	if isCSRGenerated {
		certRequest.CSR = ""
	}

	return nil
}

func (r *KafkaUserReconciler) handleDeleteCertificate(ctx context.Context, user *v1beta1.KafkaUser, l logr.Logger, certRequest *v1beta1.CertificateRequest) error {
	certID, err := r.getKafkaUserCertIDFromSecret(ctx, certRequest)
	if err != nil {
		l.Error(
			err, "Cannot get Kafka user certificate ID from secret",
			"kafka user certificate secret name", certRequest.SecretName,
			"kafka user certificate secret namespace", certRequest.SecretNamespace,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.FetchFailed,
			"Fetch user certificate ID from secret is failed. Reason: %v",
			err,
		)

		return err
	}

	err = r.API.DeleteKafkaUserCertificate(certID)
	if err != nil {
		l.Error(err, "Cannot Delete Kafka user mTLS certificate",
			"user", user.Name,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.DeletionFailed,
			"Certificate deletion is failed. Reason: %v",
			err,
		)

		return err
	}

	secret := &v1.Secret{}
	certSecretNamespacedName := types.NamespacedName{
		Name:      certRequest.SecretName,
		Namespace: certRequest.SecretNamespace,
	}
	err = r.Client.Get(ctx, certSecretNamespacedName, secret)
	if err != nil {
		l.Error(err, "Cannot get Kafka user certificate secret.",
			"kafka user certificate secret name", certRequest.SecretName,
			"kafka user certificate secret namespace", certRequest.SecretNamespace,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.FetchFailed,
			"Fetch user certificate secret is failed. Reason: %v",
			err,
		)

		return err
	}

	controllerutil.RemoveFinalizer(secret, models.DeletionFinalizer)
	err = r.Update(ctx, secret)
	if err != nil {
		l.Error(err, "Cannot remove finalizer from secret", "secret name", secret.Name)

		r.EventRecorder.Eventf(user, models.Warning, models.PatchFailed,
			"Resource patch is failed. Reason: %v", err)

		return err
	}

	err = r.Client.Delete(ctx, secret)
	if err != nil && !k8serrors.IsNotFound(err) {
		l.Error(err, "Cannot delete Kafka user certificate secret",
			"kafka user certificate secret name", certRequest.SecretName,
			"kafka user certificate secret namespace", certRequest.SecretNamespace,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.DeletionFailed,
			"Delete user certificate secret is failed. Reason: %v",
			err,
		)

		return err
	}

	l.Info("Kafka user mTLS certificate has been deleted",
		"Certificate ID", certID,
	)

	return nil
}

func (r *KafkaUserReconciler) handleRenewCertificate(ctx context.Context, user *v1beta1.KafkaUser, certRequest *v1beta1.CertificateRequest, l logr.Logger) error {
	certID, err := r.getKafkaUserCertIDFromSecret(ctx, certRequest)
	if err != nil {
		l.Error(
			err, "Cannot get Kafka user certificate ID from secret",
			"kafka user certificate secret name", certRequest.SecretName,
			"kafka user certificate secret namespace", certRequest.SecretNamespace,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.FetchFailed,
			"Fetch user certificate ID from secret is failed. Reason: %v",
			err,
		)

		return err
	}

	newCert, err := r.API.RenewKafkaUserCertificate(certID)
	if err != nil {
		l.Error(err, "Cannot Renew Kafka user mTLS certificate",
			"user", user.Name,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.DeletionFailed,
			"Certificate renew is failed. Reason: %v",
			err,
		)

		return err
	}

	secret := &v1.Secret{}
	certSecretNamespacedName := types.NamespacedName{
		Name:      certRequest.SecretName,
		Namespace: certRequest.SecretNamespace,
	}
	err = r.Client.Get(ctx, certSecretNamespacedName, secret)
	if err != nil {
		l.Error(err, "Cannot get Kafka user certificate secret.",
			"kafka user certificate secret name", certRequest.SecretName,
			"kafka user certificate secret namespace", certRequest.SecretNamespace,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.FetchFailed,
			"Fetch user certificate secret is failed. Reason: %v",
			err,
		)

		return err
	}

	err = r.UpdateCertSecret(ctx, secret, newCert)
	if err != nil {
		l.Error(err, "Cannot update certificate secret",
			"user", user.Name,
			"secret", secret.Name,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.UpdateFailed,
			"Certificate secret update is failed. Reason: %v",
			err,
		)

		return err
	}

	l.Info("Kafka user mTLS certificate has been renewed",
		"Certificate ID", certID,
		"New expiry date", newCert.ExpiryDate,
		"New ID", newCert.ID,
	)

	err = r.API.DeleteKafkaUserCertificate(certID)
	if err != nil {
		l.Error(err, "Cannot Delete Kafka user mTLS certificate",
			"user", user.Name,
		)
		r.EventRecorder.Eventf(
			user, models.Warning, models.DeletionFailed,
			"Certificate deletion is failed. Reason: %v",
			err,
		)

		return err
	}

	return nil
}
