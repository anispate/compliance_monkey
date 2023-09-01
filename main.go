package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	machineapi "github.com/openshift/api/machine/v1beta1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	s := runtime.NewScheme()
	utilruntime.Must(corev1.AddToScheme(s))
	utilruntime.Must(machineapi.AddToScheme(s))

	c, err := client.New(config, client.Options{
		Scheme: s,
	})
	if err != nil {
		panic(err)
	}
	masterMachine, err := GetMasterMachines(c)

	if err != nil {
		panic(err)
	}

	// fmt.Println(masterMachine)

	for _, machine := range masterMachine.Items {
		fmt.Println(machine.Name)
		now := time.Now()
		// 504 hours is 21 days
		// var daysago time.Duration = time.Duration(time.Duration.Hours(504))
		daysago := 504 * time.Hour
		fmt.Println(daysago.Hours())

		fmt.Println("now:", now)
		machineavailable := now.Sub(machine.GetCreationTimestamp().Time)
		fmt.Println("TimemachineAvaialbe:", machineavailable)

		if machineavailable > daysago {
			fmt.Println("TimemachineAvaialbe is more than 21 days")
		}
	}

}

const masterMachineLabel string = "machine.openshift.io/cluster-api-machine-role"

func GetMasterMachines(kclient client.Client) (*machineapi.MachineList, error) {
	machineList := &machineapi.MachineList{}
	listOptions := []client.ListOption{
		client.InNamespace("openshift-machine-api"),
		client.MatchingLabels{masterMachineLabel: "worker"},
	}
	err := kclient.List(context.TODO(), machineList, listOptions...)
	if err != nil {
		return nil, err
	}
	return machineList, nil
}
