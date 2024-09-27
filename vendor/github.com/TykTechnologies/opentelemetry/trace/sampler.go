package trace

import (
	"strings"

	"github.com/TykTechnologies/opentelemetry/config"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func getSampler(samplingType string, samplingRate float64, parentBased bool) sdktrace.Sampler {
	samplingType = strings.ToLower(samplingType)

	switch {
	case strings.EqualFold(samplingType, config.ALWAYSON):
		if parentBased {
			return sdktrace.ParentBased(sdktrace.AlwaysSample())
		} else {
			return sdktrace.AlwaysSample()
		}
	case strings.EqualFold(samplingType, config.ALWAYSOFF):
		if parentBased {
			return sdktrace.ParentBased(sdktrace.NeverSample())
		} else {
			return sdktrace.NeverSample()
		}
	case strings.EqualFold(samplingType, config.TRACEIDRATIOBASED):
		if parentBased {
			return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(samplingRate))
		} else {
			return sdktrace.TraceIDRatioBased(samplingRate)
		}
	default:
		// Default to AlwaysOn if no valid sampling type is provided
		return sdktrace.AlwaysSample()
	}
}
