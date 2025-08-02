package metrics

import (
	"sync"
	"time"
)

// Metrics holds application metrics
type Metrics struct {
	mutex                sync.RWMutex
	ConnectionCount      int64                  `json:"connection_count"`
	TotalConnections     int64                  `json:"total_connections"`
	MessagesProcessed    int64                  `json:"messages_processed"`
	ErrorCount           int64                  `json:"error_count"`
	DatabaseQueries      int64                  `json:"database_queries"`
	RequestDurations     map[string][]int64     `json:"request_durations"`
	LastUpdated          time.Time              `json:"last_updated"`
}

var globalMetrics *Metrics
var once sync.Once

// Init initializes the global metrics instance
func Init() {
	once.Do(func() {
		globalMetrics = &Metrics{
			RequestDurations: make(map[string][]int64),
			LastUpdated:      time.Now(),
		}
	})
}

// IncrementConnections increments the connection count
func IncrementConnections() {
	if globalMetrics == nil {
		return
	}
	globalMetrics.mutex.Lock()
	globalMetrics.ConnectionCount++
	globalMetrics.TotalConnections++
	globalMetrics.LastUpdated = time.Now()
	globalMetrics.mutex.Unlock()
}

// DecrementConnections decrements the connection count
func DecrementConnections() {
	if globalMetrics == nil {
		return
	}
	globalMetrics.mutex.Lock()
	if globalMetrics.ConnectionCount > 0 {
		globalMetrics.ConnectionCount--
	}
	globalMetrics.LastUpdated = time.Now()
	globalMetrics.mutex.Unlock()
}

// IncrementMessages increments the messages processed count
func IncrementMessages() {
	if globalMetrics == nil {
		return
	}
	globalMetrics.mutex.Lock()
	globalMetrics.MessagesProcessed++
	globalMetrics.LastUpdated = time.Now()
	globalMetrics.mutex.Unlock()
}

// IncrementErrors increments the error count
func IncrementErrors() {
	if globalMetrics == nil {
		return
	}
	globalMetrics.mutex.Lock()
	globalMetrics.ErrorCount++
	globalMetrics.LastUpdated = time.Now()
	globalMetrics.mutex.Unlock()
}

// IncrementDatabaseQueries increments the database queries count
func IncrementDatabaseQueries() {
	if globalMetrics == nil {
		return
	}
	globalMetrics.mutex.Lock()
	globalMetrics.DatabaseQueries++
	globalMetrics.LastUpdated = time.Now()
	globalMetrics.mutex.Unlock()
}

// RecordRequestDuration records request duration for a specific action
func RecordRequestDuration(action string, duration time.Duration) {
	if globalMetrics == nil {
		return
	}
	globalMetrics.mutex.Lock()
	if globalMetrics.RequestDurations[action] == nil {
		globalMetrics.RequestDurations[action] = make([]int64, 0)
	}
	
	// Keep only the last 100 durations for each action
	durations := globalMetrics.RequestDurations[action]
	if len(durations) >= 100 {
		durations = durations[1:]
	}
	globalMetrics.RequestDurations[action] = append(durations, duration.Milliseconds())
	globalMetrics.LastUpdated = time.Now()
	globalMetrics.mutex.Unlock()
}

// GetMetrics returns a copy of current metrics
func GetMetrics() Metrics {
	if globalMetrics == nil {
		return Metrics{}
	}
	globalMetrics.mutex.RLock()
	defer globalMetrics.mutex.RUnlock()
	
	// Create a deep copy
	metrics := Metrics{
		ConnectionCount:   globalMetrics.ConnectionCount,
		TotalConnections:  globalMetrics.TotalConnections,
		MessagesProcessed: globalMetrics.MessagesProcessed,
		ErrorCount:        globalMetrics.ErrorCount,
		DatabaseQueries:   globalMetrics.DatabaseQueries,
		LastUpdated:       globalMetrics.LastUpdated,
		RequestDurations:  make(map[string][]int64),
	}
	
	// Copy request durations
	for action, durations := range globalMetrics.RequestDurations {
		metrics.RequestDurations[action] = make([]int64, len(durations))
		copy(metrics.RequestDurations[action], durations)
	}
	
	return metrics
}

// Reset resets all metrics to zero
func Reset() {
	if globalMetrics == nil {
		return
	}
	globalMetrics.mutex.Lock()
	globalMetrics.ConnectionCount = 0
	globalMetrics.TotalConnections = 0
	globalMetrics.MessagesProcessed = 0
	globalMetrics.ErrorCount = 0
	globalMetrics.DatabaseQueries = 0
	globalMetrics.RequestDurations = make(map[string][]int64)
	globalMetrics.LastUpdated = time.Now()
	globalMetrics.mutex.Unlock()
}