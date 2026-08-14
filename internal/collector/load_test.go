package collector

import (
	"testing"

	"ragtech-supervise/internal/api"
)

func TestLoadPercentUsesPOutputAsPercent(t *testing.T) {
	// The old formula (pOutput/nominal)*100 turned 7% into 0.58%.
	got, ok := loadPercent(api.DeviceVars{POutput: 7, NominalPOutput: 1200})
	if !ok {
		t.Fatal("expected load")
	}
	if got != 7 {
		t.Fatalf("pOutput is already percent: got %v, want 7", got)
	}
}

func TestLoadPercentPrefersApparentPowerWhenFirmwareFloors(t *testing.T) {
	// 113 V × 2 A / 1200 VA = 18.83%; firmware pOutput floored to 0.
	got, ok := loadPercent(api.DeviceVars{
		POutput:        0,
		VOutput:        113,
		IOutput:        2,
		NominalPOutput: 1200,
	})
	if !ok {
		t.Fatal("expected load")
	}
	want := 113.0 * 2 / 1200 * 100
	if abs(got-want) > 0.01 {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestLoadPercentKeepsFirmwareWhenLarger(t *testing.T) {
	got, ok := loadPercent(api.DeviceVars{
		POutput:        40,
		VOutput:        220,
		IOutput:        0.1,
		NominalPOutput: 1200,
	})
	if !ok {
		t.Fatal("expected load")
	}
	if got != 40 {
		t.Fatalf("got %v, want firmware 40", got)
	}
}

func TestOutputPowerWattsFromVI(t *testing.T) {
	got := outputPowerWatts(api.DeviceVars{VOutput: 113, IOutput: 2, POutput: 7, NominalPOutput: 1200})
	if abs(got-226) > 0.01 {
		t.Fatalf("got %v, want 226", got)
	}
}

func TestOutputPowerWattsFallbackFromPercent(t *testing.T) {
	got := outputPowerWatts(api.DeviceVars{POutput: 10, NominalPOutput: 1200})
	if abs(got-120) > 0.01 {
		t.Fatalf("got %v, want 120", got)
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
