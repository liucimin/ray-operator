package main

import (
	"flag"
	"math/rand"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	corev1 "k8s.io/api/core/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	rayv1 "github.com/ray-operator/pkg/ray-controller/k8s/apis/ray.io/v1"
	clientset "github.com/ray-operator/pkg/ray-controller/k8s/client/clientset/versioned"
	rayinformers "github.com/ray-operator/pkg/ray-controller/k8s/client/informers/externalversions"
	controllers "github.com/ray-operator/pkg/ray-controller/k8s/controllers"
	"github.com/ray-operator/pkg/ray-controller/k8s/crd"
	raycrd "github.com/ray-operator/pkg/ray-controller/k8s/crd/ray"
)

type Daemon struct {
	k8sResourceSyncWaitGroup sync.WaitGroup
	podController            controllers.ControllerInterface
	rayController            controllers.ControllerInterface
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	defer glog.Flush()

	d := &Daemon{}
	cm := RayControllerCmd(d)
	if err := cm.Execute(); err != nil {
		glog.Fatal("%v\n", err)
	}

}

func RayControllerCmd(d *Daemon) *cobra.Command {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	flag.CommandLine.Parse([]string{})

	var rayControllerCmd = &cobra.Command{
		Use:  "ray-controller",
		Long: "The ray-controller is the controller of the ray in the master node.",
		Run: func(cmd *cobra.Command, args []string) {
			glog.V(4).Infoln(viper.Get("master-url"))
			d.startRayController()

		},
	}
	rayControllerCmd.PersistentFlags().StringP("master-url", "u", "",
		"The kubernetes api-server url")
	rayControllerCmd.PersistentFlags().StringP("kube-config-path", "c", "",
		"The kubernetes api-server client config path")

	viper.BindPFlag("master-url", rayControllerCmd.PersistentFlags().Lookup("master-url"))
	viper.BindPFlag("kube-config-path", rayControllerCmd.PersistentFlags().Lookup("kube-config-path"))

	viper.SetEnvPrefix("ray") //set the env prefix
	viper.AutomaticEnv()

	return rayControllerCmd
}

func (d *Daemon) startRayController() {

	//flag.Parse()
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := make(chan struct{})
	//todo add the signal process

	//init the k8s client
	cfg, err := clientcmd.BuildConfigFromFlags(viper.GetString("master-url"), viper.GetString("kube-config-path"))
	if err != nil {
		glog.Errorf("Error building kubeconfig: %s", err.Error())
	}

	//for the core k8s api
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	//for the crd k8s api
	apiExtensionsClient, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	//for the cr resouce k8s api
	rayClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %s", err.Error())
	}

	//first create or update the crd in k8s
	err = crd.CreateOrUpdateCRD(apiExtensionsClient, raycrd.GetCRD())
	if err != nil {
		glog.Fatalf("failed to create or update CustomResourceDefinition %s: %v", raycrd.FullName, err)
	}

	//init the informer factory and set 30s list from the server
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	rayInformerFactory := rayinformers.NewSharedInformerFactory(rayClient, time.Second*30)

	podController, podCacheController := controllers.NewControllerFactory(&corev1.Pod{}, kubeInformerFactory,
		func(i interface{}) func() error {
			return func() error {
				err := d.addK8sPod(i.(*corev1.Pod))
				return err
			}
		},
		func(i interface{}) func() error {
			return func() error {
				err := d.delK8sPod(i.(*corev1.Pod))
				return err
			}
		},
		func(old, new interface{}) func() error {
			return func() error {
				err := d.updateK8sPod(old.(*corev1.Pod), new.(*corev1.Pod))
				return err
			}
		},
		rayClient,
		kubeClient,
	)

	d.podController = podController
	rayController, rayCacheController := controllers.NewControllerFactory(&rayv1.Ray{}, rayInformerFactory,
		func(i interface{}) func() error {
			return func() error {

				err := d.addK8sRay(i.(*rayv1.Ray))
				return err
			}
		},
		func(i interface{}) func() error {
			return func() error {
				err := d.delK8sRay(i.(*rayv1.Ray))
				return err
			}
		},
		func(old, new interface{}) func() error {
			return func() error {
				err := d.updateK8sRay(old.(*rayv1.Ray), new.(*rayv1.Ray))
				return err
			}
		},
		rayClient,
		kubeClient,
	)
	d.rayController = rayController

	kubeInformerFactory.Start(stopCh)
	rayInformerFactory.Start(stopCh)

	blockWaitGroupToSyncResources(&d.k8sResourceSyncWaitGroup, podCacheController)
	blockWaitGroupToSyncResources(&d.k8sResourceSyncWaitGroup, rayCacheController)

	d.k8sResourceSyncWaitGroup.Wait()

	<-stopCh

}

func (d *Daemon) addK8sPod(pod *corev1.Pod) error {
	glog.V(4).Infof("addK8sPod %v", pod)
	d.podController.SyncLoop("Pod")
	return nil
}

func (d *Daemon) delK8sPod(pod *corev1.Pod) error {
	glog.V(4).Infof("delK8sPod %v", pod)
	d.podController.SyncLoop("Pod")
	return nil
}

func (d *Daemon) updateK8sPod(old, new *corev1.Pod) error {
	glog.V(4).Infof("updateK8sPod %v %v", old, new)
	d.podController.SyncLoop("Pod")
	return nil
}

func (d *Daemon) addK8sRay(ray *rayv1.Ray) error {
	glog.V(4).Infof("addK8sRay %v", ray)
	d.rayController.SyncLoop("Ray")
	return nil
}

func (d *Daemon) delK8sRay(ray *rayv1.Ray) error {
	glog.V(4).Infof("delK8sRay %v", ray)
	d.rayController.SyncLoop("Ray")
	return nil
}

func (d *Daemon) updateK8sRay(old, new *rayv1.Ray) error {

	glog.V(4).Infof("updateK8sRay: %v %v", old, new)
	d.rayController.SyncLoop("Ray")
	return nil
}

// blockWaitGroupToSyncResources ensures that anything which waits on waitGroup
// waits until all objects of the specified resource stored in Kubernetes are
// received by the informer and processed by controller.
// Fatally exits if syncing these initial objects fails.
func blockWaitGroupToSyncResources(waitGroup *sync.WaitGroup, informer cache.Controller) {

	waitGroup.Add(1)
	go func() {
		//scopedLog := log.WithField("kubernetesResource", resourceName)
		//scopedLog.Debug("waiting for cache to synchronize")
		if ok := cache.WaitForCacheSync(wait.NeverStop, informer.HasSynced); !ok {
			// Fatally exit it resource fails to sync
			//scopedLog.Fatalf("failed to wait for cache to sync")
		}
		//scopedLog.Debug("cache synced")
		waitGroup.Done()
	}()
}
