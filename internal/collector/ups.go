package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"ragtech-supervise/internal/api"
)

type UPSCollector struct {
	client *api.Client

	statusDesc          *prometheus.Desc
	inputVoltageDesc    *prometheus.Desc
	outputVoltageDesc   *prometheus.Desc
	outputCurrentDesc   *prometheus.Desc
	outputFrequencyDesc *prometheus.Desc
	outputPowerDesc     *prometheus.Desc
	batteryChargeDesc   *prometheus.Desc
	batteryVoltageDesc  *prometheus.Desc
	temperatureDesc     *prometheus.Desc
	loadPercentDesc     *prometheus.Desc
	ledRedDesc          *prometheus.Desc
	ledGreenDesc        *prometheus.Desc
	ledBlueDesc         *prometheus.Desc
	systemUptimeDesc    *prometheus.Desc
	scrapeDurationDesc  *prometheus.Desc
	scrapeErrorsDesc    *prometheus.Desc
}

func NewUPSCollector(client *api.Client) *UPSCollector {
	ns := "ragtech"
	c := &UPSCollector{client: client}

	c.statusDesc = prometheus.NewDesc(
		fmt.Sprintf("%s_ups_status", ns), "UPS connection status (1=connected, 0=disconnected)",
		[]string{"device_id", "device_name"}, nil,
	)
	c.inputVoltageDesc = prometheus.NewDesc(
		fmt.Sprintf("%s_ups_input_voltage_volts", ns), "UPS input voltage in volts",
		[]string{"device_id", "device_name"}, nil,
	)
	c.outputVoltageDesc = prometheus.NewDesc(
		fmt.Sprintf("%s_ups_output_voltage_volts", ns), "UPS output voltage in volts",
		[]string{"device_id", "device_name"}, nil,
	)
	c.outputCurrentDesc = prometheus.NewDesc(
		fmt.Sprintf("%s_ups_output_current_amps", ns), "UPS output current in amps",
		[]string{"device_id", "device_name"}, nil,
	)
	c.outputFrequencyDesc = prometheus.NewDesc(
		fmt.Sprintf("%s_ups_output_frequency_hertz", ns), "UPS output frequency in Hz",
		[]string{"device_id", "device_name"}, nil,
	)
	c.outputPowerDesc = prometheus.NewDesc(
		fmt.Sprintf("%s_ups_output_power_watts", ns), "UPS output power in watts",
		[]string{"device_id", "device_name"}, nil,
	)
	c.batteryChargeDesc = prometheus.NewDesc(
		fmt.Sprintf("%s_ups_battery_charge_percent", ns), "UPS battery charge percentage",
		[]string{"device_id", "device_name"}, nil,
	)
	c.batteryVoltageDesc = prometheus.NewDesc(
		fmt.Sprintf("%s_ups_battery_voltage_volts", ns), "UPS battery voltage in volts",
		[]string{"device_id", "device_name"}, nil,
	)
	c.temperatureDesc = prometheus.NewDesc(
		fmt.Sprintf("%s_ups_temperature_celsius", ns), "UPS temperature in Celsius",
		[]string{"device_id", "device_name"}, nil,
	)
	c.loadPercentDesc = prometheus.NewDesc(
		fmt.Sprintf("%s_ups_load_percent", ns), "UPS load as percentage of nominal power",
		[]string{"device_id", "device_name"}, nil,
	)
	c.ledRedDesc = prometheus.NewDesc(
		fmt.Sprintf("%s_ups_led_red", ns), "UPS red LED state (0 or 255)",
		[]string{"device_id", "device_name"}, nil,
	)
	c.ledGreenDesc = prometheus.NewDesc(
		fmt.Sprintf("%s_ups_led_green", ns), "UPS green LED state (0 or 255)",
		[]string{"device_id", "device_name"}, nil,
	)
	c.ledBlueDesc = prometheus.NewDesc(
		fmt.Sprintf("%s_ups_led_blue", ns), "UPS blue LED state (0 or 255)",
		[]string{"device_id", "device_name"}, nil,
	)
	c.systemUptimeDesc = prometheus.NewDesc(
		fmt.Sprintf("%s_system_uptime_milliseconds", ns), "System uptime in milliseconds since epoch",
		nil, nil,
	)
	c.scrapeDurationDesc = prometheus.NewDesc(
		fmt.Sprintf("%s_collector_scrape_duration_seconds", ns), "Duration of the last scrape",
		[]string{"collector"}, nil,
	)
	c.scrapeErrorsDesc = prometheus.NewDesc(
		fmt.Sprintf("%s_collector_scrape_errors_total", ns), "Total number of scrape errors",
		[]string{"collector", "error_type"}, nil,
	)

	return c
}

