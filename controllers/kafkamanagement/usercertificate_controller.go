/*
Copyright 2023.

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
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clustersv1beta1 "github.com/instaclustr/operator/apis/clusters/v1beta1"
	kafkamanagementv1beta1 "github.com/instaclustr/operator/apis/kafkamanagement/v1beta1"
	"github.com/instaclustr/operator/pkg/apiextensions"
	"github.com/instaclustr/operator/pkg/helpers/utils"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/ratelimiter"
)

// UserCertificateReconciler reconciles a CertificateSigningRequest object
type UserCertificateReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	API           instaclustr.API
	EventRecorder record.EventRecorder
}

//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=usercertificates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=usercertificates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kafkamanagement.instaclustr.com,resources=usercertificates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *UserCertificateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	cert := &kafkamanagementv1beta1.UserCertificate{}
	err := r.Get(ctx, req.NamespacedName, cert)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if cert.DeletionTimestamp != nil {
		return ctrl.Result{}, r.handleDelete(ctx, cert)
	}

	if cert.Status.CertID != "" {
		return ctrl.Result{}, nil
	}

	kafkaUser := &kafkamanagementv1beta1.KafkaUser{}
	err = r.Get(ctx, cert.Spec.UserRef.AsNamespacedName(), kafkaUser)
	if err != nil {
		r.EventRecorder.Eventf(cert, models.Warning, models.CreationFailed,
			"Failed to get KafkaUser resource. Reason: %v", err,
		)
		return ctrl.Result{}, err
	}

	kafkaCluster := &clustersv1beta1.Kafka{}
	err = r.Get(ctx, cert.Spec.ClusterRef.AsNamespacedName(), kafkaCluster)
	if err != nil {
		r.EventRecorder.Eventf(cert, models.Warning, models.CreationFailed,
			"Failed to get Kafka resource. Reason: %v", err,
		)
		return ctrl.Result{}, err
	}

	res, err := r.checkKafkaUserExistsOnCluster(kafkaCluster, kafkaUser)
	if err != nil {
		return res, err
	}

	if cert.Spec.CertificateRequestTemplate != nil && cert.Spec.SecretRef == nil {
		err = r.createCSR(ctx, cert, kafkaUser)
		if err != nil {
			r.EventRecorder.Eventf(cert, models.Warning, models.CreationFailed,
				"Certificate signing request creation failed. Reason: %v", err,
			)
			return ctrl.Result{}, err
		}
	}

	err = r.handleCreate(ctx, cert, kafkaCluster, kafkaUser)
	if err != nil {
		return ctrl.Result{}, err
	}

	l.Info("Certificate has been created",
		"certId", cert.Status.CertID,
		"userRef", client.ObjectKeyFromObject(kafkaUser),
		"clusterRef", client.ObjectKeyFromObject(kafkaCluster),
	)
	r.EventRecorder.Event(cert, models.Normal, models.Created, "Certificate has been created")

	return ctrl.Result{}, nil
}

func (r *UserCertificateReconciler) checkKafkaUserExistsOnCluster(
	cluster *clustersv1beta1.Kafka,
	user *kafkamanagementv1beta1.KafkaUser,
) (ctrl.Result, error) {
	var userRefExists bool
	for _, userRef := range cluster.Spec.UserRefs {
		if userRef.Name == user.Name && userRef.Namespace == user.Namespace {
			userRefExists = true
			break
		}
	}

	event, clusterEventExists := user.Status.ClustersEvents[cluster.Status.ID]

	if !userRefExists || !clusterEventExists {
		return ctrl.Result{}, fmt.Errorf("user %s/%s is not added to the cluster %s/%s",
			user.Namespace, user.Name, cluster.Namespace, cluster.Name)
	}

	if event == models.CreatingEvent {
		// Waiting until user is created
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	return ctrl.Result{}, nil
}

func (r *UserCertificateReconciler) handleCreate(
	ctx context.Context,
	cert *kafkamanagementv1beta1.UserCertificate,
	cluster *clustersv1beta1.Kafka,
	user *kafkamanagementv1beta1.KafkaUser,
) error {
	userSecret := &v1.Secret{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: user.Spec.SecretRef.Namespace,
		Name:      user.Spec.SecretRef.Name,
	}, userSecret)
	if err != nil {
		return err
	}

	csrSecret := &v1.Secret{}
	err = r.Get(ctx, cert.Spec.SecretRef.AsNamespacedName(), csrSecret)
	if err != nil {
		return err
	}

	username, _, err := utils.GetUserCreds(userSecret)
	if err != nil {
		return err
	}

	csrRaw := csrSecret.Data[cert.Spec.SecretRef.Key]

	err = r.validateCSR(csrRaw, username)
	if err != nil {
		return err
	}

	signedCert, err := r.API.CreateKafkaUserCertificate(&models.CertificateRequest{
		ClusterID:     cluster.Status.ID,
		CSR:           string(csrRaw),
		KafkaUsername: username,
		ValidPeriod:   cert.Spec.ValidPeriod,
	})
	if err != nil {
		return err
	}

	certSecret, err := r.createSignedCertificateSecret(ctx, cert, signedCert.SignedCertificate)
	if err != nil {
		return err
	}

	patch := cert.NewPatch()
	cert.Status = kafkamanagementv1beta1.UserCertificateStatus{
		CertID:     signedCert.ID,
		ExpiryDate: signedCert.ExpiryDate,
		SignedCertSecretRef: kafkamanagementv1beta1.Reference{
			Namespace: certSecret.Namespace,
			Name:      certSecret.Name,
		},
	}
	err = r.Status().Patch(ctx, cert, patch)
	if err != nil {
		return err
	}

	controllerutil.AddFinalizer(cert, models.DeletionFinalizer)
	err = r.Patch(ctx, cert, patch)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserCertificateReconciler) createSignedCertificateSecret(
	ctx context.Context,
	cert *kafkamanagementv1beta1.UserCertificate,
	signedCert string,
) (*v1.Secret, error) {
	secret := &v1.Secret{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      fmt.Sprintf("%s-signed-cert", cert.Name),
			Namespace: cert.Namespace,
		},
		StringData: map[string]string{
			models.SignedCertificateSecretKey: signedCert,
		},
	}

	_ = controllerutil.SetControllerReference(cert, secret, r.Scheme)

	err := r.Create(ctx, secret)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func (r *UserCertificateReconciler) createCSR(
	ctx context.Context,
	csr *kafkamanagementv1beta1.UserCertificate,
	user *kafkamanagementv1beta1.KafkaUser,
) error {
	userSecret := &v1.Secret{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: user.Spec.SecretRef.Namespace,
		Name:      user.Spec.SecretRef.Name,
	}, userSecret)
	if err != nil {
		return err
	}

	username, _, err := utils.GetUserCreds(userSecret)
	if err != nil {
		return err
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	template := x509.CertificateRequest{
		Subject:            csr.Spec.CertificateRequestTemplate.ToSubject(),
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	template.Subject.CommonName = username

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	err = pem.Encode(buf, &pem.Block{Type: models.CertificateRequestType, Bytes: csrBytes})
	if err != nil {
		return err
	}

	csrSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      csr.Name + "-csr",
			Namespace: csr.Namespace,
		},
		Data: map[string][]byte{
			models.CSRSecretKey:        buf.Bytes(),
			models.PrivateKeySecretKey: exportRSAPrivateKeyAsPEM(privateKey),
		},
	}

	_ = controllerutil.SetControllerReference(csr, csrSecret, r.Scheme)

	err = r.Create(ctx, csrSecret)
	if err != nil {
		return err
	}

	csr.Spec.SecretRef = &apiextensions.ObjectFieldReference{
		Name:      csrSecret.Name,
		Namespace: csrSecret.Namespace,
		Key:       models.CSRSecretKey,
	}

	return r.Update(ctx, csr)
}

func exportRSAPrivateKeyAsPEM(privateKey *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(
		&pem.Block{
			Type:  models.RSAPrivateKeyType,
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)
}

func (r *UserCertificateReconciler) validateCSR(raw []byte, username string) error {
	block, _ := pem.Decode(raw)

	req, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return err
	}

	if req.Subject.CommonName != username {
		return errors.New("common name in the secret doesn't match username")
	}

	return nil
}

func (r *UserCertificateReconciler) handleDelete(ctx context.Context, cert *kafkamanagementv1beta1.UserCertificate) error {
	l := log.FromContext(ctx)

	err := r.API.DeleteKafkaUserCertificate(cert.Status.CertID)
	if err != nil && !errors.Is(err, instaclustr.NotFound) {
		return err
	}

	l.Info("The certificate has been deleted",
		"certId", cert.Status.CertID,
	)

	patch := cert.NewPatch()
	controllerutil.RemoveFinalizer(cert, models.DeletionFinalizer)
	return r.Patch(ctx, cert, patch)
}

// SetupWithManager sets up the controller with the Manager.
func (r *UserCertificateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(
				ratelimiter.DefaultBaseDelay,
				ratelimiter.DefaultMaxDelay,
			),
		}).
		For(&kafkamanagementv1beta1.UserCertificate{}).
		Complete(r)
}
