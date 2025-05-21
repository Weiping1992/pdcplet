package vmiwatcher

import (
	"context"
	"log/slog"
	"os"
	vcache "pdcplet/pkg/cache"
	"pdcplet/pkg/internal/inpplat"
	"pdcplet/pkg/module"
	"sync"
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

const NAME = "VmiWatcher"

type vmiWatcherModule struct {
	name           string
	cache          vcache.Cache
	vmiInformer    cache.SharedIndexInformer
	kubevirtClient kubecli.KubevirtClient
	queue          workqueue.RateLimitingInterface
	proxy          inpplat.Proxy
}

func init() {
	module.RegisterInit(func() {
		module.RegisterConstructor(NAME, NewVmiWatcherModule) // register to module registry
	})
}

func NewVmiWatcherModule() module.Module {

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

	statusCache := vcache.NewVmiStatusCache()

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	vmiInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			vmi := obj.(*kubevirtv1.VirtualMachineInstance)
			slog.Debug("Recv Vmi Added Event", "vmiName", vmi.Name, "namespace", vmi.Namespace, "nodeName", vmi.Status.NodeName)
			// fmt.Printf("vqueuemi Added: %s/%s, nodeName: %s\n", vmi.Namespace, vmi.Name, vmi.Status.NodeName)
			var vmiStatus vcache.VmiStatus
			if isVmiReady(vmi) {
				vmiStatus = vcache.VmiStatusReady
			} else {
				vmiStatus = vcache.VmiStatusNotReady
			}
			if statusCache.Update(vmi.Name, vmiStatus) {
				item := workqueueItem{
					vmi: vmi,
					op:  CreateTaskOp,
				}
				queue.Add(item)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newVMI := newObj.(*kubevirtv1.VirtualMachineInstance)
			// fmt.Printf("vmi Updated: %s/%s, nodeName: %s\n", newVMI.Namespace, newVMI.Name, newVMI.Status.NodeName)
			slog.Debug("Recv Vmi Updated Event", "vmiName", newVMI.Name, "namespace", newVMI.Namespace, "nodeName", newVMI.Status.NodeName)
			// fmt.Printf("oldVmi: %v\n", oldObj.(*kubevirtv1.VirtualMachineInstance))
			// fmt.Printf("newVMI: %v\n", newVMI)
			// fmt.Println()
			var vmiStatus vcache.VmiStatus
			if isVmiReady(newVMI) {
				vmiStatus = vcache.VmiStatusReady
			} else {
				vmiStatus = vcache.VmiStatusNotReady
			}
			if statusCache.Update(newVMI.Name, vmiStatus) {
				item := workqueueItem{
					vmi: newVMI,
					op:  CreateTaskOp,
				}
				queue.Add(item)
			}
		},
		DeleteFunc: func(obj interface{}) {
			vmi := obj.(*kubevirtv1.VirtualMachineInstance)
			slog.Debug("Recv Vmi Deleted Event", "vmiName", vmi.Name, "namespace", vmi.Namespace, "nodeName", vmi.Status.NodeName)
			// fmt.Printf("vmi Deleted: %s/%s, nodeName: %s\n", vmi.Namespace, vmi.Name, vmi.Status.NodeName)
			statusCache.MarkDelete(vmi.Name)
			item := workqueueItem{
				vmi: vmi,
				op:  CloseTaskOp,
			}
			queue.Add(item)
		},
	})

	proxy := inpplat.NewMockProxy()

	return &vmiWatcherModule{
		name:           NAME,
		cache:          statusCache,
		vmiInformer:    vmiInformer,
		kubevirtClient: kubevirtClient,
		queue:          queue,
		proxy:          proxy,
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

func (a *vmiWatcherModule) Name() string {
	return a.name
}

func (a *vmiWatcherModule) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	stopCh := make(chan struct{})
	defer close(stopCh)
	go a.vmiInformer.Run(stopCh)

	if !cache.WaitForCacheSync(ctx.Done(), a.vmiInformer.HasSynced) {
		slog.Error("WaitForCacheSync timeout")
		return
	}

	queueCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-queueCtx.Done()
		slog.Info("Get exit signal", "ModuleName", a.name)
		a.queue.ShutDown()
	}()

	for {
		key, quit := a.queue.Get()
		if quit {
			return
		}
		a.doJob(key)
		a.queue.Done(key)
	}
}

func (a *vmiWatcherModule) doJob(key interface{}) {
	workItem := key.(workqueueItem)
	switch workItem.op {
	case CreateTaskOp:
		taskId, err := a.proxy.CreateTask(map[string]string{
			"name": workItem.vmi.Name,
		})
		if err != nil {
			slog.Error("CreateTask failed", "vmiName", workItem.vmi.Name, "taskId", taskId, "errMsg", err)
		} else {
			a.cache.SetTaskId(workItem.vmi.Name, taskId)
			slog.Info("CreateTask sucessfully", "taskId", taskId)
		}

	case CloseTaskOp:
		taskId, err := a.cache.GetTaskId(workItem.vmi.Name)
		if err != nil {
			slog.Error("GetTaskId from cache failed", "errMsg", err)
		}
		err = a.proxy.CloseTask(taskId)
		if err != nil {
			slog.Error("CloseTask failed", "vmiName", workItem.vmi.Name, "taskId", taskId, "errMsg", err)
		}
		a.cache.DeleteDone(workItem.vmi.Name)
	}
	slog.Debug("workqueue get vmi", "vmiName", workItem.vmi.Name)
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
