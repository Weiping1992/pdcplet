package kubevirt

import (
	"log/slog"

	"github.com/spf13/pflag"
	"k8s.io/client-go/util/retry"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"
)

func ApplyVM(vm *kubevirtv1.VirtualMachine) error {
	// kubecli.DefaultClientConfig() prepares config using kubeconfig.
	// typically, you need to set env variable, KUBECONFIG=<path-to-kubeconfig>/.kubeconfig
	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})

	// get the kubevirt client, using which kubevirt resources can be managed.
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		//log.Fatalf("cannot obtain KubeVirt client: %v\n", err)
		slog.Error("cannot obtain KubeVirt client", "err", err)
		return err
	}

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		_, err := virtClient.VirtualMachine(vm.Namespace).Create(vm)
		return err
	})
}
