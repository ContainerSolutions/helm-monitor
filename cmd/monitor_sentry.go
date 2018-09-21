package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/helm/pkg/helm"
)

const monitorSentryDesc = `
This command monitor a release by querying Sentry at a given interval and
take care of rolling back to the previous version if the query return a non-
empty result.

Example:

  $ helm monitor sentry my-release \
      --api-key <SENTRY_API_KEY> \
      --organization my-organization \
      --project my-project \
      --sentry https://sentry-endpoint/ \
      --tag environment=production \
      --tag release=2.0.0 \
      --message 'Error message'

Example with event message matching regular expression:

  $ helm monitor sentry my-release \
      --api-key <SENTRY_API_KEY> \
      --organization my-organization \
      --project my-project \
      --sentry https://sentry-endpoint/ \
      --message 'pointer.+' \
			--regexp

`

type monitorSentryCmd struct {
	name               string
	out                io.Writer
	client             helm.Interface
	sentryAddr         string
	sentryAPIKey       string
	sentryOrganization string
	sentryProject      string
	message            string
	regexp             bool
	tags               []string
}

type tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type sentryEvent struct {
	Message string `json:"message"`
	Tags    []*tag `json:"tags"`
}

func newMonitorSentryCmd(out io.Writer) *cobra.Command {
	m := &monitorSentryCmd{
		out: out,
	}

	cmd := &cobra.Command{
		Use:     "sentry [flags] RELEASE",
		Short:   "query a sentry server",
		Long:    monitorSentryDesc,
		PreRunE: setupConnection,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("This command neeeds 1 argument: release name")
			}

			m.name = args[0]
			m.client = ensureHelmClient(m.client)

			return m.run()
		},
	}

	f := cmd.Flags()
	f.StringVar(&m.sentryAddr, "sentry", "http://localhost:9000", "sentry address")
	f.StringVar(&m.sentryAPIKey, "api-key", "", "sentry api key")
	f.StringVar(&m.sentryOrganization, "organization", "", "sentry organization")
	f.StringVar(&m.sentryProject, "project", "", "sentry project")
	f.StringVar(&m.message, "message", "", "event message to match")
	f.BoolVar(&m.regexp, "regexp", false, "enable regular expression")
	f.StringSliceVar(&m.tags, "tag", []string{}, "tags, ie: --tag release=2.0.0 --tag environment=production")

	cmd.MarkFlagRequired("api-key")
	cmd.MarkFlagRequired("organization")
	cmd.MarkFlagRequired("project")

	return cmd
}

func convertStringToTags(s []string) (tagList []*tag) {
	tagList = []*tag{}
	for _, t := range s {
		a := strings.Split(t, "=")
		if len(a) != 2 {
			debug("Provided tag is malformed, should match pattern key=value, got %s", t)
			continue
		}
		tagList = append(tagList, &tag{
			Key:   a[0],
			Value: a[1],
		})
	}

	return
}

func matchEvents(eventList []*sentryEvent, message string, tagList []*tag, useRegexp bool) (output []*sentryEvent, err error) {
	if message == "" && len(tagList) == 0 {
		return eventList, nil
	}

	var r *regexp.Regexp
	if useRegexp {
		r, err = regexp.Compile(message)
		if err != nil {
			return nil, err
		}
	}

	for _, event := range eventList {
		match := false
		if useRegexp && r.MatchString(event.Message) {
			match = true
		} else if message != "" && event.Message == message {
			match = true
		}

		if match && len(tagList) > 0 && !matchTags(tagList, event.Tags) {
			continue
		}

		if match {
			output = append(output, event)
		}
	}

	return
}

func matchTags(tagList []*tag, matchTagList []*tag) bool {
	matchCount := 0
	for _, matchTag := range matchTagList {
		for _, t := range tagList {
			if matchTag.Key == t.Key && matchTag.Value == t.Value {
				matchCount++
			}
		}
	}
	if matchCount != len(tagList) {
		return false
	}

	return true
}

func (m *monitorSentryCmd) run() error {
	_, err := m.client.ReleaseContent(m.name)
	if err != nil {
		return prettyError(err)
	}

	fmt.Fprintf(m.out, "Monitoring %s...\n", m.name)

	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("GET", m.sentryAddr+"/api/0/projects/"+m.sentryOrganization+"/"+m.sentryProject+"/events/", nil)
	if err != nil {
		return prettyError(err)
	}

	req.Header.Add("Authorization", "Bearer "+m.sentryAPIKey)

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

			var response []*sentryEvent
			err = json.Unmarshal(body, &response)
			if err != nil {
				return prettyError(err)
			}

			debug("Response: %v", response)
			debug("Result count: %d", len(response))

			time.Sleep(30 * time.Second)

			events, err := matchEvents(
				response,
				m.message,
				convertStringToTags(m.tags),
				m.regexp,
			)

			if err != nil {
				return prettyError(err)
			}

			debug("Matched events: %d", len(events))

			if len(events) > int(monitor.expectedResultCount) {
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
