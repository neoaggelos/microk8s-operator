package controllers

import (
	"context"
	"time"

	microk8sv1alpha1 "github.com/neoaggelos/microk8s-operator/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type MicroK8sNodeController struct {
	Client   client.Client
	Interval time.Duration

	Node         string
	SnapRevision func(ctx context.Context) string
	SnapChannel  func(ctx context.Context) string
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
			log.Error(err, "Failed to delete node during cleanup")
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.Interval):
		}

		node := &microk8sv1alpha1.MicroK8sNode{
			ObjectMeta: v1.ObjectMeta{
				Name: c.Node,
			},
		}

		node.Status.Channel = c.SnapChannel(ctx)
		node.Status.Revision = c.SnapRevision(ctx)
		node.Status.LastUpdate.Time = time.Now()

		if err := c.Client.Update(ctx, node); err != nil {
			if apierrors.IsNotFound(err) {
				err = c.Client.Create(ctx, node)
			}
			if err != nil {
				log.Error(err, "Failed to update node")
			}
		}
	}
}
