package main

import (
	"errors"
)

// This is a file with various utilities

// A utility function to swap two positions in an int array.
func swapInt(arr []int, i int, j int) []int {
	tmp := arr[i]
	arr[i] = arr[j]
	arr[j] = tmp
	return arr
}

// A utility function to get the image of the map.
func getImage(fun map[int]int) map[int]bool {
	var res = make(map[int]bool)
	for _, v := range fun {
		res[v] = true
	}
	return res
}

// A helper function to generate permutations using Heap's algorithm.
func permutationsHelper(res [][]int, arr []int, k int) [][]int {
	if k == 1 {
		newPerm := make([]int, len(arr))
		copy(newPerm, arr)
		res = append(res, newPerm)
	} else {
		res = permutationsHelper(res, arr, k-1)
		for i := 0; i < k-1; i++ {
			if k%2 == 0 {
				swapInt(arr, i, k-1)
			} else {
				swapInt(arr, 0, k-1)
			}
			res = permutationsHelper(res, arr, k-1)
		}
	}

	return res
}

// A utility function to generate all the permutations of a list.
// Underneath uses Heap's algorithm.
func generatePermutations(arr []int) [][]int {
	res := [][]int{}
	if len(arr) > 0 {
		res = permutationsHelper(res, arr, len(arr))
	}
	return res
}

func combinationHelper(res [][]int, origArr []int, partArr []int,
	k int, sz int, idx int, jdx int) [][]int {
	if k == sz {
		newComb := make([]int, sz)
		copy(newComb, partArr)
		res = append(res, newComb)
	} else if k < sz {
		for i := idx; i < len(origArr); i++ {
			partArr[k] = origArr[i]
			res = combinationHelper(res, origArr, partArr, k+1, sz, i+1, jdx+1)
		}
	}
	return res
}

func generateCombinations(arr []int, k int) [][]int {
	res := [][]int{}
	var partArr = make([]int, k)
	res = combinationHelper(res, arr, partArr, 0, k, 0, 0)
	return res
}

func createSimpleArray(n int, start int) []int {
	var res []int = make([]int, n)
	for i := 0; i < n; i++ {
		res[i] = start
		start++
	}
	return res
}

// use -1 for empty mapping
// calls a callback for each subset
// the limit is 64 elements as limited by uint64 size.
func generateSubsets(arr []int, callback func([]int), toStop func() bool) {
	n := len(arr)
	for mask := uint64(0); (mask < (1 << uint64(n))) && !toStop(); mask++ {
		var subset []int
		for i := range arr {
			if mask&(1<<uint64(i)) != 0 {
				subset = append(subset, arr[i])
			}
		}
		callback(subset)
	}
}

// Generate all mappings, both full and partial mappings
// Note: The complexity is very high
//
// Params:
//	n = the size of a mapping, the mapping is generates with input set = [1..n]
func generateMappings(n int, callback func(map[int]int), toStop func() bool) {
	var arr []int = createSimpleArray(n, 1)
	generateMappingsWithInp(arr, callback, toStop)
}

func generateMappingsWithInp(origArr []int, callback func(map[int]int), toStop func() bool) {
	generateSubsets(origArr,
		func(arr []int) {
			n := len(arr)
			var combs = generateCombinations(origArr, n)
		outer:
			for i := range combs {
				perms := generatePermutations(combs[i])
				for j := range perms {
					if toStop() {
						break outer
					}
					var rho = make(map[int]int)
					for k := range arr {
						rho[arr[k]] = perms[j][k]
					}

					callback(rho)

				}

			}
		},
		toStop)
}

func reverseMap(fun map[int]int) (map[int]int, error) {
	var res = make(map[int]int)
	for k := range fun {
		if _, ok := res[fun[k]]; !ok {
			res[fun[k]] = k
		} else {
			return res, errors.New("the function if not bijective")
		}
	}
	return res, nil
}

func reverseMapIntString(fun map[int]string) (map[string]int, error) {
	var res = make(map[string]int)
	for k := range fun {
		if _, ok := res[fun[k]]; !ok {
			res[fun[k]] = k
		} else {
			return res, errors.New("the function if not bijective")
		}
	}
	return res, nil
}

func reverseMapStringString(fun map[string]string) (map[string]string, error) {
	var res = make(map[string]string)
	for k := range fun {
		if _, ok := res[fun[k]]; !ok {
			res[fun[k]] = k
		} else {
			return res, errors.New("the function if not bijective")
		}
	}
	return res, nil
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
