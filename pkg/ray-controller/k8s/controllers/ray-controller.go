package controllers

import (
	"github.com/golang/glog"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"

	rayv1 "github.com/ray-operator/pkg/ray-controller/k8s/apis/ray.io/v1"
	rayclientset "github.com/ray-operator/pkg/ray-controller/k8s/client/clientset/versioned"
	rayscheme "github.com/ray-operator/pkg/ray-controller/k8s/client/clientset/versioned/scheme"
	rayinformers "github.com/ray-operator/pkg/ray-controller/k8s/client/informers/externalversions"
	raylisters "github.com/ray-operator/pkg/ray-controller/k8s/client/listers/ray.io/v1"
)

func init() {

	ControllerRigist(
		&rayv1.Ray{},
		copyObjToV1Ray,
		newRayController,
	)
}

func copyObjToV1Ray(obj interface{}) meta_v1.Object {
	cnn, ok := obj.(*rayv1.Ray)
	if !ok {
		glog.Warning("Ignoring invalid k8s v1 Service")
		return nil
	}
	return cnn.DeepCopy()
}

func newRayController(informerFactory interface{},
	handlerFuncs *cache.ResourceEventHandlerFuncs,
	rayClient rayclientset.Interface,
	kubeClient clientset.Interface) (ControllerInterface, cache.Controller) {

	rayController := RayController{
		rayClient:  rayClient,
		kubeClient: kubeClient,
	}

	kubeInformerFactory, _ := informerFactory.(rayinformers.SharedInformerFactory)

	rayInformer := kubeInformerFactory.Ray().V1().Rays()

	// Add sample-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(rayscheme.AddToScheme(scheme.Scheme))

	// Set up an event handler for when Foo resources change,and the pod informer will register in factory
	//when factory start the pod informer will also start
	rayInformer.Informer().AddEventHandler(handlerFuncs)
	rayController.rayLister = rayInformer.Lister()
	return &rayController, rayInformer.Informer()

}

type RayController struct {
	rayClient  rayclientset.Interface
	kubeClient clientset.Interface
	rayLister  raylisters.RayLister
}

func (c *RayController) SyncLoop(key string) error {

	glog.Warning("RayController syncLoop")
	return nil
}
