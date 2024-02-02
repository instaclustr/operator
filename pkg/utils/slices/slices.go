package slices

func Equals[S ~[]T, T comparable](s1, s2 S) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}

	return true
}

func EqualsPtr[S ~[]*T, T comparable](s1, s2 S) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		if *s1[i] != *s2[i] {
			return false
		}
	}

	return true
}

func EqualsUnordered[S ~[]T, T comparable](s1, s2 S) bool {
	if len(s1) != len(s2) {
		return false
	}

	mapS1 := map[T]int{}
	for _, el := range s1 {
		mapS1[el]++
	}

	for _, el := range s2 {
		count, ok := mapS1[el]
		if !ok || count == 0 {
			return false
		}

		mapS1[el]--
	}

	return true
}

func EqualsUnorderedPtr[S ~[]*T, T comparable](s1, s2 S) bool {
	if len(s1) != len(s2) {
		return false
	}

	mapS1 := map[T]int{}
	for _, el := range s1 {
		mapS1[*el]++
	}

	for _, el := range s2 {
		count, ok := mapS1[*el]
		if !ok || count == 0 {
			return false
		}

		mapS1[*el]--
	}

	return true
}
