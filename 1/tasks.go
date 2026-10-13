package lesson2

// FizzBuzz возвращает FizzBuzz, Fizz, Buzz или пустую строку для обычного числа.
// Для n <= 0 возвращает invalid.
func FizzBuzz(n int) string {
	switch {
	case n <= 0:
		return "invalid"
	case n%15 == 0:
		return "FizzBuzz"
	case n%3 == 0:
		return "Fizz"
	case n%5 == 0:
		return "Buzz"
	}

	return ""
}

// Sum возвращает сумму. Для пустого среза и nil возвращает 0.
func Sum(values []int64) int64 {
	n := len(values)
	if n == 0 {
		return 0
	}

	sum := int64(0)
	for i := 0; i < n; i++ {
		sum += values[i]
	}

	return sum
}

// Min возвращает минимум. Для пустого среза и nil возвращает 0 по условию задачи.
func Min(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}

	min_val := values[0]
	for _, val := range values {
		if min_val > val {
			min_val = val
		}
	}

	return min_val
}

// Max возвращает максимум. Для пустого среза и nil возвращает 0 по условию задачи.
func Max(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}

	max_val := values[0]
	for _, val := range values {
		if max_val < val {
			max_val = val
		}
	}

	return max_val
}

// Average возвращает среднее как float64. Для пустого среза и nil возвращает 0.
func Average(values []int64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := int64(0)
	for _, val := range values {
		sum += val
	}

	return float64(sum) / float64(len(values))
}

// Category возвращает invalid, small, regular или large по сумме в копейках.
func Category(amount int64) string {
	switch {
	case amount < 0:
		return "invalid"
	case amount < 10000:
		return "small"
	case amount < 50000:
		return "regular"
	default:
		return "large"
	}
}

// RuneCount считает кодовые точки Unicode в строке через range.
func RuneCount(text string) int {
	count := 0
	for range text {
		count++
	}
	return count
}
