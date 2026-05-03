package main

func calculateMempoolSummary(mempool map[string]MempoolTx) MempoolSummary {
	summary := MempoolSummary{
		FeeDistribution: map[string]int{
			"low":    0,
			"medium": 0,
			"high":   0,
		},
	}

	txCount := len(mempool)

	if txCount == 0 {
		return summary
	}

	var totalFeeRate float64
	var minFeeRate float64
	var maxFeeRate float64

	first := true

	for _, tx := range mempool {
		if tx.VSize == 0 {
			continue
		}

		feeSats := tx.Fees.Base * 100_000_000
		feeRate := feeSats / float64(tx.VSize)

		summary.TotalVSize += tx.VSize
		totalFeeRate += feeRate

		if first {
			minFeeRate = feeRate
			maxFeeRate = feeRate
			first = false
		}

		if feeRate < minFeeRate {
			minFeeRate = feeRate
		}

		if feeRate > maxFeeRate {
			maxFeeRate = feeRate
		}

		switch {
		case feeRate < 10:
			summary.FeeDistribution["low"]++
		case feeRate < 50:
			summary.FeeDistribution["medium"]++
		default:
			summary.FeeDistribution["high"]++
		}
	}

	summary.TxCount = txCount
	summary.AvgFeeRate = totalFeeRate / float64(txCount)
	summary.MinFeeRate = minFeeRate
	summary.MaxFeeRate = maxFeeRate

	return summary
}
