package salary

// считает бонус за перевыполнение плана
func calcBonus(done int, planned int, bonusRate float64) float64 {
	if bonusRate <= 0 {
		return 0
	}

	extra := done - planned
	if extra <= 0 {
		return 0
	}

	return float64(extra) * bonusRate
}
