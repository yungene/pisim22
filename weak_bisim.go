package main

import (
	"fmt"
	"sort"

	"github.com/yungene/pifra"
)

const (
	SymbolEps pifra.SymbolType = 1592
)

// Transform an LTS with tau transitions into one suitable for weak bisimulation.
func doWeakTransform(lts pifra.Lts) pifra.Lts {
	// Transformation is done in two steps. First, =t=> transitions are calculated
	// by considering a transitive closure of tau transitions. Then these new
	// transitions are used to generate all observable weak transitions.
	// The idea is due to Cleaveland and Sokolsky @ 2001.
	var states = make(map[int]pifra.Configuration)
	var transitions []pifra.Transition = make([]pifra.Transition, 0, 2*len(lts.Transitions))

	// First, do a transitive closure of tau transitions.
	// Use a Floyd-Warshall algorithm. We just want to know the reachability via
	// tau transitions.
	// |M| is V^2, call to floydWarshall is O(V^3)
	var M [][]bool
	var dict map[int]int
	var revDict map[int]int
	if getClosureAlgortihmChoice() == 2 {
		M, dict, revDict = floydWarshall(lts)
	} else {
		// call to dfsClosure is O(V^2 + V*E), might be slightly better than floydWarshall
		// for sparse graphs.
		M, dict, revDict = dfsClosure(lts)
	}
	// if isDebug() {
	// 	fmt.Println(M)
	// }

	// Second create a new LTS without the tau transitions.
	// I.e. create new transitions. E.g. if P1-a->P2 and M[P2][P6] == true, then
	// we can add transition P1-a->P6.
	// We replace tau transitions with the reachability =eps=> transitions. Or
	// rather we add new tau transitions to make the existing bisimulation
	// algorithm to work.
	adj := ToAdjacency(lts)
	states = lts.States
	var statesArr []int
	for state := range lts.States {
		statesArr = append(statesArr, state)
	}
	sort.Ints(statesArr)
	var visited = make(map[string]bool, len(lts.Transitions))
	// O(V + E*V^2)
	for i := range statesArr { // O(V)
		state := statesArr[i]

		// second add all -a-> transitions and all =t=> X -a-> and -a-> X =t=>.

		// Add a self-transition.
		id := revDict[state]
		if M[id][id] {
			newTrans := pifra.Transition{
				Source:      state,
				Destination: state,
				Label: pifra.Label{
					Symbol:  pifra.Symbol{Type: pifra.SymbolTypTau, Value: 0},
					Symbol2: pifra.Symbol{Type: pifra.SymbolTypTau, Value: 0},
				},
			}
			key := fmt.Sprint(newTrans)
			if val, ok := visited[key]; !ok && !val {
				transitions = append(transitions, newTrans)
				visited[key] = true
			}
		}
		for t := range adj[state] {
			trans := adj[state][t]
			key := fmt.Sprint(trans)
			if val, ok := visited[key]; !ok && !val {
				transitions = append(transitions, trans)
				visited[key] = true
			}
			//transitions = append(transitions, trans)
			src := trans.Source
			dest := trans.Destination
			srcId := revDict[src]
			destId := revDict[dest]

			// prefix case
			// e.g. P1 =t=> P2 -a-> P3 results in P1 -a-> P3
			// O(V^2), in reality depends on sparsity of tau edges
			for jj := 0; jj < len(M); jj++ {
				if M[jj][srcId] {
					for ii := 0; ii < len(M[destId]); ii++ {
						if M[destId][ii] {
							newTrans := pifra.Transition{
								Source:      dict[jj],
								Destination: dict[ii],
								Label:       trans.Label,
							}
							key := fmt.Sprint(newTrans)
							if val, ok := visited[key]; !ok && !val {
								transitions = append(transitions, newTrans)
								visited[key] = true
							}
						}
					}
					// newTrans := pifra.Transition{
					// 	Source:      dict[jj],
					// 	Destination: trans.Destination,
					// 	Label:       trans.Label,
					// }
					// key := fmt.Sprint(newTrans)
					// if val, ok := visited[key]; !ok && !val {
					// 	transitions = append(transitions, newTrans)
					// 	visited[key] = true
					// }
				}
			}
			// postfix case
			// O(V)
			// for ii := 0; ii < len(M[destId]); ii++ {
			// 	if M[destId][ii] && ii != destId {
			// 		newTrans := pifra.Transition{
			// 			Source:      src,
			// 			Destination: dict[ii],
			// 			Label:       trans.Label,
			// 		}
			// 		key := fmt.Sprint(newTrans)
			// 		if val, ok := visited[key]; !ok && !val {
			// 			transitions = append(transitions, newTrans)
			// 			visited[key] = true
			// 		}
			// 	}
			// }
		}
	}

	return pifra.Lts{
		States:       states,
		Transitions:  transitions,
		FreeNamesMap: lts.FreeNamesMap,
	}
}

