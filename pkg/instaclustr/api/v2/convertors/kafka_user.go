package convertors

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/apis/kafkamanagement/v1alpha1"
	models2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
	"github.com/instaclustr/operator/pkg/models"
)

func KafkaUserToAPIv2(
	kafkaUserSpec *v1alpha1.KafkaUserSpec,
	ctx *context.Context,
	k8sClient client.Client,
) (*models2.KafkaUserAPIv2, error) {
	kafkaUser := &models2.KafkaUserAPIv2{
		ClusterID:          kafkaUserSpec.ClusterID,
		InitialPermissions: kafkaUserSpec.InitialPermissions,
		Options:            kafkaUserOptionsToIsntaKafkaUserOptions(kafkaUserSpec.Options),
	}

	err := getKafkaUserCredsFromSecret(
		kafkaUserSpec,
		kafkaUser,
		ctx,
		k8sClient)
	if err != nil {
		return nil, err
	}
	return kafkaUser, nil
}

func getKafkaUserCredsFromSecret(
	kafkaUserSpec *v1alpha1.KafkaUserSpec,
	kafkaUser *models2.KafkaUserAPIv2,
	ctx *context.Context,
	k8sClient client.Client,
) error {
	kafkaUserSecret := &v1.Secret{}
	kafkaUserSecretNamespacedName := types.NamespacedName{
		Name:      kafkaUserSpec.KafkaUserSecretName,
		Namespace: kafkaUserSpec.KafkaUserSecretNamespace,
	}
	err := k8sClient.Get(*ctx, kafkaUserSecretNamespacedName, kafkaUserSecret)
	if err != nil {
		return err
	}
	username := kafkaUserSecret.Data[models.Username]
	password := kafkaUserSecret.Data[models.Password]
	Username := string(username[:len(username)-1])
	Password := string(password[:len(password)-1])

	kafkaUser.Username = Username
	kafkaUser.Password = Password

	return nil
}

func kafkaUserOptionsToIsntaKafkaUserOptions(kafkaUserOptions *v1alpha1.Options) *models2.KafkaUserOptionsAPIv2 {
	instaKafkaUserOptions := &models2.KafkaUserOptionsAPIv2{
		OverrideExistingUser: kafkaUserOptions.OverrideExistingUser,
		SASLSCRAMMechanism:   kafkaUserOptions.SASLSCRAMMechanism,
	}
	return instaKafkaUserOptions

}
