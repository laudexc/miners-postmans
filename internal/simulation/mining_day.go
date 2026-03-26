package simulation

import (
	"context"
	"sync"

	"edu/internal/job"
)

// miner описывает исполнителя задач добычи
type Miner interface {
	Name() string
	Mine(ctx context.Context, j job.MiningJob) error
}

// miningresult хранит результат симуляции добычи за день
type MiningResult struct {
	Done       int
	Failed     int
	Cancelled  int
	TotalMined int
}

// startMiningDay запускает симуляцию дня добычи через пул майнеров
func StartMiningDay(ctx context.Context, miners []Miner, jobs []job.MiningJob) MiningResult {
	if len(miners) == 0 || len(jobs) == 0 {
		return MiningResult{}
	}

	jobsCh := make(chan job.MiningJob)
	resultCh := make(chan miningJobResult, len(jobs))
	wg := &sync.WaitGroup{}

	// поднимаем пул исполнителей, каждый воркер читает задачи из общего канала
	for _, m := range miners {
		worker := m
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case j, ok := <-jobsCh:
					if !ok {
						return
					}
					err := worker.Mine(ctx, j)
					resultCh <- miningJobResult{
						job: j,
						err: err,
					}
				}
			}
		}()
	}

	// отправляем задачи в канал до конца списка или до отмены контекста
	enqueued := 0
enqueueLoop:
	for _, j := range jobs {
		select {
		case <-ctx.Done():
			break enqueueLoop
		case jobsCh <- j:
			enqueued++
		}
	}
	close(jobsCh)

	// ждём завершения всех воркеров, после этого закрываем канал результатов
	wg.Wait()
	close(resultCh)

	// собираем итоговый отчёт по успешным и неуспешным задачам
	result := MiningResult{
		Cancelled: len(jobs) - enqueued,
	}
	for r := range resultCh {
		if r.err != nil {
			result.Failed++
			continue
		}
		result.Done++
		result.TotalMined += r.job.Amount
	}

	return result
}

type miningJobResult struct {
	job job.MiningJob
	err error
}
