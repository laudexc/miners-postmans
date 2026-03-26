package simulation

import (
	"context"
	"sync"

	"edu/internal/job"
)

// deliverer описывает исполнителя задач доставки
type Deliverer interface {
	Name() string
	Deliver(ctx context.Context, j job.DeliveryJob) error
}

// deliveryresult хранит результат симуляции доставки за день
type DeliveryResult struct {
	Done      int
	Failed    int
	Cancelled int
}

// startDeliveryDay запускает симуляцию дня доставки через пул почтальонов
func StartDeliveryDay(ctx context.Context, deliverers []Deliverer, jobs []job.DeliveryJob) DeliveryResult {
	if len(deliverers) == 0 || len(jobs) == 0 {
		return DeliveryResult{}
	}

	jobsCh := make(chan job.DeliveryJob)
	resultCh := make(chan error, len(jobs))
	wg := &sync.WaitGroup{}

	// поднимаем пул исполнителей, каждый воркер читает задачи из общего канала
	for _, d := range deliverers {
		worker := d
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
					err := worker.Deliver(ctx, j)
					resultCh <- err
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
	result := DeliveryResult{
		Cancelled: len(jobs) - enqueued,
	}
	for err := range resultCh {
		if err != nil {
			result.Failed++
			continue
		}
		result.Done++
	}

	return result
}
