package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"edu/internal/miner"
	"edu/internal/postman"
)

func main() {
	// общая сумма угля, считанного из канала шахтёров
	var coal atomic.Int64

	// количество воркеров (шахтёров и почтальонов)
	numOfWorkers := 10

	// объекты статистики создаются один раз и разделяются между всеми воркерами
	minerStats := miner.NewStats()
	postmanStats := postman.NewStats()

	mtx := sync.Mutex{}
	var mails []string

	// отдельные контексты позволяют завершать смены шахтёров и почтальонов независимо
	minerContext, minerCancel := context.WithCancel(context.Background())
	postmanContext, postmanCancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(3 * time.Second)
		fmt.Println("------------>>> Рабочий день шахтёров окончен!")
		minerCancel()
	}()

	go func() {
		time.Sleep(6 * time.Second)
		fmt.Println("------------>>> Рабочий день почтальонов окончен!")
		postmanCancel()
	}()

	// пулы получают общие указатели на stats, и каждый воркер пишет туда свои результаты
	coalTransferPoint := miner.MinerPool(minerContext, numOfWorkers, minerStats)
	mailTransferPoint := postman.PostmanPool(postmanContext, numOfWorkers, postmanStats)

	initTime := time.Now()
	wg := &sync.WaitGroup{}

	// отдельная горутина читает канал угля и обновляет общую сумму
	wg.Go(func() {
		for v := range coalTransferPoint {
			coal.Add(int64(v))
		}
	})

	// отдельная горутина читает канал писем и складывает полученные значения
	wg.Go(func() {
		for v := range mailTransferPoint {
			mtx.Lock()
			mails = append(mails, v)
			mtx.Unlock()
		}
	})

	// ждём завершения читателей, к этому моменту пулы уже остановлены и stats стабильны
	wg.Wait()

	fmt.Println("----------------")
	fmt.Println("Суммарно добыто угля:", coal.Load())

	// snapshot возвращает безопасную копию данных для вывода отчёта
	minerSnapshot := minerStats.Snapshot()
	fmt.Printf("---> Статы шахтёров: всего добыто = %d, неудачных попыток добычи = %d\n", minerSnapshot.TotalMined, minerSnapshot.FailedAttempts)

	fmt.Println("----------------")
	mtx.Lock()
	fmt.Println("Суммарно получено писем:", len(mails))
	mtx.Unlock()

	// статистика почтальонов читается аналогично через копию snapshot
	postmanSnapshot := postmanStats.Snapshot()
	fmt.Printf("---> Статы почтальонов: доставлено писем = %d, не доставлено писем = %d\n", postmanSnapshot.TotalDelivered, postmanSnapshot.TotalUndelivered)

	fmt.Println("Затраченное время:", time.Since(initTime))
}
