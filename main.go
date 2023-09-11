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

	// This should be capped at 21, for the purpose of demonstration it is not
	age_days := flag.Int("age", 21, "the maximum age (in days) to consider")
	flag.Parse()

	// This needs to broken into hours to work with the time library; it has no notion of days
	age_hours := *age_days * 24
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
	now := time.Now()
	fmt.Println("Current Time:", now)

	// Days needs to be in terms of miliseconds; to achieve this we multiply by the Hour constant
	// which is the number of miliseconds in an hour
	daysago := time.Duration(age_hours) * time.Hour
	fmt.Printf("Age in hours: %f\n", daysago.Hours())

	for _, machine := range masterMachine.Items {
		fmt.Println("Machine: " + machine.Name)

		machineavailable := now.Sub(machine.GetCreationTimestamp().Time)
		fmt.Println("Time Available:", machineavailable)

		if machineavailable > daysago {
			fmt.Printf("Time Available is more than %d days. This machine would be deleted.\n", *age_days)
		}

		fmt.Println()
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
