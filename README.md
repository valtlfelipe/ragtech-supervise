# Ragtech Supervise Docker Image with Prometheus metrics

Ragtech is a Brazilian company that produces UPS (more commonly known in Brazil as no-break)
devices. They have a software called Supervise that is used to monitor and control the UPS devices.

The existing Supervise software packaging is not really suitable with modern Linux distributions.
This project aims to provide a Docker image with Supervise supplying all required dependencies.

Also, a Prometheus metrics endpoint was added to easily export metrics data.

## Usage

You can run a new container from the computer where the UPS USB cable is plugged to. Validate that
the serial interface created by this USB is named `/dev/ttyACM0` and replace it accordingly:

```
$ docker run -d --name supervise --device /dev/ttyACM0:rw -p 4470:4470 -p 4471:4471 ghcr.io/valtlfelipe/ragtech-supervise:latest
```

## Docker Compose

```yaml
services:
  ragtech-supervise:
    image: ghcr.io/valtlfelipe/ragtech-supervise:latest
    devices:
      - /dev/ttyACM0:/dev/ttyACM0:rw
    ports:
      - "4470:4470"  # Supervise web interface and API
      - "4471:4471"  # Prometheus metrics exporter
    volumes:
      - ./data:/data
```

## Logging

All log output is written to stdout and stderr. Logs are categorized with 5 different prefixes:
  - `init`: Logs related to the container initialization/termination
  - `main`: Logs referring to the stderr/stdout of the main `supsvc` process
  - `supsvc`: Logs related to the main supervise functionality
  - `device-manager`: Unknown, but I assume it's related to the communication with the UPS'es
  - `serialhid`: Logs related to the Serial interface to the UPS

## Interface

The container exposes two HTTP ports:

- **Port 4470** — Ragtech Supervise web interface and API
- **Port 4471** — Prometheus metrics exporter (this container)

### Web Interface

Access the web interface at `http://localhost:4470` in your browser.

### Prometheus Metrics

A Prometheus-compatible metrics endpoint is available at `http://localhost:4471/metrics`.

**Available metrics:**

| Metric | Type | Description |
|--------|------|-------------|
| `ragtech_ups_status` | Gauge | UPS connection status (1=connected, 0=disconnected) |
| `ragtech_ups_input_voltage_volts` | Gauge | Input voltage in volts |
| `ragtech_ups_output_voltage_volts` | Gauge | Output voltage in volts |
| `ragtech_ups_output_current_amps` | Gauge | Output current in amps |
| `ragtech_ups_output_frequency_hertz` | Gauge | Output frequency in Hz |
| `ragtech_ups_output_power_watts` | Gauge | Output power in watts |
| `ragtech_ups_battery_charge_percent` | Gauge | Battery charge percentage |
| `ragtech_ups_battery_voltage_volts` | Gauge | Battery voltage in volts |
| `ragtech_ups_temperature_celsius` | Gauge | UPS temperature in Celsius |
| `ragtech_ups_load_percent` | Gauge | Load as percentage of nominal power |
| `ragtech_ups_led_red` | Gauge | Red LED state (0 or 255) |
| `ragtech_ups_led_green` | Gauge | Green LED state (0 or 255) |
| `ragtech_ups_led_blue` | Gauge | Blue LED state (0 or 255) |
| `ragtech_system_uptime_milliseconds` | Gauge | System uptime in milliseconds since epoch |
| `ragtech_collector_scrape_duration_seconds` | Gauge | Duration of the last scrape |
| `ragtech_collector_scrape_errors_total` | Counter | Total number of scrape errors |
| `ragtech_process_cpu_seconds_total` | Counter | Exporter process CPU time |
| `ragtech_process_open_fds` | Gauge | Exporter open file descriptors |
| `ragtech_process_max_fds` | Gauge | Exporter max file descriptors |
| `ragtech_process_resident_memory_bytes` | Gauge | Exporter resident memory (RSS) |
| `ragtech_process_virtual_memory_bytes` | Gauge | Exporter virtual memory |
| `ragtech_process_start_time_seconds` | Gauge | Exporter process start time (unix seconds) |

All UPS metrics include `device_id` and `device_name` labels. Collector scrape metrics include a `collector` label (`ups`). Scrape errors also include `error_type` (`system_status`, `list_devices`, or `device_status`).

`ragtech_ups_load_percent` is only emitted when the UPS reports a nominal output power greater than zero.

`ragtech_collector_scrape_errors_total` is typed as a counter, but the collector emits `1` only while that step is failing and omits the series on success. Treat it as a current-error flag (do not `rate()` it).

**Health check:**

```
$ curl http://localhost:4471/health
OK
```

**Example Prometheus scrape config:**

```yaml
scrape_configs:
  - job_name: 'ragtech-supervise'
    static_configs:
      - targets: ['localhost:4471']
```

### Grafana dashboard

A Grafana dashboard for the `ragtech_*` metrics lives at [`grafana/ragtech-supervise.json`](grafana/ragtech-supervise.json). It targets Grafana 10+ with a Prometheus data source.

**Import via the UI**

1. In Grafana, open **Dashboards → New → Import**.
2. Upload `grafana/ragtech-supervise.json`, or paste its contents.
3. Select the Prometheus data source that scrapes this exporter.

**Provision from disk**

Point a Grafana file provider at the `grafana` directory (or copy the JSON into an existing dashboards folder):

```yaml
apiVersion: 1
providers:
  - name: ragtech
    type: file
    options:
      path: /var/lib/grafana/dashboards
```

The dashboard UID is `ragtech-supervise`. Re-importing or re-provisioning updates the same dashboard in place.

**Variables**

| Variable | What it filters |
|----------|-----------------|
| Data source | Prometheus data source |
| Instance | Exporter `instance` label (multi-select, includes All) |
| Device | UPS `device_name` (multi-select, includes All, scoped to the selected instance) |

Default time range is the last 6 hours. Auto-refresh is 15s.

**Panels**

| Row | What it shows |
|-----|----------------|
| Overview | Connection, battery, load, temperature, input/output voltage, power, and current, plus a per-device table of the latest readings |
| Electrical | Voltage, output power/current, frequency, load, and temperature over time |
| Battery | Charge percentage and battery voltage |
| Status and LEDs | Connection timeline and red/green/blue LED timeline (LED values are normalized from 0/255 to 0/1) |
| Exporter | Scrape duration, scrape-error flags, process memory/CPU/FDs, and Supervise vs exporter uptime |

Supervise uptime is derived from `ragtech_system_uptime_milliseconds` (milliseconds since epoch):

```
time() - ragtech_system_uptime_milliseconds / 1000
```

Exporter uptime uses `ragtech_process_start_time_seconds`.

### SQLite Database

Alternatively, have programatic access to the UPS data by querying the underlying SQLite database. 

**IMPORTANT:** The SQLite database is set to use `WAL` as the journaling mode, so you can read the
database while it's being written to. Because of that, you need to also account for all the database
files:
  - /data/monit.db
  - /data/monit.db-wal
  - /data/monit.db-shm

This is how you would run the container with the database mounted to the host filesystem:

```
$ mkdir host-db-path
$ docker run -d --name supervise --device /dev/ttyACM0:rw -p 4470:4470 -p 4471:4471 -v ./host-db-path:/data ghcr.io/valtlfelipe/ragtech-supervise:latest
```

## License

Apache 2.0
