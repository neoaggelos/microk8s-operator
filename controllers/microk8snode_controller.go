package controllers

import (
	"context"
	"time"

	microk8sv1alpha1 "github.com/neoaggelos/microk8s-operator/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type SnapInfo struct {
	Revision string
	Channel  string
	Version  string
}

type MicroK8sNodeController struct {
	Client   client.Client
	Interval time.Duration

	Node     string
	SnapInfo func(ctx context.Context) (SnapInfo, error)
}

func (c *MicroK8sNodeController) Run(ctx context.Context) error {
	log := log.FromContext(ctx).WithValues("node", c.Node)
	// cleanup on exit
	defer func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		node := &microk8sv1alpha1.MicroK8sNode{
			ObjectMeta: v1.ObjectMeta{
				Name: c.Node,
			},
		}
		if err := c.Client.Delete(ctx, node, &client.DeleteOptions{}); err != nil {
			log.Error(err, "failed to delete node during cleanup")
		}
	}()

	for {
		node := &microk8sv1alpha1.MicroK8sNode{}
		if err := c.Client.Get(ctx, types.NamespacedName{Name: c.Node}, node); err != nil {
			if apierrors.IsNotFound(err) {
				if err := c.Client.Create(ctx, &microk8sv1alpha1.MicroK8sNode{ObjectMeta: v1.ObjectMeta{Name: c.Node}}); err != nil {
					log.Error(err, "failed to create node")
				}
			}

			if err := c.Client.Get(ctx, types.NamespacedName{Name: c.Node}, node); err != nil {
				log.Error(err, "failed to get node")
				continue
			}
		}

		snapInfo, err := c.SnapInfo(ctx)
		if err != nil {
			log.Error(err, "failed to retrieve microk8s snap info")
		}
		node.Status.Channel = snapInfo.Channel
		node.Status.Revision = snapInfo.Revision
		node.Status.Version = snapInfo.Version
		node.Status.LastUpdate.Time = time.Now()

		if err := c.Client.Status().Update(ctx, node); err != nil {
			log.Error(err, "failed to update node")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.Interval):
		}
	}
}
