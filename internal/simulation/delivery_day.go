package simulation

import (
	"context"
	// "sync"

	"edu/internal/job"
)

type Deliverer interface {
	Name() string
	Deliverer(ctx context.Context, j job.DeliveryJob) error
}

type DeliveryResult struct {
	Done   int
	Failed int
}

// func StartDeliveryDay(ctx context.Context, postmanCount int, 
// 	jobs []job.DeliveryJob) DeliveryResult {
// 	if postmanCount == 0 || len(jobs) == 0 {
// 		return DeliveryResult{}
// 	}

// 	jobsCh := make(chan job.DeliveryJob)
// 	resultCh := make(chan error, len(jobs))
// 	wg := &sync.WaitGroup{}

// 	for _, w := range deliverers {
// 		worker := w
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			for j := range jobsCh {
// 				resultCh <- worker.Deliverer(ctx, j)
// 			}
// 		}()
// 	}
	
// }
