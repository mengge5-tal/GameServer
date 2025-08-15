package scheduler

import (
	"log"
	"time"
	"GameServer/internal/application/service"
)

// DailyResetScheduler handles daily reset tasks
type DailyResetScheduler struct {
	killCountService *service.KillCountService
	stopChan         chan bool
}

// NewDailyResetScheduler creates a new daily reset scheduler
func NewDailyResetScheduler(killCountService *service.KillCountService) *DailyResetScheduler {
	return &DailyResetScheduler{
		killCountService: killCountService,
		stopChan:         make(chan bool),
	}
}

// Start starts the daily reset scheduler
func (s *DailyResetScheduler) Start() {
	go s.run()
	log.Println("Daily reset scheduler started - will reset kill counts at 4:00 AM daily")
}

// Stop stops the daily reset scheduler
func (s *DailyResetScheduler) Stop() {
	close(s.stopChan)
	log.Println("Daily reset scheduler stopped")
}

// run runs the scheduler loop
func (s *DailyResetScheduler) run() {
	for {
		// Calculate time until next 4:00 AM
		now := time.Now()
		next4AM := s.getNext4AM(now)
		duration := next4AM.Sub(now)
		
		log.Printf("Next kill count reset scheduled for: %s (in %v)", next4AM.Format("2006-01-02 15:04:05"), duration)
		
		// Wait until 4:00 AM or stop signal
		select {
		case <-time.After(duration):
			s.performDailyReset()
		case <-s.stopChan:
			return
		}
	}
}

// getNext4AM calculates the next 4:00 AM time
func (s *DailyResetScheduler) getNext4AM(now time.Time) time.Time {
	// Get today's 4:00 AM
	today4AM := time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, now.Location())
	
	// If it's already past 4:00 AM today, schedule for tomorrow
	if now.After(today4AM) {
		return today4AM.Add(24 * time.Hour)
	}
	
	return today4AM
}

// performDailyReset performs the daily reset operation
func (s *DailyResetScheduler) performDailyReset() {
	log.Println("Starting daily kill count reset...")
	
	start := time.Now()
	err := s.killCountService.ResetAllKillCounts()
	duration := time.Since(start)
	
	if err != nil {
		log.Printf("Error during daily kill count reset: %v", err)
	} else {
		log.Printf("Daily kill count reset completed successfully in %v", duration)
	}
}