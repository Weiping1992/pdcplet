package kubeclient

import (
	"log"

	"github.com/spf13/pflag"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubevirtv1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"
)

type KubeClient interface {
	GetAllVmListInDefaultNamespace() (*kubevirtv1.VirtualMachineList, *kubevirtv1.VirtualMachineInstanceList)
}

type defaultKubeClient struct {
	virtClient       kubecli.KubevirtClient
	defaultNameSpace string
}

func NewKubeClient() KubeClient {
	// kubecli.DefaultClientConfig() prepares config using kubeconfig.
	// typically, you need to set env variable, KUBECONFIG=<path-to-kubeconfig>/.kubeconfig
	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})

	// retrive default namespace.
	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		log.Fatalf("error in namespace : %v\n", err)
	}

	// get the kubevirt client, using which kubevirt resources can be managed.
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		log.Fatalf("cannot obtain KubeVirt client: %v\n", err)
	}

	return &defaultKubeClient{
		virtClient:       virtClient,
		defaultNameSpace: namespace,
	}
}

func (c *defaultKubeClient) GetAllVmListInDefaultNamespace() (*kubevirtv1.VirtualMachineList, *kubevirtv1.VirtualMachineInstanceList) {
	// Fetch list of VMs & VMIs
	vmList, err := c.virtClient.VirtualMachine(c.defaultNameSpace).List(&k8smetav1.ListOptions{})
	if err != nil {
		log.Fatalf("cannot obtain KubeVirt vm list: %v\n", err)
	}
	vmiList, err := c.virtClient.VirtualMachineInstance(c.defaultNameSpace).List(&k8smetav1.ListOptions{})
	if err != nil {
		log.Fatalf("cannot obtain KubeVirt vmi list: %v\n", err)
	}
	return vmList, vmiList
}
