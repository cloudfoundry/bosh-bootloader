package config

import (
	"fmt"
	"regexp"
	"strings"
)

// GCP resource label constraints:
// https://cloud.google.com/resource-manager/docs/labels-overview
// Keys must start with a lowercase letter and may contain lowercase letters,
// digits, dashes and underscores; values must start with a lowercase letter or
// digit and may additionally contain underscores. The value pattern also
// matches the google CPI's label validation, which requires a non-empty value.
var (
	gcpLabelKeyPattern   = regexp.MustCompile(`^[a-z]([-_a-z0-9]{0,61}[a-z0-9])?$`)
	gcpLabelValuePattern = regexp.MustCompile(`^[a-z0-9]([-_a-z0-9]{0,61}[a-z0-9])?$`)
)

const maxGCPLabels = 64

// parseGCPLabels converts a list of "key=value" entries into a GCP resource
// label map. Keys and values are lower-cased to satisfy GCP's label
// constraints; any entry that is still invalid returns an error. Blank entries
// (for example when BBL_GCP_LABELS is set to an empty string) are ignored, and
// a list that yields no labels returns nil so callers can distinguish "not
// provided" from "explicitly empty".
func parseGCPLabels(entries []string) (map[string]string, error) {
	labels := make(map[string]string)

	for _, entry := range entries {
		if strings.TrimSpace(entry) == "" {
			continue
		}

		key, value, found := strings.Cut(entry, "=")
		if !found {
			return nil, fmt.Errorf("invalid GCP label %q: expected key=value", entry)
		}

		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.ToLower(strings.TrimSpace(value))

		if !gcpLabelKeyPattern.MatchString(key) {
			return nil, fmt.Errorf("invalid GCP label key %q: must start with a lowercase letter and may contain only lowercase letters, digits, dashes or underscores (max 63 characters)", key)
		}
		if !gcpLabelValuePattern.MatchString(value) {
			return nil, fmt.Errorf("invalid GCP label value %q: must contain at least one lowercase letter, digit, dash or underscore (max 63 characters)", value)
		}

		labels[key] = value
	}

	if len(labels) == 0 {
		return nil, nil
	}

	if len(labels) > maxGCPLabels {
		return nil, fmt.Errorf("too many GCP labels: %d provided, maximum is %d", len(labels), maxGCPLabels)
	}

	return labels, nil
}
