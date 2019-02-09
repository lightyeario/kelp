package monitoring

import (
	"github.com/interstellar/kelp/api"
	"github.com/interstellar/kelp/support/logger"
)

type noopAlert struct{}

var _ api.Alert = &noopAlert{}

// Trigger is simply a noop for the default Alert, meaning that the client
// hasn't specified a monitoring service that's supported.
func (p *noopAlert) Trigger(description string, details interface{}) error {
	return nil
}

// MakeAlert creates an Alert based on the type of the service (eg Pager Duty) and its corresponding API key.
func MakeAlert(alertType string, apiKey string, l logger.Logger) (api.Alert, error) {
	switch alertType {
	case "PagerDuty":
		return makePagerDuty(apiKey, l)
	default:
		return &noopAlert{}, nil
	}
}
