package main

import (
	"os"
	"strings"
)

const metricsPathEnv = "REPODOCTOR_METRICS_PATH"

func resolveMetricsExportPath(analyzeRoot string) (string, bool, error) {
	raw := strings.TrimSpace(os.Getenv(metricsPathEnv))
	if raw == "" {
		return "", false, nil
	}

	path, err := sanitizeProfilePath(analyzeRoot, raw)
	if err != nil {
		return "", false, err
	}

	return path, true, nil
}
