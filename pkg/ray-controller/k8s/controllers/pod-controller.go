package controllers

import (
	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	rayclientset "github.com/ray-operator/pkg/ray-controller/k8s/client/clientset/versioned"
)

const (
	controllerName = "Pod"
)

func init() {

	ControllerRigist(
		&v1.Pod{},
		copyObjToV1Pod,
		newPodController,
	)
}

func copyObjToV1Pod(obj interface{}) meta_v1.Object {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		glog.Warning("Ignoring invalid k8s v1 Service")
		return nil
	}
	return pod.DeepCopy()
}

func newPodController(informerFactory interface{},
	handlerFuncs *cache.ResourceEventHandlerFuncs,
	rayClient rayclientset.Interface,
	kubeClient clientset.Interface) (ControllerInterface, cache.Controller) {

	podController := PodController{
		rayClient:  rayClient,
		kubeClient: kubeClient,
	}

	kubeInformerFactory, _ := informerFactory.(kubeinformers.SharedInformerFactory)

	podInformer := kubeInformerFactory.Core().V1().Pods()

	podInformer.Informer().AddEventHandler(handlerFuncs)

	return &podController, podInformer.Informer()

}

type PodController struct {
	rayClient  rayclientset.Interface
	kubeClient clientset.Interface
}

func (c *PodController) SyncLoop(key string) error {

	glog.Warning("PodController syncLoop")
	return nil

}
