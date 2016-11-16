# GitHub API Fetcher

Fetches data from the public GitHub API and ingests it into various sinks (for now InfluxDB only).

## To run locally

    $ go install github.com/mhausenblas/github-api-fetcher
    $ ./github-api-fetcher
    $ curl localhost:9393/start

## To run as a DC/OS service 

This requires a DC/OS 1.8 or above cluster:

```json
{
  "id": "/fetcher",
  "cpus": 0.5,
  "mem": 300,
  "cmd": "curl -s -L https://github.com/mhausenblas/github-api-fetcher/releases/download/0.1.0/github-api-fetcher -o gaf && chmod u+x gaf && ./gaf",
  "portDefinitions": [
    {
      "labels": {
        "VIP_0": "/fetcher:80"
      }
    }
  ],
  "env": {
    "GITHUB_TARGET_ORG": "dcos",
    "FETCH_WAIT_SEC": "60",
    "INFLUX_TARGET_DB": "githuborgs"
  }
}
```

Note that all env variable shown here are the default and above assumes that the InfluxDB package is installed (`dcos package install --options=influx-config.json influxdb`) with the following options:

```json
{
  "storage": {
    "pre_create_database": true,
    "pre_create_database_name": "githuborgs",
    "host_volume_influxdb": "/tmp"
  }
}
```

Now you can kick off fetch & ingest like so (from within the cluster):

    $ curl fetcher.marathon.l4lb.thisdcos.directory:80/start