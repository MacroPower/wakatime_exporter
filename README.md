# wakatime_exporter

<a href="#" target="blank">
  <img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/go-version/MacroPower/wakatime_exporter">
</a>
<a href="https://hub.docker.com/r/macropower/wakatime-exporter" target="blank">
  <img alt="Docker Image Size (latest by date)" src="https://img.shields.io/docker/image-size/macropower/wakatime-exporter?color=green">
</a>
<a href="https://hub.docker.com/r/macropower/wakatime-exporter" target="blank">
  <img alt="Docker Pulls" src="https://img.shields.io/docker/pulls/macropower/wakatime-exporter">
</a>
<a href="https://codeclimate.com/github/MacroPower/wakatime_exporter/maintainability">
  <img src="https://api.codeclimate.com/v1/badges/ed191a2b4937b9f87096/maintainability">
</a>

<br>

Prometheus exporter for Wakatime statistics.

[Click here](METRICS.md) to see an example of the exported metrics.

[Click here](https://grafana.com/grafana/dashboards/12790) for a simple dashboard you can use to get started.

<a href="#"><img src="example.png"></a>

## Usage

In most cases, you should only need to provide an API key and a time offset (see below for details).
All other parameters are for advanced use-cases only and you should be able to leave them as their defaults.

You can get your Wakatime API key by visiting: https://wakatime.com/api-key

Provide arguments via parameters:

```text
usage: wakatime_exporter --wakatime.api-key=WAKATIME.API-KEY [<flags>]

Flags:
  --help                               Show context-sensitive help.
  --web.listen-address=":9212"         Address to listen on for web interface and telemetry.
  --web.telemetry-path="/metrics"      Path under which to expose metrics.
  --wakatime.scrape-uri="https://wakatime.com/api/v1"
                                       Base path to query for Wakatime data.
  --wakatime.user="current"            User to query for Wakatime data.
  --wakatime.api-key=WAKATIME.API-KEY  Token to use when getting stats from Wakatime.
  --wakatime.t-offset=0s               Time offset (from UTC) for managing Wakatime dates.
  --wakatime.timeout=5s                Timeout for trying to get stats from Wakatime.
  --wakatime.ssl-verify                Flag that enables SSL certificate verification.
  --log.level=info                     Only log messages with the given severity or above.
                                       One of: [debug, info, warn, error]
  --log.format=logfmt                  Output format of log messages.
                                       One of: [logfmt, json]
  --version                            Show application version.
```

and/or via environment variables:

```
WAKA_LISTEN_ADDR=":9212"                      # Address to listen on for web interface and telemetry.
WAKA_METRICS_PATH="/metrics"                  # Path under which to expose metrics.
WAKA_SCRAPE_URI="https://wakatime.com/api/v1" # Base path to query for Wakatime data.
WAKA_USER="current"                           # User to query for Wakatime data.
WAKA_API_KEY=""                               # Token to use when getting stats from Wakatime.
WAKA_TIME_OFFSET="0s"                         # Time offset (from UTC) for managing Wakatime dates.
WAKA_TIMEOUT="5s"                             # Timeout for trying to get stats from Wakatime.
WAKA_SSL_VERIFY="true"                        # SSL certificate verification for the scrape URI.
```

## Docker

```shell
docker run -p 9212:9212 macropower/wakatime-exporter:0.0.4 --wakatime.api-key="YOUR_API_KEY"
```

## Time zones

Wakatime will use whatever timezone you have set in your preferences to choose what date to append new metrics to. For instance, a timezone of America/New_York results in the local date changing at 4AM UTC. This exporter, by default, will begin to query the next date at 12AM UTC. This will lead to innacuracies in the data, as the final hours (4 hours, in this case) will not be reported.

**Thus, you must take _one_ of the following two actions to receive correct data:**

- Set the `--wakatime.t-offset` parameter to adjust when the exporter begins querying the new date. For instance, since America/New_York is UTC−04:00, you can supply `-4h` to obtain correct results. This parameter accepts both positive and negative values.
- Change your timezone in your Wakatime preferences to `UTC` at: https://wakatime.com/settings/preferences
