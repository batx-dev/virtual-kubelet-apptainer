package provider

import (
	"context"
	"io"

	"github.com/batx-dev/virtual-kubelet-apptainer/internal/metrics"
	"github.com/virtual-kubelet/virtual-kubelet/node/api"
	corev1 "k8s.io/api/core/v1"
)

const (
	gpuResourceName = "nvidia.com/gpu"
)

type ApptainerProvider struct {
	nodeName           string
	operatingSystem    string
	cpu                string
	memory             string
	pods               string
	gpu                string
	internalIP         string
	daemonEndpointPort int32

	*metrics.ApptaienrMetricsProvider
}

func NewApptainerProvider(ctx context.Context, nodeName, operatingSystem, internalIP string,
	daemonEndpointPort int32) (*ApptainerProvider, error) {
	p := &ApptainerProvider{}

	p.nodeName = nodeName
	p.operatingSystem = operatingSystem
	p.internalIP = internalIP
	p.daemonEndpointPort = daemonEndpointPort

	if err := p.setupNodeCapacity(ctx); err != nil {
		return nil, err
	}

	p.ApptaienrMetricsProvider = metrics.NewApptaienrMetricsProver()
	return p, nil
}

// CreatePod takes a Kubernetes Pod and deploys it within the provider.
func (p *ApptainerProvider) CreatePod(ctx context.Context, pod *corev1.Pod) error {
	return nil
}

// UpdatePod takes a Kubernetes Pod and updates it within the provider.
func (p *ApptainerProvider) UpdatePod(ctx context.Context, pod *corev1.Pod) error {
	return nil
}

// DeletePod takes a Kubernetes Pod and deletes it from the provider. Once a pod is deleted, the provider is
// expected to call the NotifyPods callback with a terminal pod status where all the containers are in a terminal
// state, as well as the pod. DeletePod may be called multiple times for the same pod.
func (p *ApptainerProvider) DeletePod(ctx context.Context, pod *corev1.Pod) error {
	return nil
}

// GetPod retrieves a pod by name from the provider (can be cached).
// The Pod returned is expected to be immutable, and may be accessed
// concurrently outside of the calling goroutine. Therefore it is recommended
// to return a version after DeepCopy.
func (p *ApptainerProvider) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	return nil, nil
}

// GetPodStatus retrieves the status of a pod by name from the provider.
// The PodStatus returned is expected to be immutable, and may be accessed
// concurrently outside of the calling goroutine. Therefore it is recommended
// to return a version after DeepCopy.
func (p *ApptainerProvider) GetPodStatus(ctx context.Context, namespace, name string) (*corev1.PodStatus, error) {
	return nil, nil
}

// GetPods retrieves a list of all pods running on the provider (can be cached).
// The Pods returned are expected to be immutable, and may be accessed
// concurrently outside of the calling goroutine. Therefore it is recommended
// to return a version after DeepCopy.
func (p *ApptainerProvider) GetPods(context.Context) ([]*corev1.Pod, error) {
	return nil, nil
}

// GetContainerLogs retrieves the logs of a container by name from the provider.
func (p *ApptainerProvider) GetContainerLogs(ctx context.Context, namespace, podName, containerName string, opts api.ContainerLogOpts) (io.ReadCloser, error) {
	return nil, nil
}

// RunInContainer executes a command in a container in the pod, copying data
// between in/out/err and the container's stdin/stdout/stderr.
func (p *ApptainerProvider) RunInContainer(ctx context.Context, namespace, podName, containerName string, cmd []string, attach api.AttachIO) error {
	return nil
}
