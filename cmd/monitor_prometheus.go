package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/helm/pkg/helm"
)

const monitorPrometheusDesc = `
This command monitor a release by querying Prometheus at a given interval and
take care of rolling back to the previous version if the query return a non-
empty result.

Usage:

	$ helm monitor prometheus frontend 'rate(http_requests_total{status=~"5.."}[1m]) > 0'


Reference:

	https://prometheus.io/docs/prometheus/latest/querying/basics/

`

type monitorPrometheusCmd struct {
	name            string
	out             io.Writer
	client          helm.Interface
	timeout         int64
	rollbackTimeout int64
	interval        int64
	prometheusAddr  string
	query           string
	dryRun          bool
	wait            bool
	force           bool
	disableHooks    bool
}

type prometheusQueryResponse struct {
	Data struct {
		Result []struct{} `json:"result"`
	} `json:"data"`
}

func newMonitorPrometheusCmd(client helm.Interface, out io.Writer) *cobra.Command {
	prometheusMonitor := &monitorPrometheusCmd{
		out:    out,
		client: client,
	}

	cmd := &cobra.Command{
		Use:     "prometheus [flags] RELEASE PROMQL",
		Short:   "query a prometheus server",
		Long:    monitorPrometheusDesc,
		PreRunE: setupConnection,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("This command neeeds 2 argument: release name, promql")
			}

			prometheusMonitor.name = args[0]
			prometheusMonitor.query = args[1]
			prometheusMonitor.client = ensureHelmClient(prometheusMonitor.client)

			return prometheusMonitor.run()
		},
	}

	f := cmd.Flags()
	f.BoolVar(&prometheusMonitor.dryRun, "dry-run", false, "simulate a monitoring")
	f.Int64Var(&prometheusMonitor.timeout, "timeout", 300, "time in seconds to wait before assuming a monitoring action is successfull")
	f.Int64Var(&prometheusMonitor.rollbackTimeout, "rollback-timeout", 300, "time in seconds to wait for any individual Kubernetes operation during the rollback (like Jobs for hooks)")
	f.Int64Var(&prometheusMonitor.interval, "interval", 10, "time in seconds between each query")
	f.BoolVar(&prometheusMonitor.wait, "wait", false, "if set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment are in a ready state before marking a rollback as successful. It will wait for as long as --rollback-timeout")
	f.BoolVar(&prometheusMonitor.force, "force", false, "force resource update through delete/recreate if needed")
	f.BoolVar(&prometheusMonitor.disableHooks, "no-hooks", false, "prevent hooks from running during rollback")
	f.StringVar(&prometheusMonitor.prometheusAddr, "prometheus", "http://localhost:9090", "prometheus address")

	return cmd
}

func (m *monitorPrometheusCmd) run() error {
	_, err := m.client.ReleaseContent(m.name)
	if err != nil {
		return prettyError(err)
	}

	fmt.Fprintf(m.out, "Monitoring %s...\n", m.name)

	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("GET", m.prometheusAddr+"/api/v1/query", nil)
	if err != nil {
		return prettyError(err)
	}

	q := req.URL.Query()
	q.Add("query", m.query)
	req.URL.RawQuery = q.Encode()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	ticker := time.NewTicker(time.Second * time.Duration(m.interval))

	go func() {
		time.Sleep(time.Second * time.Duration(m.timeout))
		fmt.Fprintf(m.out, "No results after %d second(s)\n", m.timeout)
		close(quit)
	}()

	for {
		select {
		case <-ticker.C:
			res, err := client.Do(req)
			if err != nil {
				return prettyError(err)
			}

			defer res.Body.Close()

			body, err := ioutil.ReadAll(res.Body)

			if err != nil {
				return prettyError(err)
			}

			response := &prometheusQueryResponse{}
			err = json.Unmarshal(body, response)
			if err != nil {
				return prettyError(err)
			}

			if len(response.Data.Result) > 0 {
				ticker.Stop()

				fmt.Fprintf(m.out, "Failure detected, rolling back...\n")

				_, err := m.client.RollbackRelease(
					m.name,
					helm.RollbackDryRun(m.dryRun),
					helm.RollbackRecreate(false),
					helm.RollbackForce(m.force),
					helm.RollbackDisableHooks(m.disableHooks),
					helm.RollbackVersion(0),
					helm.RollbackTimeout(m.rollbackTimeout),
					helm.RollbackWait(m.wait))

				if err != nil {
					return prettyError(err)
				}

				fmt.Fprintf(m.out, "Successfully rolled back to previous revision!\n")
				return nil
			}

		case <-quit:
			ticker.Stop()
			return nil
		}
	}
}
