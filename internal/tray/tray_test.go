package tray

import "testing"

func TestGetLabels_DefaultAndChinese(t *testing.T) {
	m := &Manager{}

	// default
	lbl := m.getLabels()
	if lbl.show != "Show MrRSS" {
		t.Fatalf("unexpected default show label: %s", lbl.show)
	}

	m.lang = "zh-CN"
	lbl2 := m.getLabels()
	if lbl2.show == "Show MrRSS" {
		t.Fatalf("expected Chinese label, got default: %s", lbl2.show)
	}
}
