package homework

// PairInput соответствует объекту nums/target из условия контеста.
type PairInput struct {
	Nums   []int
	Target int
}

// longestPrefix возвращает самый длинный общий префикс строк.
func longestPrefix(words []string) string {
	if len(words) == 0 || words[0] == "" {
		return ""
	}

	prefix := []rune(words[0])

	for _, str := range words[1:] {
		i := 0
		for _, sym := range str {
			if i >= len(prefix) || sym != prefix[i] {
				break
			}
			i++
		}
		prefix = prefix[:i]
	}

	return string(prefix)
}

// pairIndices возвращает индексы двух разных элементов по возрастанию.
func pairIndices(input PairInput) [2]int {
	for i, fst_num := range input.Nums {
		for j, snd_num := range input.Nums[i+1:] {
			if fst_num+snd_num == input.Target {
				return [2]int{i, j + i + 1}
			}
		}
	}

	return [2]int{0, 0}
}

// hasDuplicate сообщает, встречается ли какое-либо значение повторно.
func hasDuplicate(nums []int) bool {
	for i, fst_num := range nums {
		for _, snd_num := range nums[i+1:] {
			if fst_num == snd_num {
				return true
			}
		}
	}

	return false
}

// compactSorted создаёт независимый слайс уникальных значений.
func compactSorted(nums []int) []int {
	if len(nums) == 0 {
		return nil
	}

	result := make([]int, 0)
	result = append(result, nums[0])
	ind_res := 0

	for _, num := range nums {
		if result[ind_res] != num {
			result = append(result, num)
			ind_res++
		}
	}

	return result
}

// balancedBrackets проверяет пары и порядок закрытия скобок.
func balancedBrackets(text string) bool {
	stack := make([]byte, 0)

	for _, sym := range text {
		switch sym {
		case '(', '{', '[':
			stack = append(stack, byte(sym))
		case ')':
			if len(stack) == 0 || stack[len(stack)-1] != '(' {
				return false
			}
			stack = stack[:len(stack)-1]
		case '}':
			if len(stack) == 0 || stack[len(stack)-1] != '{' {
				return false
			}
			stack = stack[:len(stack)-1]
		case ']':
			if len(stack) == 0 || stack[len(stack)-1] != '[' {
				return false
			}
			stack = stack[:len(stack)-1]
		default:
			return false
		}
	}

	return len(stack) == 0
}
