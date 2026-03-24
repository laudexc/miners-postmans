package miner

import "sync"

// статистика конкретного шахтёра по id воркера
type WorkerStats struct {
	MinedTotal     int
	FailedAttempts int
}

// безопасный снимок агрегированной статистики на момент вызова
type Snapshot struct {
	TotalMined     int
	FailedAttempts int
	Workers        map[int]WorkerStats
}

// потокобезопасный накопитель, в который пишут все горутины шахтёров
type Stats struct {
	mu             sync.RWMutex
	totalMined     int
	failedAttempts int
	workers        map[int]WorkerStats
}

// создаёт общий объект статистики, который передаётся в minerpool
func NewStats() *Stats {
	return &Stats{
		workers: make(map[int]WorkerStats),
	}
}

// вызывается воркером после успешной передачи угля в канал
func (s *Stats) RecordMined(workerID int, amount int) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalMined += amount
	ws := s.workers[workerID]
	ws.MinedTotal += amount
	s.workers[workerID] = ws
}

// вызывается, когда воркер завершает цикл из-за ctx.done
func (s *Stats) RecordFailedAttempt(workerID int) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.failedAttempts++
	ws := s.workers[workerID]
	ws.FailedAttempts++
	s.workers[workerID] = ws
}

// возвращает копию данных, чтобы внешний код не менял внутреннее состояние stats
func (s *Stats) Snapshot() Snapshot {
	if s == nil {
		return Snapshot{}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	workers := make(map[int]WorkerStats, len(s.workers))
	for workerID, stats := range s.workers {
		workers[workerID] = stats
	}

	return Snapshot{
		TotalMined:     s.totalMined,
		FailedAttempts: s.failedAttempts,
		Workers:        workers,
	}
}