// Cubic in |V|. Linear in |E|.
func floydWarshall(lts pifra.Lts) ([][]bool, map[int]int, map[int]int) {
	var states []int
	for state := range lts.States {
		states = append(states, state)
	}
	sort.Ints(states)

	var M = make([][]bool, len(states))
	for i := range M {
		M[i] = make([]bool, len(states))
	}
	var dict = make(map[int]int)
	var revDict = make(map[int]int)
	// O(V)
	for i := range states {
		dict[i] = states[i]
		revDict[states[i]] = i
	}

	// Reachable base case.
	// O(V^2)
	for i := range M {
		for j := range M[i] {
			if i == j {
				M[i][j] = true
			} else {
				M[i][j] = false
			}
		}
	}

	// Reachable via any directed tau edge.
	// O(E)
	for t := range lts.Transitions {
		trans := lts.Transitions[t]
		i := trans.Source
		j := trans.Destination
		if trans.Label.Symbol.Type == pifra.SymbolTypTau {
			M[revDict[i]][revDict[j]] = true
		}
	}

	V := len(states)
	// O(V^3)
	for k := 0; k < V; k++ {
		for i := 0; i < V; i++ {
			for j := 0; j < V; j++ {
				M[i][j] = M[i][j] || (M[i][k] && M[k][j])
			}
		}
	}
	return M, dict, revDict
}

func dfsClosure(lts pifra.Lts) ([][]bool, map[int]int, map[int]int) {
	var states []int
	for state := range lts.States {
		states = append(states, state)
	}
	sort.Ints(states)
	var M = make([][]bool, len(states))
	for i := range M {
		M[i] = make([]bool, len(states))
	}
	var dict = make(map[int]int, len(states))
	var revDict = make(map[int]int, len(states))
	// O(V)
	for i := range states {
		dict[i] = states[i]
		revDict[states[i]] = i
	}

	// // Reachable base case.
	// // O(V^2)
	// for i := range M {
	// 	for j := range M[i] {
	// 		if i == j {
	// 			M[i][j] = true
	// 		} else {
	// 			M[i][j] = false
	// 		}
	// 	}
	// }

	// do DFS from each vertex
	V := len(states)
	adj := ToAdjacency(lts)
	for i := 0; i < V; i++ {
		dfsUtil(adj, dict, revDict, M, i, i)
	}

	return M, dict, revDict
}

func dfsUtil(adj map[int][]pifra.Transition, dict map[int]int, revDict map[int]int,
	M [][]bool, u int, v int) {
	if u == v {
		// Whether to add reachability to itself? Yes, even without an edge.
		M[u][v] = true
	} else {
		M[u][v] = true
	}
	//uOrig := dict[u]
	vOrig := dict[v]

	for t := range adj[vOrig] {
		trans := adj[vOrig][t]
		j := trans.Destination
		if trans.Label.Symbol.Type == pifra.SymbolTypTau {
			w := revDict[j]
			if !M[u][w] {
				dfsUtil(adj, dict, revDict, M, u, w)
			}
		}
	}

}
