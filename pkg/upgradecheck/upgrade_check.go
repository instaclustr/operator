package upgradecheck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/scheduler"
)

func StartUpgradeCheckJob(ctx context.Context, client client.Client, s scheduler.Interface) error {
	job := newUpgradeCheckJob(ctx, client)
	return s.ScheduleJob(models.OperatorUpgradeChecker, scheduler.AutoUpgradeCheckInterval, job)
}

func newUpgradeCheckJob(ctx context.Context, client client.Client) scheduler.Job {
	l := log.FromContext(ctx, "components", "UpgradeCheckJob")

	return func() error {
		// TODO: change from dockerhub to custom endpoint
		latestTag, err := getLatestDockerImageTag("icoperator/instaclustr-operator")
		if err != nil {
			return fmt.Errorf("unable to get latest docker image tag: %w", err)
		}

		err = updateImageTagIfNeeded(ctx, l, client, latestTag)
		if err != nil {
			return fmt.Errorf("unable to update current image tag: %w", err)
		}

		return nil
	}
}

func getLatestDockerImageTag(imageName string) (string, error) {
	url := fmt.Sprintf("https://registry.hub.docker.com/v2/repositories/%s/tags?page_size=1", imageName)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tagsResponse models.DockerTagsResponse
	err = json.Unmarshal(body, &tagsResponse)
	if err != nil {
		return "", err
	}

	if len(tagsResponse.Results) == 0 {
		return "", fmt.Errorf("no tags found for image %s", imageName)
	}

	return tagsResponse.Results[0].Name, nil
}

func updateImageTagIfNeeded(ctx context.Context, l logr.Logger, mgrClient client.Client, latestTag string) error {
	instDeployment, err := findInstOperatorDeployment(ctx, mgrClient)
	if err != nil {
		return err
	}

	container, err := findContainer(instDeployment, models.InstOperatorContainerName)
	if err != nil {
		return err
	}

	currentImageTag, err := getCurrentImageTag(container.Image)
	if err != nil {
		return err
	}

	if currentImageTag != latestTag {
		// TODO: add rollback in error case and check health status before update
		if err := updateContainerImage(ctx, mgrClient, instDeployment, container, latestTag); err != nil {
			return fmt.Errorf("cannot update latest docker image: %w", err)
		}

		l.Info("Operator has been updated to the latest version", "old version", currentImageTag, "new version", latestTag)
	} else {
		l.Info("The operator is already up to date", "current version", currentImageTag)
	}

	return nil
}

func findInstOperatorDeployment(ctx context.Context, mgrClient client.Client) (*v1.Deployment, error) {
	labelsToQuery := fmt.Sprintf("%s=%s", "app", models.InstOperatorDeploymentLabel)
	selector, err := labels.Parse(labelsToQuery)
	if err != nil {
		return nil, fmt.Errorf("cannot parse label selector: %w", err)
	}

	deploymentList := &v1.DeploymentList{}
	if err := mgrClient.List(ctx, deploymentList, &client.ListOptions{LabelSelector: selector}); err != nil {
		return nil, fmt.Errorf("cannot get instaclustr deployment: %w", err)
	}

	if len(deploymentList.Items) != 1 {
		return nil, fmt.Errorf("expected exactly one deployment, found %d", len(deploymentList.Items))
	}

	return &deploymentList.Items[0], nil
}

func findContainer(deployment *v1.Deployment, containerName string) (*corev1.Container, error) {
	for _, c := range deployment.Spec.Template.Spec.Containers {
		if c.Name == containerName {
			return &c, nil
		}
	}

	return nil, fmt.Errorf("cannot find container %s in the deployment", containerName)
}

func getCurrentImageTag(image string) (string, error) {
	parts := strings.Split(image, ":")
	if len(parts) < 2 {
		return "", fmt.Errorf("cannot find tag in the image")
	}

	return parts[1], nil
}

func updateContainerImage(ctx context.Context, mgrClient client.Client, deployment *v1.Deployment, container *corev1.Container, newTag string) error {
	imageParts := strings.Split(container.Image, ":")
	container.Image = fmt.Sprintf("%s:%s", imageParts[0], newTag)

	return mgrClient.Update(ctx, deployment)
}
