package postman

import (
	"context"
	"fmt"
	"sync"
	"time"

	"edu/internal/job"
)

// цикл одного почтальона, который обрабатывает полученный список задач
func Postman(
	ctx context.Context,
	wg *sync.WaitGroup,
	transferPoint chan<- string,
	postmanID int,
	jobs []job.DeliveryJob,
	stats *Stats,
) {
	defer wg.Done()

	for _, deliveryJob := range jobs {
		fmt.Printf("почтальон №%d начал задачу доставки #%d\n", postmanID, deliveryJob.ID)

		select {
		case <-ctx.Done():
			// если смена завершилась до доставки, считаем задачу недоставленной
			stats.RecordUndelivered(postmanID)
			fmt.Printf("рабочий день почтальона №%d закончен\n", postmanID)
			return
		case <-time.After(250 * time.Millisecond):
		}

		message := fmt.Sprintf("письмо #%d [%s], адрес: %s, приоритет: %d", deliveryJob.ID, deliveryJob.MailText, deliveryJob.Address, deliveryJob.Priority)
		select {
		case <-ctx.Done():
			// если смена завершилась до передачи письма в канал, фиксируем недоставку
			stats.RecordUndelivered(postmanID)
			fmt.Printf("рабочий день почтальона №%d закончен\n", postmanID)
			return
		case transferPoint <- message:
			// письмо успешно передано в канал, считаем задачу доставленной
			stats.RecordDelivered(postmanID)
			fmt.Printf("почтальон №%d передал письмо по задаче #%d\n", postmanID, deliveryJob.ID)
		}
	}
}

// запускает всех почтальонов и отдаёт каждому его список задач из генератора квот
func PostmanPool(
	ctx context.Context,
	postmanCount int,
	jobsByPostman map[int][]job.DeliveryJob,
	stats *Stats,
) <-chan string {
	mailTransferPoint := make(chan string)
	wg := &sync.WaitGroup{}

	for postmanID := 1; postmanID <= postmanCount; postmanID++ {
		wg.Add(1)
		go Postman(ctx, wg, mailTransferPoint, postmanID, jobsByPostman[postmanID], stats)
	}

	go func() {
		wg.Wait()
		close(mailTransferPoint)
	}()

	return mailTransferPoint
}
