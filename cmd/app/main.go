package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"edu/internal/generator"
	"edu/internal/miner"
	"edu/internal/postman"
	"edu/internal/salary"
)

func main() {
	// общая сумма угля, считанного из канала шахтёров
	var coal atomic.Int64

	// количество воркеров в каждой группе
	numOfWorkers := 1

	// квота определяет, сколько задач и какого объёма выдаётся воркерам за смену
	baseQuota := generator.Quota{
		MiningJobsPerMiner:     4,
		DeliveryJobsPerPostman: 18,
		BaseMiningAmount:       10,
	}
	quota := generator.AdjustQuotaByWorkers(baseQuota, numOfWorkers)

	// генератор формирует план работ по квоте для каждой группы
	miningJobsByMiner := generator.GenerateMiningJobs(numOfWorkers, quota)
	deliveryJobsByPostman := generator.GenerateDeliveryJobs(numOfWorkers, quota)

	// объекты статистики разделяются между всеми воркерами внутри своей группы
	minerStats := miner.NewStats()
	postmanStats := postman.NewStats()

	mtx := sync.Mutex{}
	var mails []string

	// отдельные контексты позволяют завершать смены шахтёров и почтальонов независимо
	minerContext, minerCancel := context.WithCancel(context.Background())
	postmanContext, postmanCancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(3 * time.Second)
		fmt.Println("------------>>> рабочий день шахтёров окончен")
		fmt.Println()
		minerCancel()
	}()

	go func() {
		time.Sleep(6 * time.Second)
		fmt.Println("------------>>> рабочий день почтальонов окончен")
		fmt.Println()
		postmanCancel()
	}()

	// пулы получают подготовленные задания и обрабатывают их параллельно
	coalTransferPoint := miner.MinerPool(minerContext, numOfWorkers, miningJobsByMiner, minerStats)
	mailTransferPoint := postman.PostmanPool(postmanContext, numOfWorkers, deliveryJobsByPostman, postmanStats)

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

	// ждём завершения чтения каналов, после этого можно безопасно читать stats
	wg.Wait()

	minerSnapshot := minerStats.Snapshot()
	postmanSnapshot := postmanStats.Snapshot()

	printBlockTitle("итоги выполнения")
	fmt.Printf("авто-квота: задач шахтёру=%d, задач почтальону=%d, базовый объём добычи=%d\n\n",
		quota.MiningJobsPerMiner,
		quota.DeliveryJobsPerPostman,
		quota.BaseMiningAmount,
	)
	fmt.Printf("---> статы шахтёров: всего добыто = %d, неуспешных задач = %d\n\n", minerSnapshot.TotalMined, minerSnapshot.FailedAttempts)
	fmt.Printf("---> статы почтальонов: доставлено = %d, недоставлено = %d\n\n", postmanSnapshot.TotalDelivered, postmanSnapshot.TotalUndelivered)

	// ставки для расчёта зарплаты задаются отдельно, чтобы их можно было легко менять
	minerRates := salary.MinerRates{
		BasePerUnit:    5.0,
		BonusPerUnit:   1.2,
		PenaltyPerUnit: 0.5,
	}
	postmanRates := salary.PostmanRates{
		BasePerLetter:    28,
		BonusPerLetter:   6,
		PenaltyPerLetter: 3,
	}

	// зарплата считается по плану задач и факту выполнения из статистики
	minerSalaryReport := salary.CalculateMinerReport(miningJobsByMiner, minerSnapshot, minerRates)
	postmanSalaryReport := salary.CalculatePostmanReport(deliveryJobsByPostman, postmanSnapshot, postmanRates)

	printBlockTitle("зарплата")
	fmt.Println("пояснение: план = назначено задач за смену, факт = реально выполнено до остановки смены")
	fmt.Println()

	fmt.Printf("фонд зарплаты шахтёров: %.2f\n\n", minerSalaryReport.Total)
	for _, worker := range minerSalaryReport.Workers {
		fmt.Printf(
			"шахтёр №%d\nплан=%d, факт=%d\nбаза=%.2f, бонус=%.2f\nштраф сырой=%.2f, штраф применённый=%.2f\nитого=%.2f\n\n",
			worker.WorkerID,
			worker.PlannedUnits,
			worker.DoneUnits,
			worker.BaseSalary,
			worker.Bonus,
			worker.PenaltyRaw,
			worker.PenaltyApplied,
			worker.FinalSalary,
		)
	}

	fmt.Printf("фонд зарплаты почтальонов: %.2f\n\n", postmanSalaryReport.Total)
	for _, worker := range postmanSalaryReport.Workers {
		fmt.Printf(
			"почтальон №%d\nплан=%d, факт=%d\nбаза=%.2f, бонус=%.2f\nштраф сырой=%.2f, штраф применённый=%.2f\nитого=%.2f\n\n",
			worker.WorkerID,
			worker.PlannedDeliveries,
			worker.DoneDeliveries,
			worker.BaseSalary,
			worker.Bonus,
			worker.PenaltyRaw,
			worker.PenaltyApplied,
			worker.FinalSalary,
		)
	}

	printBlockTitle("время")
	fmt.Printf("затраченное время: %s\n\n", time.Since(initTime))
}

// единый заголовок для читаемых блоков отчёта
func printBlockTitle(title string) {
	fmt.Println("----------------")
	fmt.Println(title)
	fmt.Println("----------------")
	fmt.Println()
}
