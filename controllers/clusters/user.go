package clusters

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/instaclustr/operator/apis/clusters/v1beta1"
	"github.com/instaclustr/operator/pkg/models"
)

type userObject interface {
	Object
	GetClusterEvents() map[string]string
	SetClusterEvents(events map[string]string)
}

type clusterObject interface {
	Object
	GetUserRefs() v1beta1.References
	SetUserRefs(refs v1beta1.References)
	GetAvailableUsers() v1beta1.References
	SetAvailableUsers(users v1beta1.References)
	GetClusterID() string
	SetClusterID(id string)
}

// userResourceFactory knows which user should be created.
// It helps client.Client to understand which resource should be got
type userResourceFactory interface {
	NewUserResource() userObject
}

// handleUsersChanges handles changes of user creation or deletion events
func handleUsersChanges(
	ctx context.Context,
	c client.Client,
	userFactory userResourceFactory,
	cluster clusterObject,
) error {
	l := log.FromContext(ctx).V(1)

	l.Info("users::handleUsersChanges")

	oldUsers := cluster.GetAvailableUsers()
	newUsers := cluster.GetUserRefs()
	clusterID := cluster.GetClusterID()

	added, deleted := oldUsers.Diff(newUsers)

	for _, ref := range added {
		err := setClusterEvent(ctx, c, ref, userFactory, clusterID, models.CreatingEvent)
		if err != nil {
			return err
		}
	}

	for _, ref := range deleted {
		err := setClusterEvent(ctx, c, ref, userFactory, clusterID, models.DeletingEvent)
		if err != nil {
			return err
		}
	}

	if added != nil || deleted != nil {
		patch := cluster.NewPatch()
		cluster.SetAvailableUsers(newUsers)
		err := c.Status().Patch(ctx, cluster, patch)
		if err != nil {
			return err
		}
	}

	return nil
}

// setClusterEvent sets a given event to the resource is got by ref
func setClusterEvent(
	ctx context.Context,
	c client.Client,
	ref *v1beta1.Reference,
	userFactory userResourceFactory,
	clusterID,
	event string,
) error {
	user := userFactory.NewUserResource()
	err := c.Get(ctx, ref.AsNamespacedName(), user)
	if err != nil {
		return err
	}

	patch := user.NewPatch()

	events := user.GetClusterEvents()
	if events == nil {
		events = map[string]string{}
		user.SetClusterEvents(events)
	}

	events[clusterID] = event
	return c.Status().Patch(ctx, user, patch)
}

// detachUsers detaches users from the given cluster
func detachUsers(
	ctx context.Context,
	c client.Client,
	userFactory userResourceFactory,
	cluster clusterObject,
) error {
	l := log.FromContext(ctx).V(1)

	l.Info("users::detachUsers")

	clusterID := cluster.GetClusterID()

	for _, ref := range cluster.GetAvailableUsers() {
		err := setClusterEvent(ctx, c, ref, userFactory, clusterID, models.ClusterDeletingEvent)
		if err != nil {
			return err
		}
	}

	return nil
}
