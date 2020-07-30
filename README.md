# wakatime_exporter

Prometheus exporter for Wakatime statistics.

You can get your Wakatime API key by visiting https://wakatime.com/api-key

```text
usage: wakatime_exporter --wakatime.api-key=WAKATIME.API-KEY [<flags>]

Flags:
  -h, --help                 Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9101"
                             Address to listen on for web interface and telemetry.
      --web.telemetry-path="/metrics"
                             Path under which to expose metrics.
      --wakatime.scrape-uri="https://wakatime.com/api/v1/users/current"
                             Path to query for data.
      --wakatime.api-key=WAKATIME.API-KEY
                             Token for trying to get stats from wakatime.
      --wakatime.timeout=5s  Timeout for trying to get stats from wakatime.
      --wakatime.ssl-verify  Flag that enables SSL certificate verification for the scrape URI
      --log.level=info       Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt    Output format of log messages. One of: [logfmt, json]
      --version              Show application version.
```
