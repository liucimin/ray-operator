package controllers

import (
	"fmt"
	"reflect"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	rayclientset "github.com/ray-operator/pkg/ray-controller/k8s/client/clientset/versioned"
	"github.com/ray-operator/pkg/ray-controller/k8s/funcqueue"
)

type newController func(interface{}, *cache.ResourceEventHandlerFuncs, rayclientset.Interface, clientset.Interface) (ControllerInterface, cache.Controller)

var (
	controllerMap = map[reflect.Type]newController{}
	deepCopyMap   = map[reflect.Type]func(interface{}) meta_v1.Object{}
)

type ControllerInterface interface {
	SyncLoop(key string) error
}

func ControllerRigist(resourceObj interface{},
	deepCopyF func(interface{}) meta_v1.Object,
	f newController) {

	typeOfObj := reflect.TypeOf(resourceObj)
	controllerMap[typeOfObj] = f
	deepCopyMap[typeOfObj] = deepCopyF

}

func castFuncFactory(i interface{}) func(interface{}) meta_v1.Object {
	castFunc, ok := deepCopyMap[reflect.TypeOf(i)]
	if !ok {
		panic(fmt.Sprintf("Object type '%s' not registered", reflect.TypeOf(i)))
	}
	return castFunc
}

func NewControllerFactory(
	resourceObj interface{},
	shareInformerFactory interface{},
	addFunc, delFunc func(i interface{}) func() error,
	updateFunc func(old, new interface{}) func() error,
	rayclient rayclientset.Interface,
	kubeClient clientset.Interface) (ControllerInterface, cache.Controller) {

	//todo add relist func

	fqueue := funcqueue.NewFunctionQueue(1024)
	castToDeepCopy := castFuncFactory(resourceObj)
	rehf := &cache.ResourceEventHandlerFuncs{}
	if addFunc != nil {
		rehf.AddFunc = func(obj interface{}) {
			if metaObj := castToDeepCopy(obj); metaObj != nil {

				fqueue.Enqueue(addFunc(metaObj), funcqueue.NoRetry)
			}
		}
	}
	if updateFunc != nil {
		rehf.UpdateFunc = func(oldObj, newObj interface{}) {

			if oldMetaObj := castToDeepCopy(oldObj); oldMetaObj != nil {
				if newMetaObj := castToDeepCopy(newObj); newMetaObj != nil {
					fqueue.Enqueue(updateFunc(oldMetaObj, newMetaObj), funcqueue.NoRetry)
				}
			}
		}
	}

	if delFunc != nil {
		rehf.DeleteFunc = func(obj interface{}) {
			if metaObj := castToDeepCopy(obj); metaObj != nil {
				fqueue.Enqueue(delFunc(metaObj), funcqueue.NoRetry)
			}
		}
	}

	return controllerMap[reflect.TypeOf(resourceObj)](shareInformerFactory, rehf, rayclient, kubeClient)

}
