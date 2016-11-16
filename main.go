// Copyright 2016 Mesosphere. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The GitGub API Fetcher (GAF) is a simple app server that fetches event
// data on a certain GitHub organization defaulting to 'dcos' and that you
// can overwrite with an environment variable GITHUB_TARGET_ORG. GAF then
// ingests the events into InfluxDB. By default GAF serves on port 9393
// and this can be overwritten using the env variable PORT0. The fetch and
// ingest process may be started at any time using the host:port/start endpoint
// and the time to wait between two fetches can be defined vua env variable
// FETCH_WAIT_SEC, which defaults to 10 seconds.
package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
	"github.com/influxdata/influxdb/client/v2"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	VERSION    string = "0.1.0"
	INFLUX_API string = "http://influxdb.marathon.l4lb.thisdcos.directory:8086"
)

var (
	mux         *http.ServeMux
	ghc         *github.Client
	serviceport string
	// which GitHub org to fetch events from (defaults to 'dcos'):
	targetorg string
	// how many seconds to wait between fetching events:
	fetchwaitsec time.Duration
	// into which InfluxDB database to ingest events:
	targetdb string
)

func init() {
	serviceport = "9393"
	if sp := os.Getenv("PORT0"); sp != "" {
		serviceport = sp
	}
	targetorg = "dcos"
	if to := os.Getenv("GITHUB_TARGET_ORG"); to != "" {
		targetorg = to
	}
	fetchwaitsec = 10
	if fw := os.Getenv("FETCH_WAIT_SEC"); fw != "" {
		if fwi, err := strconv.Atoi(fw); err == nil {
			fetchwaitsec = time.Duration(fwi)
		}
	}
	targetdb = "githuborgs"
	if td := os.Getenv("INFLUX_TARGET_DB"); td != "" {
		targetdb = td
	}
	ghc = github.NewClient(nil)
}

// write batch-ingests the events from GitHub into InfluxDB
func write(c client.Client, events []github.Event) {
	if bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  targetdb,
		Precision: "s", // second resultion is fine for our purpose
	}); err != nil {
		log.WithFields(log.Fields{"func": "writePoints"}).Error(err)
	} else {
		log.WithFields(log.Fields{"func": "writePoints"}).Info("Connected to ", fmt.Sprintf("%+v", c))
		for _, event := range events {
			tags := map[string]string{
				"repo":   *event.Repo.Name,
				"action": *event.Type,
				"actor":  *event.Actor.Login,
			}
			fields := map[string]interface{}{
				"count": 1,
			}
			if pt, err := client.NewPoint("event", tags, fields, time.Now()); err != nil {
				log.WithFields(log.Fields{"func": "writePoints"}).Error(err)
			} else {
				bp.AddPoint(pt)
				log.WithFields(log.Fields{"func": "writePoints"}).Info("Added point ", fmt.Sprintf("%+v", pt))
			}
		}
		if err := c.Write(bp); err != nil {
			log.WithFields(log.Fields{"func": "writePoints"}).Error(err)
		} else {
			log.WithFields(log.Fields{"func": "writePoints"}).Info("Added and written all points in this batch")
		}
	}
}

// ingest2Influx creates a connection to InfluxDB and ingests the GitHub events
// of the targeted org that occured in the last reporting period.
func ingest2Influx(events []github.Event) {
	log.WithFields(log.Fields{"func": "ingest2Influx"}).Info("Trying to ingest data into InfluxDB")
	if c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     INFLUX_API,
		Username: "root",
		Password: "root",
	}); err != nil {
		log.WithFields(log.Fields{"func": "ingest2Influx"}).Error(err)
	} else {
		write(c, events)
	}
}

// fetch gets the events for the targeted org via the public GitHub API
func fetch() ([]github.Event, error) {
	log.WithFields(log.Fields{"func": "fetch"}).Info("Trying to fetch events from org ", targetorg)
	if events, _, err := ghc.Activity.ListEventsForOrganization(targetorg, nil); err != nil {
		log.WithFields(log.Fields{"func": "fetch"}).Error(err)
		return nil, err
	} else {
		log.WithFields(log.Fields{"func": "fetch"}).Info("Successfully fetched events")
		return events, nil
	}
}

// ingest tries to fetch event data of the targeted org from GitHub
// and if successful ingests it into InfluxDB.
func ingest() {
	for {
		if events, err := fetch(); err == nil {
			ingest2Influx(events)
		}
		time.Sleep(fetchwaitsec * time.Second)
	}
}

func main() {
	mux = http.NewServeMux()
	fmt.Printf("This is the GitHub API Fetcher in version %s listening on port %s\n", VERSION, serviceport)
	// the /start endpoint kicks off the fetch and ingest process in a separate Go routine:
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{"handle": "/start"}).Info("Starting to fetch data from GitHub")
		go ingest()
	})
	log.Fatal(http.ListenAndServe(":"+serviceport, mux))
}
