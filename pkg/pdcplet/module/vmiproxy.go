package module

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"pdcplet/pkg/config"
	"pdcplet/pkg/internal/inpplat"
	vcache "pdcplet/pkg/pdcplet/cache"
	"strconv"
	"strings"
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

const (
	DEFAULT_EVENT_HANDLER_RESYNC_PERIOD = 1 * time.Hour
)

type vmiProxyModule struct {
	name           string
	cache          vcache.Cache
	vmiInformer    cache.SharedIndexInformer
	kubevirtClient kubecli.KubevirtClient
	queue          workqueue.RateLimitingInterface
	inpplatproxy   inpplat.Client
}

type WatchMode int

const (
	WatchModeUnsupported WatchMode = iota
	WatchModeListWatch
	WatchModeWebhook
)

type VmiProxyConfig struct {
	Proxy struct {
		Addr      string
		Port      string
		BaseUrl   string
		AuthToken string
	}
}

func NewVmiProxyModule(params map[string]interface{}) (Module, error) {

	if len(params) == 0 {
		slog.Error("VmiProxyModule params is nil or empty")
		return nil, fmt.Errorf("VmiProxyModule params is nil or empty")
	}
	// fmt.Println("NewVmiProxyModule called with params:", params)

	vpm := &vmiProxyModule{
		name: VMI_PROXY_NAME,
	}

	if conns, ok := params["connections"]; !ok || len(conns.([]map[string]interface{})) == 0 {
		slog.Error("VmiProxyModule connections is nil or empty")
		return nil, fmt.Errorf("VmiProxyModule connections is nil or empty")
	}

	for _, conn := range params["connections"].([]map[string]interface{}) {
		if conn["name"] == config.INPPLAT_CONNECTION_NAME {
			proxy, err := NewInpProxy(conn)
			if err != nil {
				slog.Error("NewInpProxy failed", "errMsg", err)
				return nil, fmt.Errorf("NewInpProxy failed: %w", err)
			}
			vpm.inpplatproxy = proxy
			break
		}
	}
	if vpm.inpplatproxy == nil {
		vpm.inpplatproxy = inpplat.NewMockClient()
	}

	var defaultEventHandlerResyncPeriod time.Duration
	if v, ok := params["InformerResyncPeriod"]; !ok {
		defaultEventHandlerResyncPeriod = DEFAULT_EVENT_HANDLER_RESYNC_PERIOD
	} else {
		defaultEventHandlerResyncPeriod = convertToTimeDuration(v.(string), DEFAULT_EVENT_HANDLER_RESYNC_PERIOD)
	}

	var wm WatchMode
	if mode, ok := params["k8sWatchMode"]; ok {
		wm = parseWatchModeFlag(mode.(string))
	} else {
		wm = WatchModeListWatch
	}

	switch wm {
	case WatchModeWebhook:
		slog.Error("Webhook watchmode DO NOT IMPLEMENT")
		return nil, fmt.Errorf("webhook watchmode DO NOT IMPLEMENT")
	case WatchModeListWatch:
		slog.Debug("Using ListWatch mode for VMI Proxy Module")
		vmiInformer, kubevirtClient, statusCache, queue, err := NewVmiInformer("", defaultEventHandlerResyncPeriod)
		if err != nil {
			slog.Error("NewVmiInformer failed", "errMsg", err)
			return nil, fmt.Errorf("NewVmiInformer failed: %w", err)
		}
		vpm.vmiInformer = vmiInformer
		vpm.kubevirtClient = kubevirtClient
		vpm.cache = statusCache
		vpm.queue = queue
	default:
		slog.Error("Unknow WatchMode which must in ['listwatch', 'webhook']")
		return nil, fmt.Errorf("unknow WatchMode which must in ['listwatch', 'webhook']")
	}

	if vpm.vmiInformer == nil || vpm.kubevirtClient == nil || vpm.cache == nil || vpm.queue == nil || vpm.inpplatproxy == nil {
		slog.Error("VmiProxyModule init failed, vmiInformer, kubevirtClient, cache, queue or inpplatproxy is nil")
		return nil, fmt.Errorf("VmiProxyModule init failed, vmiInformer, kubevirtClient, cache or queue is nil")
	}

	return vpm, nil
}

