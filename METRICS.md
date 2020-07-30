```text
# HELP wakatime_editor_seconds_total Total seconds for each editor
# TYPE wakatime_editor_seconds_total counter
wakatime_editor_seconds_total{name="VS Code"} 14597.149032

# HELP wakatime_exporter_build_info A metric with a constant '1' value labeled by version, revision, branch, and goversion from which wakatime_exporter was built.
# TYPE wakatime_exporter_build_info gauge
wakatime_exporter_build_info{branch="master",goversion="go1.14.6",revision="d2c50eec79c0899ca397db86cc6b108a36cda328",version="0.0.1"} 1

# HELP wakatime_exporter_query_failures_total Number of errors.
# TYPE wakatime_exporter_query_failures_total counter
wakatime_exporter_query_failures_total 0

# HELP wakatime_exporter_scrapes_total Current total wakatime scrapes.
# TYPE wakatime_exporter_scrapes_total counter
wakatime_exporter_scrapes_total 1

# HELP wakatime_exporter_up Was the last scrape of wakatime successful.
# TYPE wakatime_exporter_up gauge
wakatime_exporter_up 1

# HELP wakatime_language_seconds_total Total seconds for each language
# TYPE wakatime_language_seconds_total counter
wakatime_language_seconds_total{name="Docker"} 1111.670769
wakatime_language_seconds_total{name="Git Config"} 102.067642
wakatime_language_seconds_total{name="Go"} 8437.99026
wakatime_language_seconds_total{name="JSON"} 121.212324
wakatime_language_seconds_total{name="Jsonnet"} 650.405094
wakatime_language_seconds_total{name="Makefile"} 1926.656756
wakatime_language_seconds_total{name="Markdown"} 1450.281681
wakatime_language_seconds_total{name="Other"} 103.456884
wakatime_language_seconds_total{name="Text"} 5.495999
wakatime_language_seconds_total{name="YAML"} 687.911623

# HELP wakatime_machine_seconds_total Total seconds for each machine
# TYPE wakatime_machine_seconds_total counter
wakatime_machine_seconds_total{id="61940d85-2733-4cbb-b702-adae5ff355d5",name="DesktopPC"} 13917.740195
wakatime_machine_seconds_total{id="8c8e152d-b735-4725-b08c-5f92f35263fb",name="MacbookPro"} 679.408837

# HELP wakatime_operating_system_seconds_total Total seconds for each operating system
# TYPE wakatime_operating_system_seconds_total counter
wakatime_operating_system_seconds_total{name="Mac"} 679.408837
wakatime_operating_system_seconds_total{name="Windows"} 13917.740195
```
