package integrations

import (
	"context"
	"fmt"
	"log"
	"time"

	"helmjet-atlas/internal/models"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// SyncK8sOnce connects to Kubernetes and syncs Services into storage once.
func (i *Integrations) SyncK8sOnce(ctx context.Context, kubeconfigPath string, namespaces []string) error {
	var cfg *rest.Config
	var err error
	// Try in-cluster first
	cfg, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig file
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return fmt.Errorf("failed to build kubernetes config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("failed to create k8s clientset: %w", err)
	}

	return i.syncServices(ctx, clientset, namespaces)
}
func (i *Integrations) syncServices(ctx context.Context, clientset *kubernetes.Clientset, namespaces []string) error {
	svcList, err := clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}

	now := time.Now()
	synced := 0
	// build namespace set if provided
	nsSet := map[string]struct{}{}
	if len(namespaces) > 0 {
		for _, n := range namespaces {
			nsSet[n] = struct{}{}
		}
	}

	for _, s := range svcList.Items {
		if len(nsSet) > 0 {
			if _, ok := nsSet[s.Namespace]; !ok {
				continue
			}
		}
		service := &models.Microservice{
			Name:      s.Name,
			Namespace: s.Namespace,
			Cluster:   "",
			Ports: func() []models.Port {
				var ports []models.Port
				for _, p := range s.Spec.Ports {
					ports = append(ports, models.Port{Name: p.Name, ContainerPort: int32(p.Port), Protocol: string(p.Protocol)})
				}
				return ports
			}(),
			UpdatedAt:  now,
			LastSynced: now,
		}

		if err := i.MS.UpsertByNameNamespace(ctx, service); err != nil {
			log.Printf("failed to upsert service %s/%s: %v", s.Namespace, s.Name, err)
			continue
		}
		synced++
	}
	log.Printf("Synced %d k8s services", synced)
	return nil
}
