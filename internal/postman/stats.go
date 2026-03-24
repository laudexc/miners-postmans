package postman

import (
	"maps"
	"sync"
)

// статистика конкретного почтальона по id воркера
type WorkerStats struct {
	DeliveredTotal   int
	UndeliveredTotal int
}

// безопасный снимок агрегированной статистики на момент вызова
type Snapshot struct {
	TotalDelivered   int
	TotalUndelivered int
	Workers          map[int]WorkerStats
}

// потокобезопасный накопитель, в который пишут все горутины почтальонов
type Stats struct {
	mu               sync.RWMutex
	totalDelivered   int
	totalUndelivered int
	workers          map[int]WorkerStats
}

// создаёт общий объект статистики, который передаётся в postmanpool
func NewStats() *Stats {
	return &Stats{
		workers: make(map[int]WorkerStats),
	}
}

// вызывается воркером после успешной передачи письма в канал
func (s *Stats) RecordDelivered(workerID int) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalDelivered++
	ws := s.workers[workerID]
	ws.DeliveredTotal++
	s.workers[workerID] = ws
}

// вызывается, когда воркер завершает цикл из-за ctx.done
func (s *Stats) RecordUndelivered(workerID int) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalUndelivered++
	ws := s.workers[workerID]
	ws.UndeliveredTotal++
	s.workers[workerID] = ws
}

// возвращает копию данных, чтобы внешний код не менял внутреннее состояние stats
func (s *Stats) Snapshot() Snapshot {
	if s == nil {
		return Snapshot{}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return Snapshot{
		TotalDelivered:   s.totalDelivered,
		TotalUndelivered: s.totalUndelivered,
		Workers:          maps.Clone(s.workers),
	}
}
