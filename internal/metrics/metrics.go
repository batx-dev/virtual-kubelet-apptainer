package metrics

import (
	"context"

	stats "github.com/virtual-kubelet/virtual-kubelet/node/api/statsv1alpha1"
)

type ApptaienrMetricsProvider struct {
}

func NewApptaienrMetricsProver() *ApptaienrMetricsProvider {
	p := &ApptaienrMetricsProvider{}
	return p
}

// GetStatsSummary returns the stats summary for pods running on ACI
func (p *ApptaienrMetricsProvider) GetStatsSummary(ctx context.Context) (summary *stats.Summary, err error) {
	s := &stats.Summary{}
	return s, nil
}
