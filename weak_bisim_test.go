package main

import (
	"fmt"
	"testing"

	"github.com/yungene/pifra"
)

func TestFloydWarshall(t *testing.T) {
	testAllVerticesReachability(t, floydWarshall)
}

func TestDfsClosure(t *testing.T) {
	testAllVerticesReachability(t, dfsClosure)
}

func testAllVerticesReachability(t *testing.T, f func(pifra.Lts) ([][]bool, map[int]int, map[int]int)) {
	var states = map[int]pifra.Configuration{
		1: {},
		2: {},
		3: {},
		4: {},
	}
	var transitions []pifra.Transition = []pifra.Transition{
		{1, 2,
			pifra.Label{
				Symbol: pifra.Symbol{pifra.SymbolTypFreshInput, 0}}},
		{2, 3,
			pifra.Label{
				Symbol: pifra.Symbol{pifra.SymbolTypFreshInput, 0}}},
		{1, 4,
			pifra.Label{
				Symbol: pifra.Symbol{pifra.SymbolTypTau, 0}}},
		{4, 3,
			pifra.Label{
				Symbol: pifra.Symbol{pifra.SymbolTypTau, 0}}},
	}

	var lts pifra.Lts = pifra.Lts{
		States:      states,
		Transitions: transitions,
	}

	M, dict, _ := f(lts)
	var expectedDict map[int]int = map[int]int{
		0: 1,
		1: 2,
		2: 3,
		3: 4,
	}
	if fmt.Sprint(dict) != fmt.Sprint(expectedDict) {
		t.Errorf("Dictionary produced is not as expected. Expected: %s, got: %s.",
			fmt.Sprint(expectedDict), fmt.Sprint(dict))
	}
	var expectedM [][]bool = [][]bool{
		{true, false, true, true},
		{false, true, false, false},
		{false, false, true, false},
		{false, false, true, true},
	}
	if fmt.Sprint(M) != fmt.Sprint(expectedM) {
		t.Errorf("Result produced is not as expected. Expected: %s, got: %s.",
			fmt.Sprint(expectedM), fmt.Sprint(M))
	}
}

func TestDoWeakTransform(t *testing.T) {
	var states = map[int]pifra.Configuration{
		1: {},
		2: {},
		3: {},
		4: {},
		5: {},
	}
	var transitions []pifra.Transition = []pifra.Transition{
		{1, 2,
			pifra.Label{
				Symbol:  pifra.Symbol{pifra.SymbolTypTau, 0},
				Symbol2: pifra.Symbol{pifra.SymbolTypTau, 0}}},
		{2, 3,
			pifra.Label{
				Symbol:  pifra.Symbol{pifra.SymbolTypTau, 0},
				Symbol2: pifra.Symbol{pifra.SymbolTypTau, 0}}},
		{3, 4,
			pifra.Label{
				Symbol: pifra.Symbol{pifra.SymbolTypFreshInput, 1}}},
		{4, 5,
			pifra.Label{
				Symbol:  pifra.Symbol{pifra.SymbolTypTau, 0},
				Symbol2: pifra.Symbol{pifra.SymbolTypTau, 0}}},
	}

	var lts pifra.Lts = pifra.Lts{
		States:      states,
		Transitions: transitions,
	}

	res := doWeakTransform(lts)

	if fmt.Sprint(lts.States) != fmt.Sprint(res.States) {
		t.Errorf("States produced are not as expected. Expected: %s, got: %s.\n",
			fmt.Sprint(lts.States), fmt.Sprint(res.States))
	}

	// Transition to itself
	var expectedNewTransitions []pifra.Transition = append(transitions, []pifra.Transition{
		{1, 1,
			pifra.Label{
				Symbol:  pifra.Symbol{pifra.SymbolTypTau, 0},
				Symbol2: pifra.Symbol{pifra.SymbolTypTau, 0}}},
		{2, 2,
			pifra.Label{
				Symbol:  pifra.Symbol{pifra.SymbolTypTau, 0},
				Symbol2: pifra.Symbol{pifra.SymbolTypTau, 0}}},
		{3, 3,
			pifra.Label{
				Symbol:  pifra.Symbol{pifra.SymbolTypTau, 0},
				Symbol2: pifra.Symbol{pifra.SymbolTypTau, 0}}},
		{4, 4,
			pifra.Label{
				Symbol:  pifra.Symbol{pifra.SymbolTypTau, 0},
				Symbol2: pifra.Symbol{pifra.SymbolTypTau, 0}}},
		{5, 5,
			pifra.Label{
				Symbol:  pifra.Symbol{pifra.SymbolTypTau, 0},
				Symbol2: pifra.Symbol{pifra.SymbolTypTau, 0}}},
	}...)

	// New transitive tau transitions
	expectedNewTransitions = append(expectedNewTransitions, []pifra.Transition{
		{1, 3,
			pifra.Label{
				Symbol:  pifra.Symbol{pifra.SymbolTypTau, 0},
				Symbol2: pifra.Symbol{pifra.SymbolTypTau, 0}}},
	}...)

	// New transitive a-transitions
	expectedNewTransitions = append(expectedNewTransitions, []pifra.Transition{
		{1, 4,
			pifra.Label{
				Symbol: pifra.Symbol{pifra.SymbolTypFreshInput, 1}}},
		{1, 5,
			pifra.Label{
				Symbol: pifra.Symbol{pifra.SymbolTypFreshInput, 1}}},
		{2, 4,
			pifra.Label{
				Symbol: pifra.Symbol{pifra.SymbolTypFreshInput, 1}}},
		{2, 5,
			pifra.Label{
				Symbol: pifra.Symbol{pifra.SymbolTypFreshInput, 1}}},
		{3, 5,
			pifra.Label{
				Symbol: pifra.Symbol{pifra.SymbolTypFreshInput, 1}}},
	}...)

	var expectedNewTransitionsSet = make(map[string]bool)
	for tt := range expectedNewTransitions {
		expectedNewTransitionsSet[fmt.Sprint(expectedNewTransitions[tt])] = true
	}
	var resTransitionsSet = make(map[string]bool)
	for tt := range res.Transitions {
		resTransitionsSet[fmt.Sprint(res.Transitions[tt])] = true
	}
	// var fail = false
	// for tt := range res.Transitions {
	// 	key := fmt.Sprint(res.Transitions[tt])
	// 	if _, ok := expectedNewTransitionsSet[key]; !ok {
	// 		fail = true
	// 		t.Logf("Got transition that was not expected. Expected: %s, got: %s.\n",
	// 			"None", key)
	// 	}
	// }
	// for tt := range res.Transitions {
	// 	key := fmt.Sprint(res.Transitions[tt])
	// 	if _, ok := expectedNewTransitionsSet[key]; !ok {
	// 		fail = true
	// 		t.Logf("Got transition that was not expected. Expected: %s, got: %s.\n",
	// 			"None", key)
	// 	}
	// }
	if fmt.Sprint(expectedNewTransitionsSet) != fmt.Sprint(resTransitionsSet) {
		t.Errorf("Transitions produced is not as expected. Expected: %s, got: %s.\n",
			fmt.Sprint(expectedNewTransitionsSet), fmt.Sprint(resTransitionsSet))
	}
}
