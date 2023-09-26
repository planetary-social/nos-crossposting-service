package internal

import "math/rand"

func CopySlice[T any](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)
	return result
}

func RandomElement[T any](slice []T) T {
	return slice[rand.Intn(len(slice))]
}

func BatchesFromSlice[T any](slice []T, n int) [][]T {
	var result [][]T
	i := 0
	for {
		start := i * n
		if start >= len(slice) {
			break
		}

		end := start + n
		if l := len(slice); end >= l {
			end = l
		}

		result = append(result, slice[start:end])
		i++
	}
	return result
}
