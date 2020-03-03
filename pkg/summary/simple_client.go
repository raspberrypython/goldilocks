package summary

import (
	"github.com/fairwindsops/goldilocks/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	v1beta2 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
)

// SimpleClient calls the Kubernetes api to get all information
type SimpleClient struct {
	KubeClient    *kube.ClientInstance
	KubeClientVPA *kube.VPAClientInstance
}

// NewSimpleClient returns a SimpleClient
func NewSimpleClient() *SimpleClient {
	return &SimpleClient{
		KubeClient:    kube.GetInstance(),
		KubeClientVPA: kube.GetVPAInstance(),
	}
}

// ListVerticalPodAutoscalers implements Client
func (c *SimpleClient) ListVerticalPodAutoscalers(vpaLabels map[string]string) ([]v1beta2.VerticalPodAutoscaler, error) {
	// Get VPAs
	vpas, err := c.KubeClientVPA.Client.AutoscalingV1beta2().VerticalPodAutoscalers("").List(metav1.ListOptions{
		LabelSelector: labels.Set(vpaLabels).String(),
	})

	// Return any errors
	if err != nil {
		return nil, err
	}

	// Return items
	return vpas.Items, nil
}

// GetDeployment implements Client
func (c *SimpleClient) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	return c.KubeClient.Client.AppsV1().Deployments(namespace).Get(name, metav1.GetOptions{})
}

// GetDaemonSet implements Client
func (c *SimpleClient) GetDaemonSet(namespace, name string) (*appsv1.DaemonSet, error) {
	return c.KubeClient.Client.AppsV1().DaemonSets(namespace).Get(name, metav1.GetOptions{})
}

// GetStatefulSet implements Client
func (c *SimpleClient) GetStatefulSet(namespace, name string) (*appsv1.StatefulSet, error) {
	return c.KubeClient.Client.AppsV1().StatefulSets(namespace).Get(name, metav1.GetOptions{})
}
