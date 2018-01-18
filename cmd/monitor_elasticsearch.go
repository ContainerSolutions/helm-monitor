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

const monitorElasticSearchDesc = `
This command monitor a release by querying ElasticSearch at a given interval
and take care of rolling back to the previous version if the query return a non-
empty result.

The query argument can be either the path of a query DSL json file or a Lucene
query string.

Usage with Lucene query:

	$ helm monitor elasticsearch frontend 'status:500 AND kubernetes.labels.app:app AND version:2.0.0'

Usage with query DSL file:

	$ helm monitor elasticsearch frontend ./examples/elasticsearch-query.json


Reference:

	https://www.elastic.co/guide/en/elasticsearch/reference/current/search-count.html

`

type monitorElasticSearchCmd struct {
	name              string
	out               io.Writer
	client            helm.Interface
	timeout           int64
	rollbackTimeout   int64
	interval          int64
	elasticSearchAddr string
	query             string
	dryRun            bool
	wait              bool
	force             bool
	disableHooks      bool
}

type elasticSearchQueryResponse struct {
	Count int64 `json:"count"`
}

func newMonitorElasticSearchCmd(client helm.Interface, out io.Writer) *cobra.Command {
	elasticSearchMonitor := &monitorElasticSearchCmd{
		out:    out,
		client: client,
	}

	cmd := &cobra.Command{
		Use:     "elasticsearch [flags] RELEASE [QUERY DSL PATH|LUCENE QUERY]",
		Short:   "query an elasticsearch server",
		Long:    monitorElasticSearchDesc,
		PreRunE: setupConnection,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("This command neeeds 2 argument: release name, query DSL path or Lucene query")
			}

			elasticSearchMonitor.name = args[0]
			elasticSearchMonitor.query = args[1]
			elasticSearchMonitor.client = ensureHelmClient(elasticSearchMonitor.client)

			return elasticSearchMonitor.run()
		},
	}

	f := cmd.Flags()
	f.BoolVar(&elasticSearchMonitor.dryRun, "dry-run", false, "simulate a monitoring")
	f.Int64Var(&elasticSearchMonitor.timeout, "timeout", 300, "time in seconds to wait before assuming a monitoring action is successfull")
	f.Int64Var(&elasticSearchMonitor.rollbackTimeout, "rollback-timeout", 300, "time in seconds to wait for any individual Kubernetes operation during the rollback (like Jobs for hooks)")
	f.Int64Var(&elasticSearchMonitor.interval, "interval", 10, "time in seconds between each query")
	f.BoolVar(&elasticSearchMonitor.wait, "wait", false, "if set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment are in a ready state before marking a rollback as successful. It will wait for as long as --rollback-timeout")
	f.BoolVar(&elasticSearchMonitor.force, "force", false, "force resource update through delete/recreate if needed")
	f.BoolVar(&elasticSearchMonitor.disableHooks, "no-hooks", false, "prevent hooks from running during rollback")
	f.StringVar(&elasticSearchMonitor.elasticSearchAddr, "elasticsearch", "http://localhost:9200", "elasticsearch address")

	return cmd
}

func (m *monitorElasticSearchCmd) run() error {
	_, err := m.client.ReleaseContent(m.name)
	if err != nil {
		return prettyError(err)
	}

	fmt.Fprintf(m.out, "Monitoring %s...\n", m.name)

	client := &http.Client{Timeout: 10 * time.Second}

	queryBody, err := os.Open(m.query)

	var req *http.Request
	if err != nil {
		req, err = http.NewRequest("GET", m.elasticSearchAddr+"/_count", nil)
		if err != nil {
			return prettyError(err)
		}

		q := req.URL.Query()
		q.Add("q", m.query)
		req.URL.RawQuery = q.Encode()
	} else {
		req, err = http.NewRequest("GET", m.elasticSearchAddr+"/_count", queryBody)
		if err != nil {
			return prettyError(err)
		}
		req.Header.Set("Content-Type", "application/json")
	}

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

			fmt.Printf("Body: %s\n", string(body))

			response := &elasticSearchQueryResponse{}
			err = json.Unmarshal(body, response)
			if err != nil {
				return prettyError(err)
			}

			if response.Count > 0 {
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
