package postman

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// цикл одного почтальона с записью результата каждой итерации в общие stats
func Postman(
	ctx context.Context,
	wg *sync.WaitGroup,
	transferPoint chan<- string,
	n int,
	mail string,
	stats *Stats,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			// смена завершилась до следующей доставки, фиксируем недоставку
			stats.RecordUndelivered(n)
			fmt.Printf("Рабочий день почтальона №%d закончен!\n", n)
			return
		case <-time.After(250 * time.Millisecond):
		}

		select {
		case <-ctx.Done():
			// смена завершилась до передачи письма в канал, фиксируем недоставку
			stats.RecordUndelivered(n)
			fmt.Printf("Рабочий день почтальона №%d закончен!\n", n)
			return
		case transferPoint <- mail:
			// письмо успешно передано в канал, учитываем как доставленное
			stats.RecordDelivered(n)
			fmt.Println("Почтальон №", n, "передал письмо:", mail)
		}
	}
}

// запускает всех почтальонов с общими stats и закрывает канал после остановки всех воркеров
func PostmanPool(ctx context.Context, postmanCount int, stats *Stats) <-chan string {
	mailTransferPoint := make(chan string)
	wg := &sync.WaitGroup{}

	for i := 1; i <= postmanCount; i++ {
		wg.Add(1)
		go Postman(ctx, wg, mailTransferPoint, i, postmanToMail(i), stats)
	}

	go func() {
		wg.Wait()
		close(mailTransferPoint)
	}()

	return mailTransferPoint
}

func postmanToMail(postmanNumber int) string {
	ptm := map[int]string{
		1: "Семейный привет",
		2: "Приглашение от друга",
		3: "Информация из автосервиса",
	}

	mail, ok := ptm[postmanNumber]
	if !ok {
		return "Лотерея"
	}

	return mail
}
