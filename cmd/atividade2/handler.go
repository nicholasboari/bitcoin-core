package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

const defaultSummarySeconds = 60

func eventsSummaryHandler(store *EventStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		query := r.URL.Query()

		var summary EventsSummary
		if rawSeconds := query.Get("seconds"); rawSeconds != "" {
			seconds, ok := parsePositiveInt(rawSeconds)
			if !ok {
				http.Error(w, "seconds must be a positive integer", http.StatusBadRequest)
				return
			}

			summary = store.SummaryLastSeconds(seconds, time.Now())
		} else if rawEvents := query.Get("events"); rawEvents != "" {
			events, ok := parsePositiveInt(rawEvents)
			if !ok {
				http.Error(w, "events must be a positive integer", http.StatusBadRequest)
				return
			}

			summary = store.SummaryLastEvents(events)
		} else {
			summary = store.SummaryLastSeconds(defaultSummarySeconds, time.Now())
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summary)
	}
}

func parsePositiveInt(value string) (int, bool) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, false
	}

	return parsed, true
}
