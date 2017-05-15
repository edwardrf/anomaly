package anomaly

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/gonum/stat"
	"github.com/mjibson/go-dsp/fft"
)

func Detect(data []float64) []int {
	// Iteratively remove data that is outside of the normal range to get a true mean and stdDev
	var mean, stdDev float64
	var anom []int

	// Find the min and max of the data
	min, max := data[0], data[0]
	for _, d := range data[1:] {
		if d > max {
			max = d
		}
		if d < min {
			min = d
		}
	}

	for {
		mean, stdDev = stat.MeanStdDev(data, nil)
		// Count data outside of 3 sigmas as abnormal
		removed := 0
		for i, d := range data {
			if math.Abs(d-mean) > 3*stdDev {
				anom = append(anom, i)
				data[i] = mean
				removed++
			}
		}
		if removed == 0 {
			break
		}
	}

	for i := range data {
		data[i] = data[i] - mean
	}

	dftCplx := fft.FFTReal(data)
	dft := make([]float64, len(dftCplx)/2+1)
	for i := range dft {
		dft[i] = cmplx.Abs(dftCplx[i])
		//if i < 100 {
		//fmt.Println(dft[i])
		//}
	}

	// Find the principle frequency
	freqStart := 2
	maxIdx := freqStart
	maxMag := 0.0
	sumV := 0.0
	for i := freqStart; i < len(dft)/2; i++ {
		v := dft[i] * math.Log(float64(i)) // times the signal to a log function to reduce significance of start of the frequencies
		sumV += v
		if v > maxMag {
			maxMag = v
			maxIdx = i
		}
	}

	// The data contains no periodic signal, maybe spiked data, aggregation might be needed
	if dft[maxIdx-1]+dft[maxIdx]+dft[maxIdx+1] == 0 {
		return anom
	}

	// If no period has been found, consider the data non-periodic
	sigRatio := maxMag / (sumV / float64(len(dft)/2-freqStart))
	fmt.Println("SigRatio :", sigRatio)
	if sigRatio < 6 { // 6 is Arbitrarily chosen value
		return anom
	}

	// Calculate the number of principle period: Average of the 3 points: before, max, after to get a better estimate
	n := (float64(maxIdx-1)*dft[maxIdx-1] + float64(maxIdx)*dft[maxIdx] + float64(maxIdx+1)*dft[maxIdx+1]) / (dft[maxIdx-1] + dft[maxIdx] + dft[maxIdx+1])
	//n = math.Floor(n + 0.5)

	// Then the period
	t := float64(len(data)) / n
	t = math.Floor(t + 0.5)

	fmt.Println("n, t ", n, t)
	hist := make([][]float64, int(math.Ceil(t)))
	weights := make([][]float64, int(math.Ceil(t)))

	for i, d := range data {
		k := int(math.Floor(math.Mod(float64(i), t)))
		//fmt.Println("i, k", i, k)
		hist[k] = append(hist[k], d)
		weight := 1.0
		if math.Abs(d) > 3*stdDev { // Ignore points already outside of the normal range
			weight = 0.0
		}
		weights[k] = append(weights[k], weight)
	}

	idxs := make([]int, len(hist))
	for i, d := range data {
		k := int(math.Floor(math.Mod(float64(i), t)))
		idx := idxs[k]
		idxs[k]++

		//Calculate the mean and standard dev without the current value
		oldWeight := weights[k][idx]
		weights[k][idx] = 0.0
		omean, osdev := stat.MeanStdDev(hist[k], weights[k])
		weights[k][idx] = oldWeight

		if math.IsNaN(osdev) { // Ignore cases where there is a single data point in the silo
			continue
		}

		//Very special case where stddev of a set of data is 0, use 0.1 percent of the range of the data as acceptable range
		//if osdev == 0 {
		//osdev = (max - min) * 0.001
		//}

		diff := math.Abs(d - omean)
		if diff > osdev*3 {
			//fmt.Println(diff, osdev*4)
			//fmt.Println(hist[k], idx)
			anom = append(anom, i)
		}
		fmt.Println(d, ",", mean+omean+osdev*3, ",", mean+omean-osdev*3)
		//fmt.Println(d, ",", mean+omean, ",", mean+omean+osdev*3, ",", mean+omean-osdev*3)
	}
	return anom
}