func NewVmiInformer(namespace string, defaultEventHandlerResyncPeriod time.Duration) (
	cache.SharedIndexInformer,
	kubecli.KubevirtClient,
	vcache.Cache,
	workqueue.RateLimitingInterface, error) {

	kubevirtClient, defaultNs := NewKubevirtClient()

	if namespace == "" {
		namespace = defaultNs
	}

	nodeName, err := getNodeName()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	slog.Debug("get NodeName from env", "nodename", nodeName)

	vmiInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options k8smetav1.ListOptions) (runtime.Object, error) {
				// options.FieldSelector = fields.OneTermEqualSelector("status.nodeName", nodeName).String()
				options.LabelSelector = "kubevirt.io/nodeName=" + nodeName
				return kubevirtClient.VirtualMachineInstance(namespace).List(&options)
			},
			WatchFunc: func(options k8smetav1.ListOptions) (watch.Interface, error) {
				// options.FieldSelector = fields.OneTermEqualSelector("status.nodeName", nodeName).String()
				options.LabelSelector = "kubevirt.io/nodeName=" + nodeName
				return kubevirtClient.VirtualMachineInstance(namespace).Watch(options)
			},
		},
		&kubevirtv1.VirtualMachineInstance{},
		defaultEventHandlerResyncPeriod,
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
	return vmiInformer, kubevirtClient, statusCache, queue, nil
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

func NewInpProxy(connConfig map[string]interface{}) (inpplat.Client, error) {
	if len(connConfig) == 0 {
		slog.Error("VmiProxyModule connConfig is nil or empty")
		return nil, fmt.Errorf("VmiProxyModule connConfig is nil or empty")
	}

	tp, ok := connConfig["type"].(string)
	if !ok {
		tp = "httpOverTcpIp"
	}
	if tp != "httpOverTcpIp" && tp != "httpOverUnixSocket" {
		slog.Error("VmiProxyModule connConfig type is not supported", "type", tp)
		return nil, fmt.Errorf("VmiProxyModule connConfig type is not supported: %s", tp)
	}

	switch tp {
	case "httpOverTcpIp":
		hConf := connConfig["httpOverTcpIp"].(map[string]interface{})
		host, hostOk := hConf["host"].(string)
		// fmt.Println("host:", host, "hostOk:", hostOk)
		port, portOk := hConf["port"].(int)
		urlPrefix, upOk := hConf["urlPrefix"].(string)
		authToken, atOk := hConf["authToken"].(string)
		timeout, tOk := hConf["timeout"].(string)

		if !hostOk || !portOk {
			slog.Error("VmiProxyModule connConfig host or port is not set")
			return nil, fmt.Errorf("VmiProxyModule connConfig host or port is not set")
		}

		if !upOk {
			urlPrefix = ""
		}
		if !atOk {
			authToken = ""
		}
		if !tOk {
			timeout = "5s"
		}

		proxy := inpplat.NewClient(host, strconv.Itoa(port), urlPrefix, authToken, convertToTimeDuration(timeout, 5*time.Second))
		return proxy, nil
	default:
		slog.Error("VmiProxyModule connConfig type is not supported", "type", tp)
		return nil, fmt.Errorf("VmiProxyModule connConfig type is not supported: %s", tp)
	}
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
		taskId, err := a.inpplatproxy.CreateTask(map[string]string{
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
		err = a.inpplatproxy.CloseTask(taskId)
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
		return "", fmt.Errorf("cannot read environment variable: NODE_NAME")
	}
	return nodeName, nil
}

func parseWatchModeFlag(mode string) (m WatchMode) {
	if strings.ToLower(mode) == "webhook" {
		m = WatchModeWebhook
	} else if strings.ToLower(mode) == "listwatch" {
		m = WatchModeListWatch
	} else {
		m = WatchModeUnsupported
	}
	return
}

func convertToTimeDuration(t string, defaultValue time.Duration) time.Duration {
	if t == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(t)
	if err != nil {
		slog.Error("Failed to parse duration", "duration", t, "error", err)
		return defaultValue
	}
	return d
}
