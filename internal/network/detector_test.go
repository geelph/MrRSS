package network

import "testing"

func TestCalculateSpeedLevel(t *testing.T) {
	d := NewDetector()

	tests := []struct {
		lat             int64
		bw              float64
		want            SpeedLevel
		wantConcurrency int
	}{
		{300, 0.5, SpeedSlow, 5},
		{150, 5.0, SpeedMedium, 8},
		{50, 20.0, SpeedFast, 15},
	}

	for _, tt := range tests {
		got, conc := d.calculateSpeedLevel(tt.lat, tt.bw)
		if got != tt.want || conc != tt.wantConcurrency {
			t.Fatalf("lat=%d bw=%.1f: got=%v,%d want=%v,%d", tt.lat, tt.bw, got, conc, tt.want, tt.wantConcurrency)
		}
	}
}
