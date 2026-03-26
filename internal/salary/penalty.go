package salary

// считает штраф за недовыполнение плана
func calcPenalty(done int, planned int, penaltyRate float64) float64 {
	if penaltyRate <= 0 {
		return 0
	}

	missing := planned - done
	if missing <= 0 {
		return 0
	}

	return float64(missing) * penaltyRate
}
