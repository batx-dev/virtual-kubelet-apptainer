package provider

import (
	"context"
	"os"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigureNode enables a provider to configure the node object that
// will be used for Kubernetes.
func (p *ApptainerProvider) ConfigureNode(ctx context.Context, node *v1.Node) {
	node.Status.Capacity = p.capacity()
	node.Status.Allocatable = p.capacity()
	node.Status.Conditions = p.nodeConditions()
	node.Status.Addresses = p.nodeAddresses()
	node.Status.DaemonEndpoints = p.nodeDaemonEndpoints()
	node.Status.NodeInfo.OperatingSystem = p.operatingSystem
	node.ObjectMeta.Labels["alpha.service-controller.kubernetes.io/exclude-balancer"] = "true"
	node.ObjectMeta.Labels["node.kubernetes.io/exclude-from-external-load-balancers"] = "true"

	// report both old and new styles of OS information
	os := strings.ToLower(p.operatingSystem)
	node.ObjectMeta.Labels["beta.kubernetes.io/os"] = os
	node.ObjectMeta.Labels["kubernetes.io/os"] = os
}

// capacity returns a resource list containing the capacity limits set for apptainer.
func (p *ApptainerProvider) capacity() v1.ResourceList {
	resourceList := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse(p.cpu),
		v1.ResourceMemory: resource.MustParse(p.memory),
		v1.ResourcePods:   resource.MustParse(p.pods),
	}

	if p.gpu != "" {
		resourceList[gpuResourceName] = resource.MustParse(p.gpu)
	}

	return resourceList
}

// nodeConditions returns a list of conditions (Ready, OutOfDisk, etc), for updates to the node status
// within Kubernetes.
func (p *ApptainerProvider) nodeConditions() []v1.NodeCondition {
	return []v1.NodeCondition{
		{
			Type:               "Ready",
			Status:             v1.ConditionTrue,
			LastHeartbeatTime:  metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "KubeletReady",
			Message:            "kubelet is ready.",
		},
		{
			Type:               "OutOfDisk",
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "KubeletHasSufficientDisk",
			Message:            "kubelet has sufficient disk space available",
		},
		{
			Type:               "MemoryPressure",
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "KubeletHasSufficientMemory",
			Message:            "kubelet has sufficient memory available",
		},
		{
			Type:               "DiskPressure",
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "KubeletHasNoDiskPressure",
			Message:            "kubelet has no disk pressure",
		},
		{
			Type:               "NetworkUnavailable",
			Status:             v1.ConditionFalse,
			LastHeartbeatTime:  metav1.Now(),
			LastTransitionTime: metav1.Now(),
			Reason:             "RouteCreated",
			Message:            "RouteController created a route",
		},
	}
}

// nodeAddresses returns a list of addresses for the node status
// within Kubernetes.
func (p *ApptainerProvider) nodeAddresses() []v1.NodeAddress {
	return []v1.NodeAddress{
		{
			Type:    "InternalIP",
			Address: p.internalIP,
		},
	}
}

// nodeDaemonEndpoints returns NodeDaemonEndpoints for the node status
// within Kubernetes.
func (p *ApptainerProvider) nodeDaemonEndpoints() v1.NodeDaemonEndpoints {
	return v1.NodeDaemonEndpoints{
		KubeletEndpoint: v1.DaemonEndpoint{
			Port: p.daemonEndpointPort,
		},
	}
}

func (p *ApptainerProvider) setupNodeCapacity(ctx context.Context) error {
	// Set sane defaults for Capacity in case config is not supplied
	p.cpu = "64"
	p.memory = "512Gi"
	p.pods = "640"

	if cpuQuota := os.Getenv("APPTAINER_QUOTA_CPU"); cpuQuota != "" {
		p.cpu = cpuQuota
	}
	if memoryQuota := os.Getenv("APPTAINER_QUOTA_MEMORY"); memoryQuota != "" {
		p.memory = memoryQuota
	}
	if podsQuota := os.Getenv("APPTAINER_QUOTA_POD"); podsQuota != "" {
		p.pods = podsQuota
	}
	if gpuQuota := os.Getenv("APPTAINER_QUOTA_GPU"); gpuQuota != "" {
		p.gpu = gpuQuota
	}

	return nil
}
