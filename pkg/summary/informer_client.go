package summary

import (
	"sort"

	"github.com/fairwindsops/goldilocks/pkg/kube"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	v1beta2 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"
	"k8s.io/client-go/tools/cache"
)

// InformerClient uses informer caches to return the data
type InformerClient struct {
	vpaInf cache.SharedIndexInformer
	depInf cache.SharedIndexInformer
	dsInf  cache.SharedIndexInformer
}

// NewInformerClient returns a InformerClient
func NewInformerClient() *InformerClient {
	client := kube.GetInstance()
	clientVPA := kube.GetVPAInstance()

	return &InformerClient{
		depInf: cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
					return client.Client.AppsV1().Deployments("").List(metav1.ListOptions{})
				},
				WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
					return client.Client.AppsV1().Deployments("").Watch(metav1.ListOptions{})
				},
			},
			&appsv1.Deployment{},
			0,
			cache.Indexers{},
		),

		dsInf: cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
					return client.Client.AppsV1().DaemonSets("").List(metav1.ListOptions{})
				},
				WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
					return client.Client.AppsV1().DaemonSets("").Watch(metav1.ListOptions{})
				},
			},
			&appsv1.DaemonSet{},
			0,
			cache.Indexers{},
		),

		vpaInf: cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
					return clientVPA.Client.AutoscalingV1beta2().VerticalPodAutoscalers("").List(metav1.ListOptions{})
				},
				WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
					return clientVPA.Client.AutoscalingV1beta2().VerticalPodAutoscalers("").Watch(metav1.ListOptions{})
				},
			},
			&v1beta2.VerticalPodAutoscaler{},
			0,
			cache.Indexers{},
		),
	}
}

// Run starts all the informers
func (c *InformerClient) Run(stopc chan struct{}) {
	go c.vpaInf.Run(stopc)
	go c.depInf.Run(stopc)
	go c.dsInf.Run(stopc)
	<-stopc
}

// ListVerticalPodAutoscalers implements Client
func (c *InformerClient) ListVerticalPodAutoscalers(vpaLabels map[string]string) ([]v1beta2.VerticalPodAutoscaler, error) {
	// Get all VPAs from the cache
	allObjects := c.vpaInf.GetStore().List()

	// Create list for filtered values
	vpaList := make([]v1beta2.VerticalPodAutoscaler, 0, len(allObjects))

	// Create selector
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: vpaLabels,
	})

	// Handle error creatign selector
	if err != nil {
		return nil, err
	}

	// Filter items from cache
	for _, vpa := range allObjects {
		if selector.Matches(labels.Set(vpa.(*v1beta2.VerticalPodAutoscaler).Labels)) {
			vpaList = append(vpaList, *vpa.(*v1beta2.VerticalPodAutoscaler).DeepCopy())
		}
	}

	// Sort by name
	sort.Slice(vpaList, func(i, j int) bool {
		return vpaList[i].Name < vpaList[j].Name
	})

	// Return filtered and sorted list
	return vpaList, nil
}

// GetDeployment implements Client
func (c *InformerClient) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	// Get item from informer cache
	item, exists, err := c.depInf.GetStore().GetByKey(namespace + "/" + name)

	// Handle error
	if err != nil {
		return nil, err
	}

	// If item does not exist return an error
	if !exists {
		return nil, errors.NewNotFound(schema.GroupResource{Group: "apps", Resource: "deployments"}, name)
	}

	// Return the object
	return item.(*appsv1.Deployment).DeepCopy(), nil
}

// GetDaemonSet implements Client
func (c *InformerClient) GetDaemonSet(namespace, name string) (*appsv1.DaemonSet, error) {
	// Get item from informer cache
	item, exists, err := c.dsInf.GetStore().GetByKey(namespace + "/" + name)

	// Handle error
	if err != nil {
		return nil, err
	}

	// If item does not exist return an error
	if !exists {
		return nil, errors.NewNotFound(schema.GroupResource{Group: "apps", Resource: "daemonsets"}, name)
	}

	// Return the object
	return item.(*appsv1.DaemonSet).DeepCopy(), nil
}
