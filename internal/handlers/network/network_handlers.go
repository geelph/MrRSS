package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/network"
)

// HandleDetectNetwork detects network speed and updates settings
func HandleDetectNetwork(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	detector := network.NewDetector()
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	result := detector.DetectSpeed(ctx)

	// Store results in settings
	if result.DetectionSuccess {
		h.DB.SetSetting("network_speed", string(result.SpeedLevel))
		h.DB.SetSetting("network_bandwidth_mbps", fmt.Sprintf("%.2f", result.BandwidthMbps))
		h.DB.SetSetting("network_latency_ms", strconv.FormatInt(result.LatencyMs, 10))
		h.DB.SetSetting("max_concurrent_refreshes", strconv.Itoa(result.MaxConcurrency))
		h.DB.SetSetting("last_network_test", result.DetectionTime.Format(time.RFC3339))
	}

	// Return results to frontend
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleGetNetworkInfo returns current network detection info from settings
func HandleGetNetworkInfo(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	speedLevel, _ := h.DB.GetSetting("network_speed")
	bandwidthStr, _ := h.DB.GetSetting("network_bandwidth_mbps")
	latencyStr, _ := h.DB.GetSetting("network_latency_ms")
	concurrencyStr, _ := h.DB.GetSetting("max_concurrent_refreshes")
	lastTestStr, _ := h.DB.GetSetting("last_network_test")

	bandwidth, err := strconv.ParseFloat(bandwidthStr, 64)
	if err != nil {
		bandwidth = 0
	}

	latency, err := strconv.ParseInt(latencyStr, 10, 64)
	if err != nil {
		latency = 0
	}

	concurrency, err := strconv.Atoi(concurrencyStr)
	if err != nil || concurrency < 1 {
		concurrency = 5 // Default
	}

	var lastTest time.Time
	if lastTestStr != "" {
		lastTest, _ = time.Parse(time.RFC3339, lastTestStr)
	}

	result := network.DetectionResult{
		SpeedLevel:       network.SpeedLevel(speedLevel),
		BandwidthMbps:    bandwidth,
		LatencyMs:        latency,
		MaxConcurrency:   concurrency,
		DetectionTime:    lastTest,
		DetectionSuccess: speedLevel != "",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
