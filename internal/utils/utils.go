package utils

import "math"

func CosineSimilarity(a, b []float32) float64 {
	dotProduct := float64(0)
	aMagnitude := float64(0)
	bMagnitude := float64(0)

	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		aMagnitude += math.Pow(float64(a[i]), 2)
		bMagnitude += math.Pow(float64(b[i]), 2)
	}

	return dotProduct / (math.Sqrt(aMagnitude) * math.Sqrt(bMagnitude))
}
