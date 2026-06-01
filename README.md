# Ragtech Supervise Docker Image

Ragtech is a Brazilian company that produces UPS (more commonly known in Brazil as no-break)
devices. They have a software called Supervise that is used to monitor and control the UPS devices.

The existing Supervise software packaging is not really suitable with modern Linux distributions.
This project aims to provide a Docker image with Supervise supplying all required dependencies.

## Usage

You can run a new container from the computer where the UPS USB cable is plugged to. Validate that
the serial interface created by this USB is named `/dev/ttyACM0` and replace it accordingly:

```
$ docker run -d --name supervise --device /dev/ttyACM0:rw -p 4470:4470 ghcr.io/valtlfelipe/ragtech-supervise:latest
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

All UPS metrics include `device_id` and `device_name` labels.

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
$ docker run [...] -v ./host-db-path:/data ghcr.io/valtlfelipe/ragtech-supervise:latest
```

## License

Apache 2.0
