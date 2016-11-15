package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	// "github.com/google/go-github/github"
	"github.com/influxdata/influxdb/client/v2"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	VERSION    string = "0.1.0"
	GITHUB_API string = "https://api.github.com/"
	INFLUX_API string = "http://influxdb.marathon.l4lb.thisdcos.directory:8086"
)

var (
	mux         *http.ServeMux
	serviceport string
)

func init() {
	serviceport = "9393"
	if sp := os.Getenv("PORT0"); sp != "" {
		serviceport = sp
	}
}

func writePoints(c client.Client) {
	if bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "githuborgs",
		Precision: "s",
	}); err != nil {
		log.WithFields(log.Fields{"func": "writePoints"}).Error(err)
	} else {
		log.WithFields(log.Fields{"func": "writePoints"}).Info("Connected to ", c)
		for i := 0; i < 10; i++ {
			tags := map[string]string{
				"action": "SOMEACTION",
				"actor":  "SOMEUSER",
			}
			fields := map[string]interface{}{
				"repo":  "SOMEREPO",
				"count": rand.Float64() * 100.0,
			}

			if pt, err := client.NewPoint("activity", tags, fields, time.Now()); err != nil {
			} else {
				bp.AddPoint(pt)
			}
		}
		if err := c.Write(bp); err != nil {
			log.WithFields(log.Fields{"func": "writePoints"}).Error(err)
		} else {
			log.WithFields(log.Fields{"func": "writePoints"}).Info("Written ", bp)
		}
	}
}

func ingest2Influx() {
	log.WithFields(log.Fields{"func": "ingest2Influx"}).Info("Starting to ingest data into InfluxDB")

	if c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     INFLUX_API,
		Username: "root",
		Password: "root",
	}); err != nil {
		log.WithFields(log.Fields{"func": "ingest2Influx"}).Error(err)
	} else {
		for {
			writePoints(c)
			time.Sleep(5 * time.Second)
		}
	}
}

func main() {
	mux = http.NewServeMux()
	fmt.Printf("This is the GitHub API Fetcher in version %s listening on port %s\n", VERSION, serviceport)
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{"handle": "/start"}).Info("Starting ingest process from ", GITHUB_API)
		go ingest2Influx()
	})
	log.Fatal(http.ListenAndServe(":"+serviceport, mux))
}
