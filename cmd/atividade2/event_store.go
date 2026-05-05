package main

import (
	"time"
)

func NewEventStore(limit int) *EventStore {
	return &EventStore{
		limit:  limit,
		events: make([]ObservedEvent, 0),
	}
}

func (s *EventStore) Add(topic string, hash string, observedAt int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append(s.events, ObservedEvent{
		Topic:      topic,
		Hash:       hash,
		ObservedAt: observedAt,
	})

	if s.limit > 0 && len(s.events) > s.limit {
		s.events = s.events[len(s.events)-s.limit:]
	}
}

func (s *EventStore) SummaryLastSeconds(seconds int, now time.Time) EventsSummary {
	cutoff := now.Unix() - int64(seconds)

	s.mu.RLock()
	defer s.mu.RUnlock()

	var selected []ObservedEvent
	for _, event := range s.events {
		if event.ObservedAt >= cutoff {
			selected = append(selected, event)
		}
	}

	summary := summarizeEvents(selected)
	if seconds > 0 {
		summary.TxPerSecond = float64(summary.TxObserved) / float64(seconds)
	}

	return summary
}

func (s *EventStore) SummaryLastEvents(count int) EventsSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start := len(s.events) - count
	if start < 0 {
		start = 0
	}

	summary := summarizeEvents(s.events[start:])
	if len(s.events[start:]) < 2 {
		return summary
	}

	first := s.events[start].ObservedAt
	last := s.events[len(s.events)-1].ObservedAt
	duration := last - first
	if duration > 0 {
		summary.TxPerSecond = float64(summary.TxObserved) / float64(duration)
	}

	return summary
}

func (s *EventStore) Latest(limit int) LatestEvents {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var latest LatestEvents

	for i := len(s.events) - 1; i >= 0; i-- {
		event := s.events[i]

		switch event.Topic {
		case blockTopic:
			if len(latest.Blocks) < limit {
				latest.Blocks = append(latest.Blocks, LatestBlockEvent{
					Hash: event.Hash,
					Ts:   event.ObservedAt,
				})
			}
		case txTopic:
			if len(latest.Txs) < limit {
				latest.Txs = append(latest.Txs, LatestTxEvent{
					TxID: event.Hash,
					Ts:   event.ObservedAt,
				})
			}
		}

		if len(latest.Blocks) == limit && len(latest.Txs) == limit {
			break
		}
	}

	reverseBlocks(latest.Blocks)
	reverseTxs(latest.Txs)

	return latest
}

func (s *EventStore) LastSeenBlock() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := len(s.events) - 1; i >= 0; i-- {
		if s.events[i].Topic == blockTopic {
			return s.events[i].Hash
		}
	}

	return ""
}

func summarizeEvents(events []ObservedEvent) EventsSummary {
	var summary EventsSummary

	for _, event := range events {
		switch event.Topic {
		case blockTopic:
			summary.BlocksObserved++
		case txTopic:
			summary.TxObserved++
		}

		if event.ObservedAt > summary.LastEventTime {
			summary.LastEventTime = event.ObservedAt
		}
	}

	return summary
}

func reverseBlocks(events []LatestBlockEvent) {
	for left, right := 0, len(events)-1; left < right; left, right = left+1, right-1 {
		events[left], events[right] = events[right], events[left]
	}
}

func reverseTxs(events []LatestTxEvent) {
	for left, right := 0, len(events)-1; left < right; left, right = left+1, right-1 {
		events[left], events[right] = events[right], events[left]
	}
}
