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

	$ helm monitor prometheus frontend 'rate(http_requests_total{code=~"^5.*$"}[5m]) > 0'


Reference:

	https://prometheus.io/docs/prometheus/latest/querying/basics/

`

type monitorPrometheusCmd struct {
	name           string
	out            io.Writer
	client         helm.Interface
	prometheusAddr string
	query          string
}

type prometheusQueryResponse struct {
	Data struct {
		Result []struct{} `json:"result"`
	} `json:"data"`
}

func newMonitorPrometheusCmd(out io.Writer) *cobra.Command {
	m := &monitorPrometheusCmd{
		out: out,
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

			m.name = args[0]
			m.query = args[1]
			m.client = ensureHelmClient(m.client)

			return m.run()
		},
	}

	f := cmd.Flags()
	f.StringVar(&m.prometheusAddr, "prometheus", "http://localhost:9090", "prometheus address")

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

	ticker := time.NewTicker(time.Second * time.Duration(monitor.interval))

	go func() {
		time.Sleep(time.Second * time.Duration(monitor.timeout))
		fmt.Fprintf(m.out, "No results after %d second(s)\n", monitor.timeout)
		close(quit)
	}()

	for {
		select {
		case <-ticker.C:
			debug("Processing URL %s", req.URL.String())

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

			debug("Response: %v", response)
			debug("Result count: %d", len(response.Data.Result))

			if len(response.Data.Result) > int(monitor.expectedResultCount) {
				ticker.Stop()

				fmt.Fprintf(m.out, "Failure detected, rolling back...\n")

				_, err := m.client.RollbackRelease(
					m.name,
					helm.RollbackDryRun(monitor.dryRun),
					helm.RollbackRecreate(false),
					helm.RollbackForce(monitor.force),
					helm.RollbackDisableHooks(monitor.disableHooks),
					helm.RollbackVersion(0),
					helm.RollbackTimeout(monitor.rollbackTimeout),
					helm.RollbackWait(monitor.wait))

				if err != nil {
					return prettyError(err)
				}

				fmt.Fprintf(m.out, "Successfully rolled back to previous revision!\n")
				return nil
			}

		case <-quit:
			ticker.Stop()
			debug("Quitting...")
			return nil
		}
	}
}
