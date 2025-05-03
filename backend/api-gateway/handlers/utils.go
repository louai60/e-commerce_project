package handlers

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// formatTimestamp formats a protobuf timestamp to ISO 8601 string
func formatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().Format(time.RFC3339)
}
