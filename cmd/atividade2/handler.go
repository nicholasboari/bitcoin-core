package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

const defaultSummarySeconds = 60
const defaultLatestEvents = 10

func eventsSummaryHandler(store *EventStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		query := r.URL.Query()

		seconds := defaultSummarySeconds
		if rawSeconds := query.Get("seconds"); rawSeconds != "" {
			parsedSeconds, ok := parsePositiveInt(rawSeconds)
			if !ok {
				http.Error(w, "seconds must be a positive integer", http.StatusBadRequest)
				return
			}

			seconds = parsedSeconds
		}

		summary := store.SummaryLastSeconds(seconds, time.Now())

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(summary)
	}
}

func latestEventsHandler(store *EventStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		latest := store.Latest(defaultLatestEvents)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(latest)
	}
}

func parsePositiveInt(value string) (int, bool) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, false
	}

	return parsed, true
}