func (c *UPSCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.statusDesc
	ch <- c.inputVoltageDesc
	ch <- c.outputVoltageDesc
	ch <- c.outputCurrentDesc
	ch <- c.outputFrequencyDesc
	ch <- c.outputPowerDesc
	ch <- c.batteryChargeDesc
	ch <- c.batteryVoltageDesc
	ch <- c.temperatureDesc
	ch <- c.loadPercentDesc
	ch <- c.ledRedDesc
	ch <- c.ledGreenDesc
	ch <- c.ledBlueDesc
	ch <- c.systemUptimeDesc
	ch <- c.scrapeDurationDesc
	ch <- c.scrapeErrorsDesc
}

func (c *UPSCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	start := time.Now()

	// Get system status
	status, err := c.client.GetSystemStatus(ctx)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.scrapeErrorsDesc, prometheus.CounterValue, 1, "ups", "system_status")
	} else {
		ch <- prometheus.MustNewConstMetric(c.systemUptimeDesc, prometheus.GaugeValue, float64(status.UpTime))
	}

	// Get devices
	devices, err := c.client.GetDevices(ctx)
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.scrapeErrorsDesc, prometheus.CounterValue, 1, "ups", "list_devices")
		return
	}

	for _, device := range devices {
		deviceStatus, err := c.client.GetDeviceStatus(ctx, device.ID)
		if err != nil {
			ch <- prometheus.MustNewConstMetric(c.scrapeErrorsDesc, prometheus.CounterValue, 1, "ups", "device_status")
			continue
		}

		v := deviceStatus.Vars
		labels := []string{deviceStatus.ID, deviceStatus.Name}

		ch <- prometheus.MustNewConstMetric(c.statusDesc, prometheus.GaugeValue, float64(deviceStatus.Status.Connected), labels...)
		ch <- prometheus.MustNewConstMetric(c.inputVoltageDesc, prometheus.GaugeValue, v.VInput, labels...)
		ch <- prometheus.MustNewConstMetric(c.outputVoltageDesc, prometheus.GaugeValue, v.VOutput, labels...)
		ch <- prometheus.MustNewConstMetric(c.outputCurrentDesc, prometheus.GaugeValue, v.IOutput, labels...)
		ch <- prometheus.MustNewConstMetric(c.outputFrequencyDesc, prometheus.GaugeValue, v.FOutput, labels...)
		ch <- prometheus.MustNewConstMetric(c.outputPowerDesc, prometheus.GaugeValue, v.POutput, labels...)
		ch <- prometheus.MustNewConstMetric(c.batteryChargeDesc, prometheus.GaugeValue, v.CBattery, labels...)
		ch <- prometheus.MustNewConstMetric(c.batteryVoltageDesc, prometheus.GaugeValue, v.VBattery, labels...)
		ch <- prometheus.MustNewConstMetric(c.temperatureDesc, prometheus.GaugeValue, v.Temperature, labels...)

		if v.NominalPOutput > 0 {
			loadPercent := (v.POutput / v.NominalPOutput) * 100
			ch <- prometheus.MustNewConstMetric(c.loadPercentDesc, prometheus.GaugeValue, loadPercent, labels...)
		}

		ch <- prometheus.MustNewConstMetric(c.ledRedDesc, prometheus.GaugeValue, float64(v.LedRed), labels...)
		ch <- prometheus.MustNewConstMetric(c.ledGreenDesc, prometheus.GaugeValue, float64(v.LedGreen), labels...)
		ch <- prometheus.MustNewConstMetric(c.ledBlueDesc, prometheus.GaugeValue, float64(v.LedBlue), labels...)
	}

	duration := time.Since(start).Seconds()
	ch <- prometheus.MustNewConstMetric(c.scrapeDurationDesc, prometheus.GaugeValue, duration, "ups")
}
