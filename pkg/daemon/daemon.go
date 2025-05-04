// Package daemon create an entrypoint for daemon
package daemon

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"

	"github.com/transform-ia/mcp-tools/pkg/telemetry"
)

func runDaemon(serviceName, version string, logic func() error) error {
	ctx := context.Background()

	shutdown, err := telemetry.InitTelemetry(ctx, serviceName, version)
	if err != nil {
		return errors.Wrap(err, "telemetry.InitTelemetry")
	}

	if err = logic(); err != nil {
		sErr := shutdown(ctx)
		if sErr != nil {
			fmt.Println(sErr.Error())
		}

		return errors.Wrap(err, "logic")
	}

	if err = shutdown(ctx); err != nil {
		return errors.Wrap(err, "shutdown")
	}

	return nil
}

// RunDaemon is entrypoint for a deamon
func RunDaemon(serviceName, version string, logic func() error) {
	if err := runDaemon(serviceName, version, logic); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
