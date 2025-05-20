package app

import (
	"log/slog"
	"os"
	"time"

	"github.com/spf13/pflag"
	k8sv1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"
)

type APP interface {
	Run()
}

type app struct {
	vmiInformer    cache.SharedIndexInformer
	kubevirtClient kubecli.KubevirtClient
	queue          workqueue.RateLimitingInterface
}

func NewApp() *app {

	kubevirtClient, defaultNs := NewKubevirtClient()

	nodeName := getNodeName()
	slog.Debug("get NodeName from env", "nodename", nodeName)

	vmiInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options k8smetav1.ListOptions) (runtime.Object, error) {
				// options.FieldSelector = fields.OneTermEqualSelector("status.nodeName", nodeName).String()
				options.LabelSelector = "kubevirt.io/nodeName=" + nodeName
				return kubevirtClient.VirtualMachineInstance(defaultNs).List(&options)
			},
			WatchFunc: func(options k8smetav1.ListOptions) (watch.Interface, error) {
				// options.FieldSelector = fields.OneTermEqualSelector("status.nodeName", nodeName).String()
				options.LabelSelector = "kubevirt.io/nodeName=" + nodeName
				return kubevirtClient.VirtualMachineInstance(defaultNs).Watch(options)
			},
		},
		&kubevirtv1.VirtualMachineInstance{},
		10*time.Minute,
		cache.Indexers{},
	)

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	vmiInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			vmi := obj.(*kubevirtv1.VirtualMachineInstance)
			slog.Debug("Recv Vmi Added Event", "vmiName", vmi.Name, "namespace", vmi.Namespace, "nodeName", vmi.Status.NodeName)
			// fmt.Printf("vmi Added: %s/%s, nodeName: %s\n", vmi.Namespace, vmi.Name, vmi.Status.NodeName)
			if isVmiReady(vmi) {
				queue.Add(vmi.GetName())
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newVMI := newObj.(*kubevirtv1.VirtualMachineInstance)
			// fmt.Printf("vmi Updated: %s/%s, nodeName: %s\n", newVMI.Namespace, newVMI.Name, newVMI.Status.NodeName)
			slog.Debug("Recv Vmi Updated Event", "vmiName", newVMI.Name, "namespace", newVMI.Namespace, "nodeName", newVMI.Status.NodeName)
			// fmt.Printf("oldVmi: %v\n", oldObj.(*kubevirtv1.VirtualMachineInstance))
			// fmt.Printf("newVMI: %v\n", newVMI)
			// fmt.Println()
			if isVmiReady(newVMI) {
				queue.Add(newVMI.GetName())
			}
		},
		DeleteFunc: func(obj interface{}) {
			vmi := obj.(*kubevirtv1.VirtualMachineInstance)
			slog.Debug("Recv Vmi Deleted Event", "vmiName", vmi.Name, "namespace", vmi.Namespace, "nodeName", vmi.Status.NodeName)
			// fmt.Printf("vmi Deleted: %s/%s, nodeName: %s\n", vmi.Namespace, vmi.Name, vmi.Status.NodeName)
		},
	})

	return &app{
		vmiInformer:    vmiInformer,
		kubevirtClient: kubevirtClient,
		queue:          queue,
	}
}

func NewKubevirtClient() (kubecli.KubevirtClient, string) {
	// kubecli.DefaultClientConfig() prepares config using kubeconfig.
	// typically, you need to set env variable, KUBECONFIG=<path-to-kubeconfig>/.kubeconfig
	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})

	// retrive default namespace.
	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		//log.Fatalf("error in namespace : %v\n", err)
		slog.Error("retrive default namespace failed", "err", err)
	}

	// get the kubevirt client, using which kubevirt resources can be managed.
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		//log.Fatalf("cannot obtain KubeVirt client: %v\n", err)
		slog.Error("cannot obtain KubeVirt client", "err", err)
	}
	return virtClient, namespace
}

func (a *app) Run() {
	// vmList, vmiList := a.c.GetAllVmListInDefaultNamespace()
	// w := tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', 0)
	// fmt.Fprintln(w, "Type\tName\tNamespace\tStatus")

	// for _, vm := range vmList.Items {
	// 	fmt.Fprintf(w, "%s\t%s\t%s\t%v\n", vm.Kind, vm.Name, vm.Namespace, vm.Status.Ready)
	// }
	// for _, vmi := range vmiList.Items {
	// 	fmt.Fprintf(w, "%s\t%s\t%s\t%v\n", vmi.Kind, vmi.Name, vmi.Namespace, vmi.Status.Phase)
	// }
	// w.Flush()

	go func() {
		for {
			key, quit := a.queue.Get()
			if quit {
				return
			}
			defer a.queue.Done(key)

			// // 触发 HTTP 请求
			// resp, err := http.Post("https://your-api-endpoint", "application/json", nil)
			// if err != nil {
			// 	queue.AddRateLimited(key) // 失败重试
			// }
			slog.Debug("workqueue get vmi", "key", key)
		}
	}()

	stopCh := make(chan struct{})
	go a.vmiInformer.Run(stopCh)
	<-stopCh
}

func isVmiReady(vmi *kubevirtv1.VirtualMachineInstance) bool {
	for _, condition := range vmi.Status.Conditions {
		if condition.Type == kubevirtv1.VirtualMachineInstanceReady && condition.Status == k8sv1.ConditionTrue {
			// fmt.Printf("VMI %s is ready\n", vmi.Name)
			slog.Debug("VMI is ready", "vmiName", vmi.Name)
			return true
		}
	}
	return false
}

func getNodeName() string {
	nodeName := os.Getenv("NODE_NAME")
	if len(nodeName) == 0 {
		//fmt.Fprintln(os.Stderr, "Cannot read environment variable: NODE_NAME")
		slog.Error("Cannot read environment variable: NODE_NAME")
	}
	return nodeName
}
