# wakatime_exporter

Prometheus exporter for Wakatime statistics.

[Click here](METRICS.md) to see an example of the exported metrics.

<a href="#"><img src="example.png"></a>


## Usage

You can get your Wakatime API key by visiting: https://wakatime.com/api-key

```text
usage: wakatime_exporter --wakatime.api-key=WAKATIME.API-KEY [<flags>]

Flags:
  -h, --help                  Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9212"
                              Address to listen on for web interface and telemetry.
      --web.telemetry-path="/metrics"
                              Path under which to expose metrics.
      --wakatime.scrape-uri="https://wakatime.com/api/v1/users/current"
                              Base path to query for Wakatime data.
      --wakatime.api-key=WAKATIME.API-KEY
                              Token to use when getting stats from Wakatime.
      --wakatime.t-offset=0s  Time offset (from UTC) for managing Wakatime dates.
      --wakatime.timeout=5s   Timeout for trying to get stats from Wakatime.
      --wakatime.ssl-verify   Flag that enables SSL certificate verification for the scrape URI
      --log.level=info        Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt     Output format of log messages. One of: [logfmt, json]
      --version               Show application version.
```

## Docker

```shell
docker run -p 9212:9212 macropower/wakatime-exporter:0.0.3 --wakatime.api-key="YOUR_API_KEY"
```

## Time zones

Wakatime will use whatever timezone you have set in your preferences to choose what date to append new metrics to. For instance, a timezone of America/New_York results in the local date changing at 4AM UTC. This exporter, by default, will begin to query the next date at 12AM UTC. This will lead to innacuracies in the data, as the final hours (4 hours, in this case) will not be reported.

**Thus, you must take _one_ of the following two actions to recieve correct data:**
- Set the `--wakatime.t-offset` parameter to adjust when the exporter begins querying the new date. For instance, since America/New_York is UTC−04:00, you can supply `-4h` to obtain correct results. This parameter accepts both positive and negative values.
- Change your timezone in your Wakatime preferences to `UTC` at: https://wakatime.com/settings/preferences
