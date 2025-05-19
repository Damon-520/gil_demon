/**
   @author: tony.zhao
   @data : 2022/12/12 22:00
   @desc:
**/

package sinterx

// 并集
func union(slice1, slice2 []string) []string {
	m := make(map[string]int)
	for _, v := range slice1 {
		m[v]++
	}

	for _, v := range slice2 {
		times, _ := m[v]
		if times == 0 {
			slice1 = append(slice1, v)
		}
	}

	return slice1
}

// 交集
func intersenct(slice1, slice2 []string) []string {
	m := make(map[string]int)
	nn := make([]string, 0)
	for _, v := range slice1 {
		m[v]++
	}

	for _, v := range slice2 {
		times, _ := m[v]
		if times == 1 {
			nn = append(nn, v)
		}
	}
	return nn
}

// 差集
func difference(slice1, slice2 []string) []string {
	m := make(map[string]int)
	nn := make([]string, 0)
	inter := intersenct(slice1, slice2)
	for _, v := range inter {
		m[v]++
	}

	for _, v := range slice1 {
		times, _ := m[v]
		if times == 0 {
			nn = append(nn, v)
		}
	}
	return nn
}
