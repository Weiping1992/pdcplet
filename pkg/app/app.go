package app

import (
	"fmt"
	"os"
	kclient "pdcplet/pkg/kubeclient"
	"text/tabwriter"
)

type APP interface {
	Run()
}

type app struct {
	c kclient.KubeClient
}

func NewApp() *app {
	return &app{
		c: kclient.NewKubeClient(),
	}
}

func (a *app) Run() {
	vmList, vmiList := a.c.GetAllVmListInDefaultNamespace()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', 0)
	fmt.Fprintln(w, "Type\tName\tNamespace\tStatus")

	for _, vm := range vmList.Items {
		fmt.Fprintf(w, "%s\t%s\t%s\t%v\n", vm.Kind, vm.Name, vm.Namespace, vm.Status.Ready)
	}
	for _, vmi := range vmiList.Items {
		fmt.Fprintf(w, "%s\t%s\t%s\t%v\n", vmi.Kind, vmi.Name, vmi.Namespace, vmi.Status.Phase)
	}
	w.Flush()
}
