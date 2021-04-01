package newreilcmetric

import "github.com/newrelic/go-agent/v3/newrelic"

func NewApp(name string, license string) *newrelic.Application {
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(name),
		newrelic.ConfigLicense(license),
		newrelic.ConfigDistributedTracerEnabled(true),
	)
	if err != nil {
		panic(err)
	}

	return app
}
