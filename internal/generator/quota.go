package generator

import "edu/internal/job"

// хранит настройки, сколько задач выдаётся воркерам за смену
type Quota struct {
	MiningJobsPerMiner     int
	DeliveryJobsPerPostman int
	BaseMiningAmount       int
}

// автоматически повышает или снижает квоту на одного воркера
// при меньшем числе воркеров квота на человека растёт, при большем — снижается
func AdjustQuotaByWorkers(base Quota, workerCount int) Quota {
	if workerCount <= 0 {
		return base
	}

	const referenceWorkers = 20

	result := base
	factor := float64(referenceWorkers) / float64(workerCount)
	// ограничиваем фактор, чтобы квота менялась плавно и не схлопывалась до 1 задачи
	if factor < 0.9 {
		factor = 0.9
	}
	if factor > 1.8 {
		factor = 1.8
	}

	// на одного шахтёра квота чуть масштабируется относительно размера команды
	result.MiningJobsPerMiner = scaleByFactor(
		base.MiningJobsPerMiner,
		factor,
		1,
		max(1, base.MiningJobsPerMiner*2),
	)
	// на одного почтальона квота тоже масштабируется, но остаётся в разумных границах
	result.DeliveryJobsPerPostman = scaleByFactor(
		base.DeliveryJobsPerPostman,
		factor,
		1,
		max(1, base.DeliveryJobsPerPostman*2),
	)

	// объём добычи в одной задаче держим стабильным
	result.BaseMiningAmount = clampInt(base.BaseMiningAmount, 5, max(5, base.BaseMiningAmount*2))

	return result
}

// создаёт задания добычи по квоте и раскладывает их по id шахтёров
func GenerateMiningJobs(minerCount int, q Quota) map[int][]job.MiningJob {
	result := make(map[int][]job.MiningJob, minerCount)
	if minerCount <= 0 || q.MiningJobsPerMiner <= 0 || q.BaseMiningAmount <= 0 {
		return result
	}

	jobID := 1
	for minerID := 1; minerID <= minerCount; minerID++ {
		workerJobs := make([]job.MiningJob, 0, q.MiningJobsPerMiner)
		for n := 1; n <= q.MiningJobsPerMiner; n++ {
			// объём задачи растёт от номера, чтобы задания были разной сложности
			amount := q.BaseMiningAmount * n
			workerJobs = append(workerJobs, job.MiningJob{
				ID:      jobID,
				MinerID: minerID,
				Amount:  amount,
			})
			jobID++
		}
		result[minerID] = workerJobs
	}

	return result
}

// создаёт задания доставки по квоте и раскладывает их по id почтальонов
func GenerateDeliveryJobs(postmanCount int, q Quota) map[int][]job.DeliveryJob {
	result := make(map[int][]job.DeliveryJob, postmanCount)
	if postmanCount <= 0 || q.DeliveryJobsPerPostman <= 0 {
		return result
	}

	jobID := 1
	for postmanID := 1; postmanID <= postmanCount; postmanID++ {
		workerJobs := make([]job.DeliveryJob, 0, q.DeliveryJobsPerPostman)
		for n := 1; n <= q.DeliveryJobsPerPostman; n++ {
			workerJobs = append(workerJobs, job.DeliveryJob{
				ID:        jobID,
				PostmanID: postmanID,
				MailText:  mailTextByIndex(n),
				Address:   addressByIndex(postmanID, n),
				Priority:  priorityByIndex(n),
			})
			jobID++
		}
		result[postmanID] = workerJobs
	}

	return result
}

// ограничивает значение диапазоном
func clampInt(v int, minValue int, maxValue int) int {
	if v < minValue {
		return minValue
	}
	if v > maxValue {
		return maxValue
	}
	return v
}

// масштабирует базовое значение с округлением до ближайшего целого
func scaleByFactor(base int, factor float64, minValue int, maxValue int) int {
	if base <= 0 {
		return minValue
	}

	scaled := int(float64(base)*factor + 0.5)
	return clampInt(scaled, minValue, maxValue)
}

// возвращает шаблонный текст письма по номеру задачи
func mailTextByIndex(n int) string {
	switch n % 3 {
	case 1:
		return "семейный привет"
	case 2:
		return "приглашение от друга"
	default:
		return "информация из автосервиса"
	}
}

// возвращает простой адрес для демонстрации маршрута доставки
func addressByIndex(postmanID int, n int) string {
	return "улица " + itoa(postmanID) + ", дом " + itoa(10+n)
}

// priorityByIndex задаёт приоритет письма в диапазоне 1..3
func priorityByIndex(n int) int {
	return (n % 3) + 1
}

// чтобы не тянуть fmt для одной операции
func itoa(v int) string {
	if v == 0 {
		return "0"
	}

	sign := ""
	if v < 0 {
		sign = "-"
		v = -v
	}

	buf := [20]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + (v % 10))
		v /= 10
	}

	return sign + string(buf[i:])
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
