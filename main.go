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

package main

import (
	"context"
	"flag"
	"os"
	"time"

	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	virtcorev1 "kubevirt.io/api/core/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	clusterresourcesv1beta1 "github.com/instaclustr/operator/apis/clusterresources/v1beta1"
	clustersv1beta1 "github.com/instaclustr/operator/apis/clusters/v1beta1"
	kafkamanagementv1beta1 "github.com/instaclustr/operator/apis/kafkamanagement/v1beta1"
	clusterresourcescontrollers "github.com/instaclustr/operator/controllers/clusterresources"
	clusterscontrollers "github.com/instaclustr/operator/controllers/clusters"
	kafkamanagementcontrollers "github.com/instaclustr/operator/controllers/kafkamanagement"
	"github.com/instaclustr/operator/pkg/instaclustr"
	"github.com/instaclustr/operator/pkg/ratelimiter"
	"github.com/instaclustr/operator/pkg/scheduler"
	"github.com/instaclustr/operator/pkg/upgradecheck"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(clustersv1beta1.AddToScheme(scheme))
	utilruntime.Must(clusterresourcesv1beta1.AddToScheme(scheme))
	utilruntime.Must(kafkamanagementv1beta1.AddToScheme(scheme))
	utilruntime.Must(cdiv1beta1.AddToScheme(scheme))
	utilruntime.Must(virtcorev1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.DurationVar(&scheduler.ClusterStatusInterval, "cluster-status-interval", 60*time.Second,
		"An interval to check cluster status")
	flag.DurationVar(&scheduler.ClusterBackupsInterval, "cluster-backups-interval", 60*time.Second,
		"An interval to check cluster backups")
	flag.DurationVar(&scheduler.UserCreationInterval, "user-creation-interval", 60*time.Second,
		"An interval to try to create user during cluster creation")
	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.RFC3339TimeEncoder,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "680bba91.instaclustr.com",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	username := os.Getenv("USERNAME")
	key := os.Getenv("APIKEY")
	serverHostname := os.Getenv("HOSTNAME")

	instaClient := instaclustr.NewClient(
		username,
		key,
		serverHostname,
		instaclustr.DefaultTimeout,
	)

	s := scheduler.NewScheduler(log.Log.WithValues("component", "scheduler"))

	// TODO: take this variable from helm env
	autoUpgradeEnabled := false
	if autoUpgradeEnabled {
		setupLog.Info("auto upgrade operator is enabled")

		err = upgradecheck.StartUpgradeCheckJob(context.TODO(), mgr.GetClient(), s)
		if err != nil {
			setupLog.Error(err, "unable to start operator upgrade check job")
		}
	}

	eventRecorder := mgr.GetEventRecorderFor("instaclustr-operator")

	if err = (&clusterscontrollers.CassandraReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		Scheduler:     s,
		EventRecorder: eventRecorder,
		RateLimiter:   ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Cassandra")
		os.Exit(1)
	}
	if err = (&clusterscontrollers.PostgreSQLReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		Scheduler:     s,
		EventRecorder: eventRecorder,
		RateLimiter:   ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PostgreSQL")
		os.Exit(1)
	}
	if err = (&clusterscontrollers.OpenSearchReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		Scheduler:     s,
		EventRecorder: eventRecorder,
		RateLimiter:   ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "OpenSearch")
		os.Exit(1)
	}
	if err = (&clusterscontrollers.RedisReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		Scheduler:     s,
		EventRecorder: eventRecorder,
		RateLimiter:   ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Redis")
		os.Exit(1)
	}
	if err = (&clusterscontrollers.CadenceReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		Scheduler:     s,
		EventRecorder: eventRecorder,
		RateLimiter:   ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Cadence")
		os.Exit(1)
	}
	if err = (&clusterscontrollers.KafkaReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		Scheduler:     s,
		EventRecorder: eventRecorder,
		RateLimiter:   ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Kafka")
		os.Exit(1)
	}
	if err = (&clusterresourcescontrollers.AWSVPCPeeringReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		Scheduler:     s,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AWSVPCPeering")
		os.Exit(1)
	}
	if err = (&clusterresourcescontrollers.AzureVNetPeeringReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		Scheduler:     s,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AzureVNetPeering")
		os.Exit(1)
	}
	if err = (&clusterresourcescontrollers.GCPVPCPeeringReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		Scheduler:     s,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GCPVPCPeering")
		os.Exit(1)
	}
	if err = (&clusterscontrollers.KafkaConnectReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		Scheduler:     s,
		EventRecorder: eventRecorder,
		RateLimiter:   ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KafkaConnect")
		os.Exit(1)
	}
	if err = (&clusterresourcescontrollers.ClusterNetworkFirewallRuleReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		Scheduler:     s,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterNetworkFirewallRule")
		os.Exit(1)
	}
	if err = (&clusterresourcescontrollers.AWSSecurityGroupFirewallRuleReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		Scheduler:     s,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AWSSecurityGroupFirewallRule")
		os.Exit(1)
	}
	if err = (&clusterresourcescontrollers.ClusterBackupReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterBackup")
		os.Exit(1)
	}
	if err = (&kafkamanagementcontrollers.KafkaUserReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KafkaUser")
		os.Exit(1)
	}
	if err = (&clusterscontrollers.ZookeeperReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		Scheduler:     s,
		EventRecorder: eventRecorder,
		RateLimiter:   ratelimiter.NewItemExponentialFailureRateLimiterWithMaxTries(ratelimiter.DefaultBaseDelay, ratelimiter.DefaultMaxDelay),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Zookeeper")
		os.Exit(1)
	}
	if err = (&kafkamanagementcontrollers.TopicReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TopicName")
		os.Exit(1)
	}
	if err = (&kafkamanagementcontrollers.MirrorReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		Scheduler:     s,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Mirror")
		os.Exit(1)
	}
	if err = (&clusterresourcescontrollers.NodeReloadReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NodeReload")
		os.Exit(1)
	}
	if err = (&kafkamanagementcontrollers.KafkaACLReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KafkaACL")
		os.Exit(1)
	}
	if err = (&clusterresourcescontrollers.MaintenanceEventsReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		Scheduler:     s,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MaintenanceEvents")
		os.Exit(1)
	}
	if err = (&clustersv1beta1.Cassandra{}).SetupWebhookWithManager(mgr, instaClient); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Cassandra")
		os.Exit(1)
	}
	if err = (&clustersv1beta1.PostgreSQL{}).SetupWebhookWithManager(mgr, instaClient); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "PostgreSQL")
		os.Exit(1)
	}
	if err = (&clustersv1beta1.Redis{}).SetupWebhookWithManager(mgr, instaClient); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Redis")
		os.Exit(1)
	}
	if err = (&clusterresourcesv1beta1.AWSVPCPeering{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "AWSVPCPeering")
		os.Exit(1)
	}
	if err = (&clusterresourcesv1beta1.AWSSecurityGroupFirewallRule{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "AWSSecurityGroupFirewallRule")
		os.Exit(1)
	}
	if err = (&clustersv1beta1.OpenSearch{}).SetupWebhookWithManager(mgr, instaClient); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "OpenSearch")
		os.Exit(1)
	}
	if err = (&kafkamanagementv1beta1.KafkaACL{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "KafkaACL")
		os.Exit(1)
	}
	if err = (&clustersv1beta1.Kafka{}).SetupWebhookWithManager(mgr, instaClient); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Kafka")
		os.Exit(1)
	}
	if err = (&clusterresourcesv1beta1.ClusterNetworkFirewallRule{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ClusterNetworkFirewallRule")
		os.Exit(1)
	}
	if err = (&clustersv1beta1.KafkaConnect{}).SetupWebhookWithManager(mgr, instaClient); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "KafkaConnect")
		os.Exit(1)
	}
	if err = (&clusterresourcesv1beta1.MaintenanceEvents{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "MaintenanceEvents")
		os.Exit(1)
	}
	if err = (&clusterresourcesv1beta1.AzureVNetPeering{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "AzureVNetPeering")
		os.Exit(1)
	}
	if err = (&clusterresourcesv1beta1.GCPVPCPeering{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "GCPVPCPeering")
		os.Exit(1)
	}
	if err = (&clusterresourcesv1beta1.NodeReload{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "NodeReload")
		os.Exit(1)
	}
	if err = (&clusterresourcescontrollers.RedisUserReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RedisUser")
		os.Exit(1)
	}
	if err = (&clustersv1beta1.Zookeeper{}).SetupWebhookWithManager(mgr, instaClient); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Zookeeper")
		os.Exit(1)
	}
	if err = (&clustersv1beta1.Cadence{}).SetupWebhookWithManager(mgr, instaClient); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Cadence")
		os.Exit(1)
	}
	if err = (&clusterresourcescontrollers.AWSEncryptionKeyReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		Scheduler:     s,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AWSEncryptionKey")
		os.Exit(1)
	}
	if err = (&clusterresourcesv1beta1.AWSEncryptionKey{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "AWSEncryptionKey")
		os.Exit(1)
	}
	if err = (&clusterresourcescontrollers.CassandraUserReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CassandraUser")
		os.Exit(1)
	}
	if err = (&clusterresourcescontrollers.OpenSearchUserReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "OpenSearchUser")
		os.Exit(1)
	}
	if err = (&clusterresourcesv1beta1.OpenSearchUser{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "OpenSearchUser")
		os.Exit(1)
	}
	if err = (&clusterresourcesv1beta1.CassandraUser{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "CassandraUser")
		os.Exit(1)
	}
	if err = (&clusterresourcesv1beta1.RedisUser{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "RedisUser")
		os.Exit(1)
	}
	if err = (&kafkamanagementv1beta1.Mirror{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Mirror")
		os.Exit(1)
	}
	if err = (&kafkamanagementv1beta1.Topic{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Topic")
		os.Exit(1)
	}
	if err = (&kafkamanagementv1beta1.KafkaUser{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "KafkaUser")
		os.Exit(1)
	}
	if err = (&clusterresourcescontrollers.AWSEndpointServicePrincipalReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AWSEndpointServicePrincipal")
		os.Exit(1)
	}
	if err = (&clusterresourcesv1beta1.AWSEndpointServicePrincipal{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "AWSEndpointServicePrincipal")
		os.Exit(1)
	}
	if err = (&clusterresourcescontrollers.ExclusionWindowReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ExclusionWindow")
		os.Exit(1)
	}
	if err = (&clusterresourcesv1beta1.ExclusionWindow{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ExclusionWindow")
		os.Exit(1)
	}
	if err = (&clusterresourcesv1beta1.ClusterBackup{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ClusterBackup")
		os.Exit(1)
	}
	if err = (&clusterresourcescontrollers.OpenSearchEgressRulesReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "OpenSearchEgressRules")
		os.Exit(1)
	}
	if err = (&clusterresourcesv1beta1.OpenSearchEgressRules{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "OpenSearchEgressRules")
		os.Exit(1)
	}
	if err = (&kafkamanagementcontrollers.UserCertificateReconciler{
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		API:           instaClient,
		EventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "UserCertificate")
		os.Exit(1)
	}
	if err = (&kafkamanagementv1beta1.UserCertificate{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "UserCertificate")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
