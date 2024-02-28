package utils

const (
	base int64 = 10
)

func Valid(number int64) bool {
	return (number%base+checksum(number/base))%base == 0
}

func checksum(number int64) int64 {
	var luhn int64

	for i := 0; number > 0; i++ {
		cur := number % base

		if i%2 == 0 { // even
			cur *= 2
			if cur > base-1 {
				cur = cur%base + cur/base
			}
		}

		luhn += cur
		number /= base
	}
	return luhn % base
}
