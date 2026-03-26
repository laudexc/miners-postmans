package job

// описывает одну задачу добычи угля для конкретного шахтёра
type MiningJob struct {
	ID      int
	MinerID int
	Amount  int
}
