package miner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"edu/internal/job"
)

// цикл одного шахтёра, который обрабатывает полученный список задач
func Miner(
	ctx context.Context,
	wg *sync.WaitGroup,
	transferPoint chan<- int,
	minerID int,
	jobs []job.MiningJob,
	stats *Stats,
) {
	defer wg.Done()

	for _, miningJob := range jobs {
		fmt.Printf("шахтёр №%d начал задачу добычи #%d\n", minerID, miningJob.ID)

		select {
		case <-ctx.Done():
			// если смена завершилась до добычи, считаем задачу неуспешной
			stats.RecordFailedAttempt(minerID)
			fmt.Printf("рабочий день шахтёра №%d закончен\n", minerID)
			return
		case <-time.After(1 * time.Second):
			fmt.Printf("шахтёр №%d добыл уголь по задаче #%d: %d\n", minerID, miningJob.ID, miningJob.Amount)
		}

		select {
		case <-ctx.Done():
			// если смена завершилась до передачи в канал, задача тоже неуспешная
			stats.RecordFailedAttempt(minerID)
			fmt.Printf("рабочий день шахтёра №%d закончен\n", minerID)
			return
		case transferPoint <- miningJob.Amount:
			// успешная передача результата задачи в общий канал
			stats.RecordMined(minerID, miningJob.Amount)
			fmt.Printf("шахтёр №%d передал уголь по задаче #%d: %d\n", minerID, miningJob.ID, miningJob.Amount)
		}
	}
}

// запускает всех шахтёров и отдаёт каждому его список задач из генератора квот
func MinerPool(
	ctx context.Context,
	minerCount int,
	jobsByMiner map[int][]job.MiningJob,
	stats *Stats,
) <-chan int {
	coalTransferPoint := make(chan int)
	wg := &sync.WaitGroup{}

	for minerID := 1; minerID <= minerCount; minerID++ {
		wg.Add(1)
		go Miner(ctx, wg, coalTransferPoint, minerID, jobsByMiner[minerID], stats)
	}

	go func() {
		wg.Wait()
		close(coalTransferPoint)
	}()

	return coalTransferPoint
}
