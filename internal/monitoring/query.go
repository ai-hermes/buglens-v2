package monitoring

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func ParseLastDurationMS(value string) (int64, error) {
	raw := strings.ToLower(strings.TrimSpace(value))
	if len(raw) < 2 {
		return 0, fmt.Errorf("last must look like 15m/1h/24h")
	}
	unit := raw[len(raw)-1]
	numText := raw[:len(raw)-1]
	n, err := strconv.Atoi(numText)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("last must use an integer value, e.g. 1h")
	}
	switch unit {
	case 'm':
		return int64(n) * 60_000, nil
	case 'h':
		return int64(n) * 3_600_000, nil
	case 'd':
		return int64(n) * 86_400_000, nil
	default:
		return 0, fmt.Errorf("last unit must be one of m/h/d")
	}
}

func ResolveTimeRange(fromMS, toMS *int64, last *string) (int64, int64, error) {
	if fromMS != nil && toMS != nil {
		return *fromMS, *toMS, nil
	}
	now := time.Now().UnixMilli()
	if last != nil && strings.TrimSpace(*last) != "" {
		d, err := ParseLastDurationMS(*last)
		if err != nil {
			return 0, 0, err
		}
		return now - d, now, nil
	}
	defaultFrom := now - 3_600_000
	if fromMS == nil {
		fromMS = &defaultFrom
	}
	if toMS == nil {
		toMS = &now
	}
	return *fromMS, *toMS, nil
}

func BuildRUMSearchQuery(query, eventType, appID string, appTypes []string, exceptionMessage, keyword string) string {
	if strings.TrimSpace(query) != "" {
		return query
	}
	terms := []string{"*"}
	filteredAppTypes := make([]string, 0)
	for _, t := range appTypes {
		t = strings.TrimSpace(t)
		if t != "" {
			filteredAppTypes = append(filteredAppTypes, t)
		}
	}
	if len(filteredAppTypes) > 0 {
		parts := make([]string, 0, len(filteredAppTypes))
		for _, t := range filteredAppTypes {
			parts = append(parts, fmt.Sprintf("app.type : %s", t))
		}
		terms = append(terms, fmt.Sprintf("(%s)", strings.Join(parts, " or ")))
	}
	if strings.TrimSpace(eventType) == "" {
		eventType = "exception"
	}
	terms = append(terms, fmt.Sprintf("event_type: %s", eventType))
	if appID != "" {
		terms = append(terms, fmt.Sprintf("app.id : %s", quoteSLS(appID)))
	}
	if exceptionMessage != "" {
		terms = append(terms, fmt.Sprintf("exception.message : %s", quoteSLS(exceptionMessage)))
	}
	if keyword != "" {
		terms = append(terms, keyword)
	}
	return strings.Join(terms, " and ")
}

func quoteSLS(v string) string {
	v = strings.ReplaceAll(v, "\\", "\\\\")
	v = strings.ReplaceAll(v, "\"", "\\\"")
	return fmt.Sprintf("\"%s\"", v)
}
