# GitHub API Fetcher

Fetches data from the public GitHub API and ingests it into various sinks (for now InfluxDB only).

To run locally:

    $ go install github.com/mhausenblas/github-api-fetcher
    $ ./github-api-fetcher

To run as a DC/OS service (1.8 or above required):

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
  ]
}
```

Above assumes that the InfluxDB package is installed with the following options:

```json
{
  "storage": {
    "pre_create_database": true,
    "pre_create_database_name": "githuborgs",
    "host_volume_influxdb": "/tmp"
  }
}
```