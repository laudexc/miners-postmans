package salary

import (
	"sort"

	"edu/internal/job"
	"edu/internal/miner"
	"edu/internal/postman"
)

// ставки для расчёта зарплаты шахтёров
type MinerRates struct {
	BasePerUnit    float64
	BonusPerUnit   float64
	PenaltyPerUnit float64
}

// ставки для расчёта зарплаты почтальонов
type PostmanRates struct {
	BasePerLetter    float64
	BonusPerLetter   float64
	PenaltyPerLetter float64
}

// детализация зарплаты по одному шахтёру
type MinerWorkerSalary struct {
	WorkerID       int
	PlannedUnits   int
	DoneUnits      int
	BaseSalary     float64
	Bonus          float64
	PenaltyRaw     float64
	PenaltyApplied float64
	FinalSalary    float64
	FailedAttempts int
}

// детализация зарплаты по одному почтальону
type PostmanWorkerSalary struct {
	WorkerID          int
	PlannedDeliveries int
	DoneDeliveries    int
	BaseSalary        float64
	Bonus             float64
	PenaltyRaw        float64
	PenaltyApplied    float64
	FinalSalary       float64
	UndeliveredTotal  int
}

// расчёт зарплаты по всем шахтёрам
type MinerReport struct {
	Workers []MinerWorkerSalary
	Total   float64
}

// расчёт зарплаты по всем почтальонам
type PostmanReport struct {
	Workers []PostmanWorkerSalary
	Total   float64
}

// считает зарплату шахтёров по плану задач и фактическим результатам
func CalculateMinerReport(
	jobsByMiner map[int][]job.MiningJob,
	stats miner.Snapshot,
	rates MinerRates,
) MinerReport {
	ids := collectMinerIDs(jobsByMiner, stats)
	report := MinerReport{
		Workers: make([]MinerWorkerSalary, 0, len(ids)),
	}

	for _, workerID := range ids {
		planned := plannedMiningUnits(jobsByMiner[workerID])
		workerStats := stats.Workers[workerID]
		done := workerStats.MinedTotal

		base := float64(done) * rates.BasePerUnit
		bonus := calcBonus(done, planned, rates.BonusPerUnit)
		penaltyRaw := calcPenalty(done, planned, rates.PenaltyPerUnit)
		penaltyApplied := applyPenaltyCap(base, penaltyRaw)
		final := base + bonus - penaltyApplied
		if final < 0 {
			final = 0
		}

		workerReport := MinerWorkerSalary{
			WorkerID:       workerID,
			PlannedUnits:   planned,
			DoneUnits:      done,
			BaseSalary:     base,
			Bonus:          bonus,
			PenaltyRaw:     penaltyRaw,
			PenaltyApplied: penaltyApplied,
			FinalSalary:    final,
			FailedAttempts: workerStats.FailedAttempts,
		}
		report.Workers = append(report.Workers, workerReport)
		report.Total += final
	}

	return report
}

// считает зарплату почтальонов по плану задач и фактическим результатам
func CalculatePostmanReport(
	jobsByPostman map[int][]job.DeliveryJob,
	stats postman.Snapshot,
	rates PostmanRates,
) PostmanReport {
	ids := collectPostmanIDs(jobsByPostman, stats)
	report := PostmanReport{
		Workers: make([]PostmanWorkerSalary, 0, len(ids)),
	}

	for _, workerID := range ids {
		planned := len(jobsByPostman[workerID])
		workerStats := stats.Workers[workerID]
		done := workerStats.DeliveredTotal

		base := float64(done) * rates.BasePerLetter
		bonus := calcBonus(done, planned, rates.BonusPerLetter)
		penaltyRaw := calcPenalty(done, planned, rates.PenaltyPerLetter)
		penaltyApplied := applyPenaltyCap(base, penaltyRaw)
		final := base + bonus - penaltyApplied
		if final < 0 {
			final = 0
		}

		workerReport := PostmanWorkerSalary{
			WorkerID:          workerID,
			PlannedDeliveries: planned,
			DoneDeliveries:    done,
			BaseSalary:        base,
			Bonus:             bonus,
			PenaltyRaw:        penaltyRaw,
			PenaltyApplied:    penaltyApplied,
			FinalSalary:       final,
			UndeliveredTotal:  workerStats.UndeliveredTotal,
		}
		report.Workers = append(report.Workers, workerReport)
		report.Total += final
	}

	return report
}

// считает суммарный план добычи по задачам одного шахтёра
func plannedMiningUnits(jobs []job.MiningJob) int {
	total := 0
	for _, j := range jobs {
		total += j.Amount
	}
	return total
}

// собирает список id шахтёров из плана и из фактической статистики
func collectMinerIDs(jobsByMiner map[int][]job.MiningJob, stats miner.Snapshot) []int {
	set := make(map[int]struct{}, len(jobsByMiner)+len(stats.Workers))
	for workerID := range jobsByMiner {
		set[workerID] = struct{}{}
	}
	for workerID := range stats.Workers {
		set[workerID] = struct{}{}
	}

	ids := make([]int, 0, len(set))
	for workerID := range set {
		ids = append(ids, workerID)
	}
	sort.Ints(ids)

	return ids
}

// собирает список id почтальонов из плана и из фактической статистики
func collectPostmanIDs(jobsByPostman map[int][]job.DeliveryJob, stats postman.Snapshot) []int {
	set := make(map[int]struct{}, len(jobsByPostman)+len(stats.Workers))
	for workerID := range jobsByPostman {
		set[workerID] = struct{}{}
	}
	for workerID := range stats.Workers {
		set[workerID] = struct{}{}
	}

	ids := make([]int, 0, len(set))
	for workerID := range set {
		ids = append(ids, workerID)
	}
	sort.Ints(ids)

	return ids
}

// ограничивает штраф 50% от базы, чтобы отчёт был реалистичнее для коротких смен
func applyPenaltyCap(base float64, rawPenalty float64) float64 {
	if base <= 0 || rawPenalty <= 0 {
		return 0
	}

	maxPenalty := base * 0.5
	if rawPenalty > maxPenalty {
		return maxPenalty
	}

	return rawPenalty
}
