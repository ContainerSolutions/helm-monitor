package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"k8s.io/helm/pkg/helm"
	helm_env "k8s.io/helm/pkg/helm/environment"
)

var (
	settings helm_env.EnvSettings
	monitor  *monitorCmd
)

type monitorCmd struct {
	disableHooks    bool
	dryRun          bool
	force           bool
	interval        int64
	rollbackTimeout int64
	timeout         int64
	verbose         bool
	wait            bool
}

const monitorDesc = `
This command monitor a release by querying Prometheus or Elasticsearch at a
given interval and take care of rolling back to the previous version if the
query return a non-empty result.
`

func setupConnection(c *cobra.Command, args []string) error {
	settings.TillerHost = os.Getenv("TILLER_HOST")
	return nil
}

func ensureHelmClient(h helm.Interface) helm.Interface {
	if h != nil {
		return h
	}

	return helm.NewClient(helm.Host(settings.TillerHost))
}

func prettyError(err error) error {
	if err == nil {
		return nil
	}
	return errors.New(grpc.ErrorDesc(err))
}

func debug(format string, args ...interface{}) {
	if monitor.verbose {
		format = fmt.Sprintf("[debug] %s\n", format)
		fmt.Printf(format, args...)
	}
}

func newMonitorCmd(out io.Writer) *cobra.Command {
	monitor = &monitorCmd{}

	cmd := &cobra.Command{
		Use:   "monitor prometheus|elasticsearch",
		Short: "monitor a release",
		Long:  monitorDesc,
	}

	p := cmd.PersistentFlags()
	p.BoolVar(&monitor.disableHooks, "no-hooks", false, "prevent hooks from running during rollback")
	p.BoolVar(&monitor.dryRun, "dry-run", false, "simulate a rollback if triggered by query result")
	p.BoolVar(&monitor.force, "force", false, "force resource update through delete/recreate if needed")
	p.BoolVar(&monitor.wait, "wait", false, "if set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment are in a ready state before marking a rollback as successful. It will wait for as long as --rollback-timeout")
	p.BoolVarP(&monitor.verbose, "verbose", "v", false, "enable verbose output")
	p.Int64Var(&monitor.rollbackTimeout, "rollback-timeout", 300, "time in seconds to wait for any individual Kubernetes operation during the rollback (like Jobs for hooks)")
	p.Int64Var(&monitor.timeout, "timeout", 300, "time in seconds to wait before assuming a monitoring action is successfull")
	p.Int64VarP(&monitor.interval, "interval", "i", 10, "time in seconds between each query")

	cmd.AddCommand(
		newMonitorPrometheusCmd(out),
		newMonitorElasticSearchCmd(out),
	)

	return cmd
}

func main() {
	cmd := newMonitorCmd(os.Stdout)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
