package kubevirt

import (
	"encoding/json"
	"log/slog"

	"github.com/spf13/pflag"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"kubevirt.io/client-go/kubecli"
)

var virtClient kubecli.KubevirtClient

func init() {
	// kubecli.DefaultClientConfig() prepares config using kubeconfig.
	// typically, you need to set env variable, KUBECONFIG=<path-to-kubeconfig>/.kubeconfig
	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})

	// get the kubevirt client, using which kubevirt resources can be managed.
	client, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		//log.Fatalf("cannot obtain KubeVirt client: %v\n", err)
		slog.Error("cannot obtain KubeVirt client", "err", err)

	}
	virtClient = client
}

func CreateVM(vm *kubevirtv1.VirtualMachine) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		_, err := virtClient.VirtualMachine(vm.Namespace).Create(vm)
		return err
	})
}

func DeleteVM(vmName string, vmNamespace string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err := virtClient.VirtualMachine(vmNamespace).Delete(vmName, &k8smetav1.DeleteOptions{})
		return err
	})
}

func StartVM(vm *kubevirtv1.VirtualMachine) error {
	patchData := map[string]interface{}{
		"spec": map[string]interface{}{
			"running": true,
		},
	}
	patchBytes, _ := json.Marshal(patchData)
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		_, err := virtClient.VirtualMachine(vm.Namespace).Patch(vm.Name, types.JSONPatchType, patchBytes, &k8smetav1.PatchOptions{}, "")
		return err
	})
}

func Stop(vm *kubevirtv1.VirtualMachine) error {
	patchData := map[string]interface{}{
		"spec": map[string]interface{}{
			"running": false,
		},
	}
	patchBytes, _ := json.Marshal(patchData)
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		_, err := virtClient.VirtualMachine(vm.Namespace).Patch(vm.Name, types.JSONPatchType, patchBytes, &k8smetav1.PatchOptions{}, "")
		return err
	})
}
