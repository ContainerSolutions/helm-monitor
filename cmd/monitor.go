package main

import (
	"errors"
	"io"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"k8s.io/helm/pkg/helm"
	helm_env "k8s.io/helm/pkg/helm/environment"
)

var (
	settings helm_env.EnvSettings
)

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

func newMonitorCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor prometheus|elasticsearch",
		Short: "monitor a release",
		Long:  monitorDesc,
	}

	cmd.AddCommand(newMonitorPrometheusCmd(nil, out))
	cmd.AddCommand(newMonitorElasticSearchCmd(nil, out))

	return cmd
}

func main() {
	cmd := newMonitorCmd(os.Stdout)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
