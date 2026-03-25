package miner

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// цикл одного шахтёра с записью результата каждой итерации в общие stats
func Miner(
	ctx context.Context,
	wg *sync.WaitGroup,
	transferPoint chan<- int,
	n int,
	power int,
	stats *Stats,
) {
	defer wg.Done()

	for {
		fmt.Println("Шахтёр №", n, "начал добывать уголь")

		select {
		case <-ctx.Done():
			// смена завершилась до окончания добычи, считаем это неуспешной попыткой
			stats.RecordFailedAttempt(n)
			fmt.Printf("Рабочий день шахтёра №%d закончен!\n", n)
			return
		case <-time.After(1 * time.Second):
			fmt.Println("Шахтёр №", n, "добыл уголь:", power)
		}

		select {
		case <-ctx.Done():
			// смена завершилась до передачи угля в канал, это тоже неуспешная попытка
			stats.RecordFailedAttempt(n)
			fmt.Printf("Рабочий день шахтёра №%d закончен!\n", n)
			return
		case transferPoint <- power:
			// уголь успешно передан в канал, учитываем как успешную добычу
			stats.RecordMined(n, power)
			fmt.Println("Шахтёр №", n, "передал уголь:", power)
		}
	}
}

// запускает всех шахтёров с общими stats и закрывает канал после остановки всех воркеров
func MinerPool(ctx context.Context, minerCount int, stats *Stats) <-chan int {
	coalTransferPoint := make(chan int)
	wg := &sync.WaitGroup{}

	for i := 1; i <= minerCount; i++ {
		wg.Add(1)
		go Miner(ctx, wg, coalTransferPoint, i, i*10, stats)
	}

	go func() {
		wg.Wait()
		close(coalTransferPoint)
	}()

	return coalTransferPoint
}
