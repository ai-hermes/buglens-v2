package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	GitLabURL       string
	GitLabToken     string
	GitLabProjectID string

	AlibabaAccessKeyID     string
	AlibabaAccessKeySecret string
	AlibabaSecurityToken   string
	AlibabaRegionID        string
	ARMSEndpoint           string
	SLSEndpoint            string
	RUMSLSProject          string
	RUMSLSLogstore         string

	MonitoringMaxRetries        int
	MonitoringBaseBackoffSecond float64
	MonitoringMaxBackoffSecond  float64
}

func FromEnv() Config {
	return Config{
		GitLabURL:                   os.Getenv("GITLAB_URL"),
		GitLabToken:                 os.Getenv("GITLAB_TOKEN"),
		GitLabProjectID:             os.Getenv("GITLAB_PROJECT_ID"),
		AlibabaAccessKeyID:          os.Getenv("BUGLENS_ALIBABA_ACCESS_KEY_ID"),
		AlibabaAccessKeySecret:      os.Getenv("BUGLENS_ALIBABA_ACCESS_KEY_SECRET"),
		AlibabaSecurityToken:        os.Getenv("BUGLENS_ALIBABA_SECURITY_TOKEN"),
		AlibabaRegionID:             os.Getenv("BUGLENS_ALIBABA_REGION_ID"),
		ARMSEndpoint:                defaultIfEmpty(os.Getenv("BUGLENS_ARMS_ENDPOINT"), "arms.aliyuncs.com"),
		SLSEndpoint:                 os.Getenv("BUGLENS_SLS_ENDPOINT"),
		RUMSLSProject:               os.Getenv("BUGLENS_RUM_SLS_PROJECT"),
		RUMSLSLogstore:              os.Getenv("BUGLENS_RUM_SLS_LOGSTORE"),
		MonitoringMaxRetries:        parseIntDefault(os.Getenv("BUGLENS_MONITORING_MAX_RETRIES"), 2),
		MonitoringBaseBackoffSecond: parseFloatDefault(os.Getenv("BUGLENS_MONITORING_BASE_BACKOFF_SECONDS"), 0.25),
		MonitoringMaxBackoffSecond:  parseFloatDefault(os.Getenv("BUGLENS_MONITORING_MAX_BACKOFF_SECONDS"), 5.0),
	}
}

func (c Config) HasMonitoringCredentials() bool {
	return strings.TrimSpace(c.AlibabaAccessKeyID) != "" &&
		strings.TrimSpace(c.AlibabaAccessKeySecret) != "" &&
		strings.TrimSpace(c.AlibabaRegionID) != ""
}

func defaultIfEmpty(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func parseIntDefault(v string, fallback int) int {
	i, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return fallback
	}
	return i
}

func parseFloatDefault(v string, fallback float64) float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
	if err != nil {
		return fallback
	}
	return f
}
