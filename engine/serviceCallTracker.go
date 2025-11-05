package engine

import (
	"math"
	"sync"
	"time"
)

// ServiceCallEntry tracks a service call context for an entity
type ServiceCallEntry struct {
	ContextID string
	Timestamp time.Time
	Params    map[string]any // e.g., {"brightness": 87.0}
}

// ServiceCallTracker tracks recent service call contexts per entity to enable
// linking real device events back to their originating service calls
// Uses FIFO queues to handle multiple service calls for the same entity
type ServiceCallTracker struct {
	mu      sync.RWMutex
	entries map[string][]*ServiceCallEntry
	ttl     time.Duration
}

// NewServiceCallTracker creates a new ServiceCallTracker with the specified TTL
func NewServiceCallTracker(ttl time.Duration) *ServiceCallTracker {
	tracker := &ServiceCallTracker{
		entries: make(map[string][]*ServiceCallEntry),
		ttl:     ttl,
	}

	// Start cleanup goroutine
	tracker.startCleanup()

	return tracker
}

// Record stores a service call context for the given entity
func (s *ServiceCallTracker) Record(entityID string, contextID string, params map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := &ServiceCallEntry{
		ContextID: contextID,
		Timestamp: time.Now(),
		Params:    params,
	}

	s.entries[entityID] = append(s.entries[entityID], entry)
}

// startCleanup starts a goroutine that periodically removes expired entries
func (s *ServiceCallTracker) startCleanup() {
	go func() {
		ticker := time.NewTicker(s.ttl / 2) // Clean up twice as often as TTL
		defer ticker.Stop()

		for range ticker.C {
			s.cleanup()
		}
	}()
}

// cleanup removes expired entries from queues
func (s *ServiceCallTracker) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for entityID, queue := range s.entries {
		// Filter out expired entries
		filtered := make([]*ServiceCallEntry, 0, len(queue))
		for _, entry := range queue {
			if now.Sub(entry.Timestamp) <= s.ttl {
				filtered = append(filtered, entry)
			}
		}

		// Update queue or delete if empty
		if len(filtered) == 0 {
			delete(s.entries, entityID)
		} else {
			s.entries[entityID] = filtered
		}
	}
}

// FindBestMatch finds the closest matching service call entry by comparing all changed fields
// Returns the context ID of the best match with lowest total score, or false if no match found
func (s *ServiceCallTracker) FindBestMatch(entityID string, eventChanges map[string]any) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	queue, ok := s.entries[entityID]
	if !ok || len(queue) == 0 {
		return "", false
	}

	var bestMatchIndex = -1
	var bestScore = math.MaxFloat64

	now := time.Now()
	for i, entry := range queue {
		// Skip expired entries
		if now.Sub(entry.Timestamp) > s.ttl {
			continue
		}

		// Calculate total score by comparing overlapping fields
		totalScore := 0.0
		fieldCount := 0

		for key, valueInEvent := range eventChanges {
			expectedValue, exists := entry.Params[key]
			if !exists {
				continue // Field not in recorded params, skip
			}

			fieldCount++
			distance := s.calculateFieldDistance(expectedValue, valueInEvent)

			// If any field has infinite distance (mismatch), skip this entry
			if math.IsInf(distance, 1) {
				totalScore = math.MaxFloat64
				break
			}

			totalScore += distance
		}

		// Skip if total score exceeds tolerance
		if totalScore > 0.2 {
			continue
		}

		// Only consider entries with at least one matching field
		if fieldCount > 0 && totalScore < bestScore {
			bestScore = totalScore
			bestMatchIndex = i
		}
	}

	if bestMatchIndex == -1 {
		return "", false
	}

	matchedEntry := queue[bestMatchIndex]
	// Remove the matched entry from the queue
	s.entries[entityID] = append(queue[:bestMatchIndex], queue[bestMatchIndex+1:]...)
	// Clean up empty queue
	if len(s.entries[entityID]) == 0 {
		delete(s.entries, entityID)
	}
	return matchedEntry.ContextID, true
}

func (s *ServiceCallTracker) PopOldestForDomain(domain string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	queue, ok := s.entries[domain]
	if !ok || len(queue) == 0 {
		return "", false
	}

	firstEntry := queue[0]
	contextID := firstEntry.ContextID

	if len(queue) == 1 {
		delete(s.entries, domain)
	} else {
		s.entries[domain] = queue[1:]
	}

	return contextID, true
}

// calculateFieldDistance calculates normalized distance for a single field
func (s *ServiceCallTracker) calculateFieldDistance(expected, actual any) float64 {
	switch expectedVal := expected.(type) {
	case float64:
		actualVal, ok := actual.(float64)
		if ok {
			return math.Abs(expectedVal-actualVal) / 2
		}
		return math.Inf(1)

	case bool:
		actualVal, ok := actual.(bool)
		if ok {
			if expectedVal == actualVal {
				return 0
			}
		}
		return math.Inf(1)

	case string:
		actualVal, ok := actual.(string)
		if ok {
			if expectedVal == actualVal {
				return 0
			}
		}
		return math.Inf(1)

	default:
		// For unknown types or XY struct, use exact matching
		if expected == actual {
			return 0
		}
		return math.Inf(1)
	}
}
