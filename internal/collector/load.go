package collector

import "ragtech-supervise/internal/api"

// Supervise's pOutput is already a 0–100 load percentage (the web UI labels it
// "Potência (%)"). The firmware integer often floors to 0 below ~1% load, so
// when voltage and current are present we also compute apparent-power load and
// keep the larger of the two.
func loadPercent(v api.DeviceVars) (float64, bool) {
	fromAPI := v.POutput
	fromVI := 0.0
	haveVI := v.NominalPOutput > 0 && v.VOutput > 0 && v.IOutput > 0
	if haveVI {
		fromVI = (v.VOutput * v.IOutput / v.NominalPOutput) * 100
	}

	switch {
	case haveVI && fromVI > fromAPI:
		return fromVI, true
	case fromAPI > 0 || v.NominalPOutput > 0 || haveVI:
		return fromAPI, true
	default:
		return 0, false
	}
}

// outputPowerWatts is apparent output power (V×I). Falls back to
// load% × nominal when current or voltage is missing.
func outputPowerWatts(v api.DeviceVars) float64 {
	if v.VOutput > 0 && v.IOutput > 0 {
		return v.VOutput * v.IOutput
	}
	if v.NominalPOutput > 0 && v.POutput > 0 {
		return v.POutput / 100 * v.NominalPOutput
	}
	return 0
}
