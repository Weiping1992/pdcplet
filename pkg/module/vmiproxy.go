package module

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	vcache "pdcplet/pkg/cache"
	"pdcplet/pkg/internal/inpplat"
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

const VMI_PROXY_NAME = "VmiProxy"

type vmiProxyModule struct {
	name           string
	cache          vcache.Cache
	vmiInformer    cache.SharedIndexInformer
	kubevirtClient kubecli.KubevirtClient
	queue          workqueue.RateLimitingInterface
	inpclient      inpplat.Client
}



type VmiProxyConfig struct {
	Proxy struct{
			Addr      string
			Port      string
			BaseUrl   string
			AuthToken string
		}

}

func NewVmiProxyModule(params ...interface{}) (Module, error) {

	if len(params) != 1 || params[0]. == nil {
		return nil, fmt.Errorf("VmiProxyModule params error")

	}


	kubevirtClient, defaultNs := NewKubevirtClient()

	nodeName, err := getNodeName()
	if err != nil {
		return nil, err
	}
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
			var vmiStatus vcache.VmiStatus
			if isVmiReady(vmi) {
				vmiStatus = vcache.VmiStatusReady
			} else {
				vmiStatus = vcache.VmiStatusNotReady
			}
			isStatusChanged := statusCache.Update(vmi.Name, vmiStatus)
			isTaskCreated, err := statusCache.IsTaskCreated(vmi.Name)
			if err != nil {
				slog.Error("IsTaskCreated() get vmi taskStatus failed", "vmiName", vmi.Name, "errMsg", err)
				isTaskCreated = false
			}
			if isStatusChanged && vmiStatus == vcache.VmiStatusReady && !isTaskCreated {
				item := workqueueItem{
					vmi: vmi,
					op:  CreateTaskOp,
				}
				queue.Add(item)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newVMI := newObj.(*kubevirtv1.VirtualMachineInstance)
			slog.Debug("Recv Vmi Updated Event", "vmiName", newVMI.Name, "namespace", newVMI.Namespace, "nodeName", newVMI.Status.NodeName)
			var vmiStatus vcache.VmiStatus
			if isVmiReady(newVMI) {
				vmiStatus = vcache.VmiStatusReady
			} else {
				vmiStatus = vcache.VmiStatusNotReady
			}
			isStatusChanged := statusCache.Update(newVMI.Name, vmiStatus)
			isTaskCreated, err := statusCache.IsTaskCreated(newVMI.Name)
			if err != nil {
				slog.Error("IsTaskCreated() get vmi taskStatus failed", "vmiName", newVMI.Name, "errMsg", err)
				isTaskCreated = false
			}
			if isStatusChanged {
				if vmiStatus == vcache.VmiStatusReady && !isTaskCreated {
					item := workqueueItem{
						vmi: newVMI,
						op:  CreateTaskOp,
					}
					queue.Add(item)
				}
				if vmiStatus == vcache.VmiStatusNotReady && isTaskCreated {
					item := workqueueItem{
						vmi: newVMI,
						op:  CloseTaskOp,
					}
					queue.Add(item)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			vmi := obj.(*kubevirtv1.VirtualMachineInstance)
			slog.Debug("Recv Vmi Deleted Event", "vmiName", vmi.Name, "namespace", vmi.Namespace, "nodeName", vmi.Status.NodeName)
			isTaskClosed, err := statusCache.IsTaskClosed(vmi.Name)
			if err != nil {
				slog.Error("IsTaskClosed() get vmi taskStatus failed", "vmiName", vmi.Name, "errMsg", err)
				isTaskClosed = false
			}
			isTaskCreated, err := statusCache.IsTaskCreated(vmi.Name)
			if err != nil {
				slog.Error("IsTaskCreated() get vmi taskStatus failed", "vmiName", vmi.Name, "errMsg", err)
				isTaskCreated = false
			}
			if isTaskCreated && !isTaskClosed {
				item := workqueueItem{
					vmi: vmi,
					op:  CloseTaskOp,
				}
				queue.Add(item)
			}
			statusCache.Delete(vmi.Name)
		},
	})

	proxy := inpplat.NewClient(
		addr: 
	)

	return &vmiProxyModule{
		name:           VMI_PROXY_NAME,
		cache:          statusCache,
		vmiInformer:    vmiInformer,
		kubevirtClient: kubevirtClient,
		queue:          queue,
		inpclient:      proxy,
	}, nil
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

func (a *vmiProxyModule) Name() string {
	return a.name
}

func (a *vmiProxyModule) Run(ctx context.Context, wg *sync.WaitGroup) {
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

func (a *vmiProxyModule) doJob(key interface{}) {
	workItem := key.(workqueueItem)
	slog.Debug("workqueue get vmi", "vmiName", workItem.vmi.Name)
	switch workItem.op {
	case CreateTaskOp:
		taskId, err := a.inpclient.CreateTask(map[string]string{
			"name": workItem.vmi.Name,
		})
		if err != nil {
			slog.Error("CreateTask failed", "vmiName", workItem.vmi.Name, "taskId", taskId, "errMsg", err)
			a.queue.AddRateLimited(workItem)
		} else {
			a.cache.SetTaskId(workItem.vmi.Name, taskId)
			slog.Info("CreateTask sucessfully", "taskId", taskId)
			a.cache.MarkTaskCreated(workItem.vmi.Name)
			a.queue.Forget(workItem)
		}

	case CloseTaskOp:
		taskId, err := a.cache.GetTaskId(workItem.vmi.Name)
		if err != nil {
			slog.Error("GetTaskId from cache failed", "errMsg", err)
		}
		err = a.inpclient.CloseTask(taskId)
		if err != nil {
			slog.Error("CloseTask failed", "vmiName", workItem.vmi.Name, "taskId", taskId, "errMsg", err)
			a.queue.AddRateLimited(workItem)
		} else {
			slog.Info("CloseTask sucessfully", "taskId", taskId)
			a.cache.MarkTaskClosed(workItem.vmi.Name)
			a.queue.Forget(workItem)
		}
	}
}

type OperateType int

const (
	CreateTaskOp OperateType = iota
	CloseTaskOp
)

type workqueueItem struct {
	vmi *kubevirtv1.VirtualMachineInstance
	op  OperateType
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

func getNodeName() (string, error) {
	nodeName := os.Getenv("NODE_NAME")
	if len(nodeName) == 0 {
		slog.Error("Cannot read environment variable: NODE_NAME")
		return "", fmt.Errorf("Cannot read environment variable: NODE_NAME")
	}
	return nodeName, nil
}
