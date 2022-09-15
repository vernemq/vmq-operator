package controllers

import (
	"fmt"

	vernemqv1alpha1 "github.com/vernemq/vmq-operator/api/v1alpha1"
)

func volumeName(name string) string {
	return fmt.Sprintf("%s-db", prefixedName(name))
}

func deploymentName(name string) string {
	return fmt.Sprintf("%s-deployment", prefixedName(name))
}

func prefixedName(name string) string {
	return fmt.Sprintf("%s-%s", vernemqName, name)
}

func getHostname(instance *vernemqv1alpha1.VerneMQ) string {
	clusterName := "" // todo: fix back to instance.ClusterName
	if clusterName == "" {
		clusterName = "cluster.local"
	}
	return fmt.Sprintf("%s.%s.svc.%s", serviceName(instance.Name), instance.Namespace, clusterName)
}
