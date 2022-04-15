package main

import (
	"fmt"
	"sync"

	"github.com/yungene/pifra"
)

// #############################################################################
// ############################## GLOBALS ######################################
// #############################################################################
var stackDepth = 0
var notR sync.Map

func resetBisim() {
	// Reset the global variables, necessary for testing, etc.
	stackDepth = 0
	notR = sync.Map{}
	IC.resetBisim()
}

// #############################################################################
// ############################ SINGLE THREAD ##################################
// #############################################################################
func checkBisim(leftLts pifra.Lts, rightLts pifra.Lts,
	weakLeftLts pifra.Lts, weakRightLts pifra.Lts,
	regSizeOverride int, debugSpecificFlag int, findAllFlag bool) ResultType {
	resetBisim()
	var res = ResultNotRelated
	//counter := 0
	if regSizeOverride > 0 {
		setRegSize(regSizeOverride)
	} else {
		n1 := getMaxMinRegSize(leftLts)
		n2 := getMaxMinRegSize(rightLts)
		setRegSize(maxInt(n1, n2))
	}

	// FREE_NAMES: generate the required maximum mapping here
	var initRho = make(map[int]int)
	var startOrigConfLeft = leftLts.States[0]
	var startOrigConfRight = rightLts.States[0]

	// We have lts.FreeNamesMap which maps new names to original names
	// we want inverse registers, to get the index by name, then for each free name,
	// we get its new name on left and new name on right, we
	var invRegLeft map[string]int
	var invRegRight map[string]int
	invRegLeft_, err_ := reverseMapIntString(startOrigConfLeft.Registers.Registers)
	if err_ != nil {
		fmt.Printf("%s\n", err_.Error())
		return ResultNotRelated
	} else {
		invRegLeft = invRegLeft_
	}

	invRegRight_, err_ := reverseMapIntString(startOrigConfRight.Registers.Registers)
	if err_ != nil {
		fmt.Printf("%s\n", err_.Error())
		return ResultNotRelated
	} else {
		invRegRight = invRegRight_
	}

	var allFreeNames = make(map[string]bool)
	for _, origName := range leftLts.FreeNamesMap {
		allFreeNames[origName] = true
	}
	for _, origName := range rightLts.FreeNamesMap {
		allFreeNames[origName] = true
	}

	var invFreeNamesLeft map[string]string
	invFreeNamesLeft_, err_ := reverseMapStringString(leftLts.FreeNamesMap)
	if err_ != nil {
		fmt.Printf("%s\n", err_.Error())
		return ResultNotRelated
	} else {
		invFreeNamesLeft = invFreeNamesLeft_
	}
	var invFreeNamesRight map[string]string
	invFreeNamesRight_, err_ := reverseMapStringString(rightLts.FreeNamesMap)
	if err_ != nil {
		fmt.Printf("%s\n", err_.Error())
		return ResultNotRelated
	} else {
		invFreeNamesRight = invFreeNamesRight_
	}

	if isVerbose() {
		fmt.Printf("Registers left: %s.\n", pifra.PrettyPrintRegister(startOrigConfLeft.Registers))
		fmt.Printf("Left free names map: %s.\n", leftLts.FreeNamesMap)
		fmt.Printf("Registers right: %s.\n", pifra.PrettyPrintRegister(startOrigConfRight.Registers))
		fmt.Printf("Right free names map: %s.\n", rightLts.FreeNamesMap)
	}
	for freeName := range allFreeNames {
		// find the new name in left LTS
		if newNameLeft, ok := invFreeNamesLeft[freeName]; ok {
			if newNameRight, ok := invFreeNamesRight[freeName]; ok {
				if regIdxLeft, ok := invRegLeft[newNameLeft]; ok {
					if redIdxRight, ok := invRegRight[newNameRight]; ok {
						initRho[regIdxLeft] = redIdxRight
					}
				}
			}
		}
	}

	res, err_ = cleavelandBisim(leftLts, rightLts, weakLeftLts, weakRightLts, initRho)
	if err_ != nil {
		fmt.Printf("%s\n", err_.Error())
		return ResultNotRelated
	}
	//fmt.Printf("\nTotal number of callback calls performed for these inputs for N=%d is %d.\n", getRegSize(), counter)
	if isVerbose() {
		fmt.Printf("N was chosen to be %d.\n", getRegSize())
	}
	return res
}

// Bisimulation algorithm as per Cleaveland & Sokolsky 2001 paper.
// The algorithm is due to Celikkan.
//
// The system 1 is referred to as "left" and the system 2 is referred to as "right".
func cleavelandBisim(leftLts pifra.Lts, rightLts pifra.Lts,
	weakLeftLts pifra.Lts, weakRightLts pifra.Lts,
	initPerm map[int]int) (res ResultType, err error) {
	// SECTION 1: Create the global state.

	var state *CleavelandState = NewCleavelandState(leftLts, rightLts, weakLeftLts, weakRightLts)

	// SECTION 2: Create the starting states, we assume they are both at index 0.
	startRhoLeft := make(map[int]int)

	startOrigConfLeft := leftLts.States[0]
	startOrigConfRight := rightLts.States[0]

	for k := range initPerm {
		startRhoLeft[k] = initPerm[k]
	}
	startStateLeft := FRAConfiguration{
		Process:   startOrigConfLeft.Process,
		Registers: startOrigConfLeft.Registers,
		Label:     startOrigConfLeft.Label,
		Rho:       startRhoLeft,
		N:         getRegSize(),
	}
	var startRhoRight map[int]int
	revRho, err_ := reverseMap(startRhoLeft)
	if err_ != nil {
		err = err_
		return
	} else {
		startRhoRight = revRho
	}
	startStateRight := FRAConfiguration{
		Process:   startOrigConfRight.Process,
		Registers: startOrigConfRight.Registers,
		Label:     startOrigConfRight.Label,
		Rho:       startRhoRight,
		N:         getRegSize(),
	}

	if isDebug() {
		fmt.Printf("Start states: %s, %s\n", startStateLeft, startStateRight)
	}
	state.addNPState(startStateLeft, 0)
	state.addNQState(startStateRight, 0)

	res = preorder(state, startStateLeft, startStateRight)
	if res == ResultRelated {
		fmt.Printf("\n*** Systems are BISIMILAR for rho %s, N=%d.\n\n", fmt.Sprint(initPerm), getRegSize())
		if isDebug() {
			fmt.Printf("Graph is %s.\n", fmt.Sprint(state.G))
		}
	} else {
		fmt.Printf("\n^^^ Systems are NOT bisimilar for rho %s, N=%d.\n\n", fmt.Sprint(initPerm), getRegSize())
	}

	if isOutputGraph() {
		fmt.Printf("Graph is %s.\n", fmt.Sprint(state.G))
	}

	if isVerbose() {
		fmt.Printf("Bisimulation graph has %d states and %d transitions.\n", len(state.G.States), len(state.G.TransitionsSet))
	}

	if fn := getOutputBisimLtsName(); fn != "" {
		m := getBisimilarStates(state)
		fmt.Print(bisimilarStatesToString(m))
		data := generateBisimGraphVizFile(leftLts, rightLts, m)
		check(writeFile(fn+".bisim"+".dot", data))
		dataTex := generateBisimGraphVizTexFile(leftLts, rightLts, m)
		check(writeFile(fn+".bisim.tex"+".dot", dataTex))
	}

	return
}

func preorderGeneric(state *CleavelandState, nP FRAConfiguration, pId int,
	nQ FRAConfiguration, qId int, isLeft bool) ResultType {
	state.addNState(nP, pId, isLeft)
	state.addNState(nQ, qId, !isLeft)
	if isLeft {
		return preorder(state, nP, nQ)
	} else {
		return preorder(state, nQ, nP)
	}
}

// #############################################################################
// ############################### PREORDER ####################################
// #############################################################################

// This corresponds to BISIM() in the report.
func preorder(state *CleavelandState, nP FRAConfiguration, nQ FRAConfiguration) ResultType {
	pairKey := getFRAPairKey(nP, nQ)
	IC.enterToPreorder++
	if isDebug() {
		fmt.Printf("%d. preorder 0: %s.\n", stackDepth, pairKey)
	}
	if _, ok := notR.Load(pairKey); ok {
		return ResultNotRelated
	}

	var vertex gVertex = gVertex{nP, nQ}
	var vertexKey string = gVertexToString(&vertex)
	if _, ok := state.G.States[vertexKey]; ok {
		return ResultRelated
	}
	IC.preorderStackDepth++
	IC.maxPreorderStackDepth = maxInt(IC.maxPreorderStackDepth, IC.preorderStackDepth)
	state.G.States[vertexKey] = vertex

	status := ResultRelated
	if isDebug() {
		fmt.Printf("%d. PREORDER: %s.\n", stackDepth, pairKey)
	}
	// Match each a-derivative of p with some a-derivative of q.
	// Here generate all a transitions from nP given nQ.
	nPId, ok := state.NStateToId[getFRAConfigurationKey(nP, true)]
	if !ok {
		status = ResultNotRelated
	}
	pId := state.RevMap[nPId]

	var A []AKey
	for lk := range state.AdjLeft[pId] {
		for i := range state.AdjLeft[pId][lk] {
			if status == ResultNotRelated {
				break
			}
			status = ResultNotRelated
			if isDebug() {
				fmt.Printf("%d. Checking next % s P transitions %s.\n",
					stackDepth,
					pairKey, fmt.Sprint(state.AdjLeft[pId][lk][i]))
			}
			// We now build a new transition and a new destination nPDest
			status = processDerivativesGeneric(state, nP, nQ, i, lk,
				state.AdjLeft, state.AdjRight, state.WeakAdjLeft, state.WeakAdjRight,
				state.High, state.HighTwo, &state.LeftLts,
				&state.RightLts, true, GLabelOne)

			if status == ResultNotRelated {
				if isDebug() {
					fmt.Printf("%d. Was not able to find a match for state % s P transitions %s.\n",
						stackDepth,
						pairKey, fmt.Sprint(state.AdjLeft[pId][lk][i]))
					fmt.Println(state.G)
				}
				A = state.populateA(A, vertexKey)
			}
		}
	}
	qPId, ok := state.NStateToId[getFRAConfigurationKey(nQ, false)]
	if !ok {
		status = ResultNotRelated
	}
	qId := state.RevMap[qPId]
	for lk := range state.AdjRight[qId] {
		for i := range state.AdjRight[qId][lk] {
			if status == ResultNotRelated {
				break
			}
			status = ResultNotRelated
			if isDebug() {
				fmt.Printf("%d. Checking next % s Q transitions %s.\n",
					stackDepth,
					pairKey, fmt.Sprint(state.AdjRight[qId][lk][i]))
			}
			// We now build a new transition and a new destination nPDest
			status = processDerivativesGeneric(state, nQ, nP, i, lk,
				state.AdjRight, state.AdjLeft, state.WeakAdjRight, state.WeakAdjLeft,
				state.Low, state.LowTwo, &state.RightLts,
				&state.LeftLts, false, GLabelTwo)

			if status == ResultNotRelated {
				if isDebug() {
					fmt.Printf("%d. Was not able to find a match for state % s Q transitions %s.\n",
						stackDepth,
						pairKey, fmt.Sprint(state.AdjRight[qId][lk][i]))
					fmt.Println(state.G)
				}
				A = state.populateA(A, vertexKey)
			}
		}
	}

	if status == ResultNotRelated {
		if isDebug() {
			fmt.Println("Starting processing A.")
		}
		// remove the vertex
		delete(state.G.States, vertexKey)
		// remove both incoming and outgoing edges
		state.removeIncidentEdges(vertexKey)
		notR.Store(vertexKey, true)
		if isDebug() {
			fmt.Printf("%d. Added to not R %s.\n",
				stackDepth,
				vertexKey)
		}
		i := 0
		for ; true; i++ {
			if i >= len(A) {
				break
			}
			IC.reevalA++
			key := A[i]
			if isDebug() {
				fmt.Printf("Popped %s.\n", fmt.Sprint(key))
			}
			nRId, ok := state.NStateToId[key.NP]
			if !ok {
				continue
			}
			nR := state.States[nRId]
			rId := state.RevMap[nRId]
			nSId, ok := state.NStateToId[key.NQ]
			if !ok {
				continue
			}
			nS := state.States[nSId]
			sId := state.RevMap[nSId]
			if key.Type == GLabelOne {
				var act string = fmt.Sprint(state.AdjLeft[rId][key.LabelsKey][key.TransId])
				if key.KPrime == noKPrime {
					var hlKey HLKey = HLKey{
						Dest: getFRAConfigurationKey(nR, true),
						Src:  getFRAConfigurationKey(nS, false),
						Act:  act,
					}
					state.High[hlKey] += 1
				} else {
					var hlPrimeKey = HLKeyFINP{
						Dest:   getFRAConfigurationKey(nR, true),
						Src:    getFRAConfigurationKey(nS, false),
						Act:    act,
						KPrime: key.KPrime,
					}
					state.HighTwo[hlPrimeKey] += 1
				}
				status = processDerivativesGeneric(state, nR, nS, key.TransId, key.LabelsKey,
					state.AdjLeft, state.AdjRight, state.WeakAdjLeft, state.WeakAdjRight,
					state.High, state.HighTwo,
					&state.LeftLts, &state.RightLts, true, GLabelOne)
				if status == ResultNotRelated {
					rsKey := gVertexToString(&gVertex{nR, nS})
					A = state.populateA(A, rsKey)

					delete(state.G.States, rsKey)
					// remove both incoming and outgoing edges
					state.removeIncidentEdges(rsKey)
					notR.Store(rsKey, true)
					if isDebug() {
						fmt.Printf("%d. Added to not R %s.\n",
							stackDepth,
							rsKey)
					}
				}
			} else if key.Type == GLabelTwo {
				var act string = fmt.Sprint(state.AdjRight[sId][key.LabelsKey][key.TransId])

				if key.KPrime == noKPrime {
					var hlKey HLKey = HLKey{
						Dest: getFRAConfigurationKey(nS, false),
						Src:  getFRAConfigurationKey(nR, true),
						Act:  act,
					}
					state.Low[hlKey] += 1
				} else {
					var hlPrimeKey = HLKeyFINP{
						Dest:   getFRAConfigurationKey(nS, false),
						Src:    getFRAConfigurationKey(nR, true),
						Act:    act,
						KPrime: key.KPrime,
					}
					state.LowTwo[hlPrimeKey] += 1
				}

				status = processDerivativesGeneric(state, nS, nR, key.TransId, key.LabelsKey,
					state.AdjRight, state.AdjLeft, state.WeakAdjRight, state.WeakAdjLeft,
					state.Low, state.LowTwo, &state.RightLts,
					&state.LeftLts, false, GLabelTwo)
				if status == ResultNotRelated {
					rsKey := gVertexToString(&gVertex{nR, nS})
					A = state.populateA(A, rsKey)

					delete(state.G.States, rsKey)
					// remove both incoming and outgoing edges
					state.removeIncidentEdges(rsKey)
					notR.Store(rsKey, true)
					if isDebug() {
						fmt.Printf("%d. Added to not R %s.\n",
							stackDepth,
							rsKey)
					}
				}
			}

		}
	}

	if isDebug() {
		fmt.Printf("%d. Return from preorder with key: %s, status: %d.\n",
			stackDepth,
			vertexKey, status)
	}
	IC.fullExecutePreorder++
	IC.preorderStackDepth--
	return status
}

// #############################################################################
// ########################## PROCESS_DERIVATIVES ##############################
// #############################################################################

// This corresponds to MATCH_LEFT() and MATCH_RIGHT() in the report.
func processDerivativesGeneric(state *CleavelandState, nP FRAConfiguration,
	nQ FRAConfiguration, transId int, labelsKey LabelsKey,
	adjLeft AdvAdj,
	adjRight AdvAdj,
	weakAdjLeft AdvAdj,
	weakAdjRight AdvAdj,
	high map[HLKey]int,
	highTwo map[HLKeyFINP]int,
	leftLts *pifra.Lts,
	rightLts *pifra.Lts,
	isLeft bool,
	edgeLabel gLabel) ResultType {

	IC.enterProcessDerivatives++

	// var derivatives []FRAConfiguration
	status := ResultNotRelated
	nPId, ok := state.NStateToId[getFRAConfigurationKey(nP, isLeft)]
	if !ok {
		return ResultNotRelated
	}
	pId := state.RevMap[nPId]
	nQId, ok := state.NStateToId[getFRAConfigurationKey(nQ, !isLeft)]
	if !ok {
		return ResultNotRelated
	}
	qId := state.RevMap[nQId]
	trans := adjLeft[pId][labelsKey][transId]
	pXId := trans.Destination
	pX := leftLts.States[pXId]
	newLabel := trans.Label
	newRho := nP.Rho
	stackDepth++

	var hlKey HLKey = HLKey{
		Dest: getFRAConfigurationKey(nP, isLeft),
		Src:  getFRAConfigurationKey(nQ, !isLeft),
		Act:  fmt.Sprint(trans),
	}
	if _, ok := high[hlKey]; !ok {
		high[hlKey] = 0
	}
	if isDebug() {
		fmt.Printf("%d. Enter processDerivativeGeneric with %s, %s, %s, trans:%s. hlKey is %s.\n",
			stackDepth,
			nP.String(), nQ.String(), fmt.Sprint(isLeft), fmt.Sprint(trans),
			fmt.Sprint(hlKey))
	}

	if trans.Label.Symbol.Type == pifra.SymbolTypTau {
		IC.tauRule++
		// NT rule 1, TAU
		var nPX, nQX FRAConfiguration
		nPX = FRAConfiguration{
			Process:   pX.Process,
			Registers: pX.Registers,
			Label:     newLabel,
			Rho:       newRho,
			N:         getRegSize(),
		}
		nLk := LabelsKey{pifra.SymbolTypTau, pifra.SymbolTypTau}
		for idx := high[hlKey]; idx < len(weakAdjRight[qId][nLk]) && status == ResultNotRelated; idx++ {
			trans2 := weakAdjRight[qId][nLk][idx]
			// TODO: this check should be redundant in theory.
			if trans2.Label.Symbol.Type == pifra.SymbolTypTau {
				qXId := trans2.Destination
				qX := rightLts.States[qXId]
				revRho, err := reverseMap(newRho)
				if err != nil {
					// TODO: increment the high[hlKey] here as well
					high[hlKey] += 1
					continue
				}
				nQX = FRAConfiguration{
					Process:   qX.Process,
					Registers: qX.Registers,
					Rho:       revRho,
					N:         getRegSize(),
				}

				if enableGarbageCollection() {
					// Reset the value of Rho in nPX that might have been changed by previous iteration.
					nPX = FRAConfiguration{
						Process:   pX.Process,
						Registers: pX.Registers,
						Label:     newLabel,
						Rho:       newRho,
						N:         getRegSize(),
					}
					err := fixGC(&nPX, &nQX)
					if err != nil {
						high[hlKey] += 1
						continue
					}
				}

				status = preorderGeneric(state, nPX, pXId, nQX, qXId, isLeft)
				if status == ResultRelated {
					state.createAndAddTransition(&nP, &nQ, &nPX, &nQX, isLeft, edgeLabel, transId,
						noKPrime, labelsKey)
				} else {
					high[hlKey] += 1
				}
			}
		}
		if status == ResultNotRelated {
			if isDebug() {
				fmt.Printf("Not related due to rule 1.\n")
			}
		}
	} else if trans.Label.Symbol.Type == pifra.SymbolTypInput &&
		trans.Label.Symbol2.Type == pifra.SymbolTypKnown {
		// NT rules 2 and 3, INP1 and INP2
		i := trans.Label.Symbol.Value
		pi := nP.Rho[i]
		j := trans.Label.Symbol2.Value
		var nPX, nQX FRAConfiguration
		// check if j is in domain of rho
		if _, ok := nP.Rho[j]; ok {
			IC.inp1Rule++
			pj := nP.Rho[j]
			// if in domain then rule 2, INP1
			if isDebug() {
				fmt.Printf("For nP=%s. The original transitions is %d%d, the translation is %d%d.\n", fmt.Sprint(nP), i, j, pi, pj)
			}
			newLabel = pifra.Label{
				Symbol:  pifra.Symbol{Type: pifra.SymbolTypInput, Value: pi},
				Symbol2: pifra.Symbol{Type: pifra.SymbolTypKnown, Value: pj},
			}
			nPX = FRAConfiguration{
				Process:   pX.Process,
				Registers: pX.Registers,
				Label:     newLabel,
				Rho:       newRho,
				N:         getRegSize(),
			}
			// Find a matching nQX, this is what SEARCH_HIGH should do
			nLk := LabelsKey{pifra.SymbolTypInput, pifra.SymbolTypKnown}
			for idx := high[hlKey]; idx < len(weakAdjRight[qId][nLk]) && status == ResultNotRelated; idx++ {
				trans2 := weakAdjRight[qId][nLk][idx]
				if isDebug() {
					fmt.Println(trans2)
				}
				if trans2.Label.Symbol.Type == pifra.SymbolTypInput &&
					trans2.Label.Symbol2.Type == pifra.SymbolTypKnown &&
					trans2.Label.Symbol.Value == pi &&
					trans2.Label.Symbol2.Value == pj {
					qXId := trans2.Destination
					qX := rightLts.States[qXId]
					revRho, err := reverseMap(newRho)
					if err != nil {
						high[hlKey] += 1
						continue
					}
					nQX = FRAConfiguration{
						Process:   qX.Process,
						Registers: qX.Registers,
						Rho:       revRho,
						N:         getRegSize(),
					}

					if enableGarbageCollection() {
						nPX = FRAConfiguration{
							Process:   pX.Process,
							Registers: pX.Registers,
							Label:     newLabel,
							Rho:       newRho,
							N:         getRegSize(),
						}
						err := fixGC(&nPX, &nQX)
						if err != nil {
							high[hlKey] += 1
							continue
						}
					}

					status = preorderGeneric(state, nPX, pXId, nQX, qXId, isLeft)
					if status == ResultRelated {
						state.createAndAddTransition(&nP, &nQ, &nPX, &nQX, isLeft, edgeLabel, transId,
							noKPrime, labelsKey)
						if isDebug() {
							fmt.Printf("Related due to rule 2 for %s, %s, %s.\n",
								fmt.Sprint(nPX), fmt.Sprint(nQX), fmt.Sprint(isLeft))
						}
					} else {
						high[hlKey] += 1
					}
				}
			}
			if status == ResultNotRelated {
				if isDebug() {
					fmt.Printf("Not related due to rule 2.\n")
				}
			}
		} else {
			IC.inp2Rule++
			// else rule 3, INP2
			nLk := LabelsKey{pifra.SymbolTypInput, pifra.SymbolTypFreshInput}
			for idx := high[hlKey]; idx < len(weakAdjRight[qId][nLk]) && status == ResultNotRelated; idx++ {
				trans2 := weakAdjRight[qId][nLk][idx]
				if trans2.Label.Symbol.Type == pifra.SymbolTypInput &&
					trans2.Label.Symbol2.Type == pifra.SymbolTypFreshInput &&
					trans2.Label.Symbol.Value == pi {
					k := trans2.Label.Symbol2.Value
					qXId := trans2.Destination
					qX := rightLts.States[qXId]
					newRho = makeNewRho(nP.Rho, trans.Label.Symbol2.Value, k)
					if isDebug() {
						fmt.Println(newRho)
					}
					newLabel = pifra.Label{
						Symbol:  pifra.Symbol{Type: pifra.SymbolTypInput, Value: pi},
						Symbol2: pifra.Symbol{Type: pifra.SymbolTypKnown, Value: k},
					}
					nPX = FRAConfiguration{
						Process:   pX.Process,
						Registers: pX.Registers,
						Label:     newLabel,
						Rho:       newRho,
						N:         getRegSize(),
					}
					revRho, err := reverseMap(newRho)
					if err != nil {
						high[hlKey] += 1
						continue
					}
					nQX = FRAConfiguration{
						Process:   qX.Process,
						Registers: qX.Registers,
						Rho:       revRho,
						N:         getRegSize(),
					}

					if enableGarbageCollection() {
						err := fixGC(&nPX, &nQX)
						if err != nil {
							high[hlKey] += 1
							continue
						}
					}

					status = preorderGeneric(state, nPX, pXId, nQX, qXId, isLeft)
					if status == ResultRelated {
						state.createAndAddTransition(&nP, &nQ, &nPX, &nQX, isLeft, edgeLabel, transId,
							noKPrime, labelsKey)
					} else {
						high[hlKey] += 1
					}
				}
			}
			if status == ResultNotRelated {
				if isDebug() {
					fmt.Printf("Not related due to rule 3.\n")
				}
			}
		}

	} else if trans.Label.Symbol.Type == pifra.SymbolTypOutput &&
		trans.Label.Symbol2.Type == pifra.SymbolTypKnown {
		IC.outRule++
		// NT rule 5, OUT
		i := trans.Label.Symbol.Value
		pi := nP.Rho[i]
		j := trans.Label.Symbol2.Value
		var nPX, nQX FRAConfiguration
		// check if j in in domain of rho
		if _, ok := nP.Rho[j]; ok {
			pj := nP.Rho[j]
			newLabel = pifra.Label{
				Symbol:  pifra.Symbol{Type: pifra.SymbolTypOutput, Value: pi},
				Symbol2: pifra.Symbol{Type: pifra.SymbolTypKnown, Value: pj},
			}
			nPX = FRAConfiguration{
				Process:   pX.Process,
				Registers: pX.Registers,
				Label:     newLabel,
				Rho:       newRho,
				N:         getRegSize(),
			}
			nLk := LabelsKey{pifra.SymbolTypOutput, pifra.SymbolTypKnown}
			for idx := high[hlKey]; idx < len(weakAdjRight[qId][nLk]) && status == ResultNotRelated; idx++ {
				trans2 := weakAdjRight[qId][nLk][idx]
				if trans2.Label.Symbol.Type == pifra.SymbolTypOutput &&
					trans2.Label.Symbol2.Type == pifra.SymbolTypKnown &&
					trans2.Label.Symbol.Value == pi &&
					trans2.Label.Symbol2.Value == pj {
					qXId := trans2.Destination
					qX := rightLts.States[qXId]
					revRho, err := reverseMap(newRho)
					if err != nil {
						high[hlKey] += 1
						continue
					}
					nQX = FRAConfiguration{
						Process:   qX.Process,
						Registers: qX.Registers,
						Rho:       revRho,
						N:         getRegSize(),
					}
					if enableGarbageCollection() {
						nPX = FRAConfiguration{
							Process:   pX.Process,
							Registers: pX.Registers,
							Label:     newLabel,
							Rho:       newRho,
							N:         getRegSize(),
						}
						err := fixGC(&nPX, &nQX)
						if err != nil {
							high[hlKey] += 1
							continue
						}
					}
					status = preorderGeneric(state, nPX, pXId, nQX, qXId, isLeft)
					if status == ResultRelated {
						state.createAndAddTransition(&nP, &nQ, &nPX, &nQX, isLeft, edgeLabel, transId,
							noKPrime, labelsKey)
					} else {
						high[hlKey] += 1
					}
				}
			}
		}
	} else if trans.Label.Symbol.Type == pifra.SymbolTypInput &&
		trans.Label.Symbol2.Type == pifra.SymbolTypFreshInput {
		IC.finpRule++
		// NT rule 4, FINP

		// First half -> find the matching fresh transition, FINP.1
		i := trans.Label.Symbol.Value
		pi := nP.Rho[i]
		var nPX, nQX FRAConfiguration
		var edges []gTransition
		nLk := LabelsKey{pifra.SymbolTypInput, pifra.SymbolTypFreshInput}
		for idx := high[hlKey]; idx < len(weakAdjRight[qId][nLk]) && status == ResultNotRelated; idx++ {
			trans2 := weakAdjRight[qId][nLk][idx]
			if trans2.Label.Symbol.Type == pifra.SymbolTypInput &&
				trans2.Label.Symbol2.Type == pifra.SymbolTypFreshInput &&
				trans2.Label.Symbol.Value == pi {
				if isDebug() {
					fmt.Printf("Rule 4.1 trans2: %s\n", fmt.Sprint(trans2))
				}
				k := trans2.Label.Symbol2.Value
				qXId := trans2.Destination
				qX := rightLts.States[qXId]
				newRho = makeNewRho(nP.Rho, trans.Label.Symbol2.Value, k)
				newLabel = pifra.Label{
					Symbol:  pifra.Symbol{Type: pifra.SymbolTypInput, Value: pi},
					Symbol2: pifra.Symbol{Type: pifra.SymbolTypFreshInput, Value: k},
				}
				nPX = FRAConfiguration{
					Process:   pX.Process,
					Registers: pX.Registers,
					Label:     newLabel,
					Rho:       newRho,
					N:         getRegSize(),
				}
				revRho, err := reverseMap(newRho)
				if err != nil {
					high[hlKey] += 1
					continue
				}
				nQX = FRAConfiguration{
					Process:   qX.Process,
					Registers: qX.Registers,
					Rho:       revRho,
					N:         getRegSize(),
				}
				if enableGarbageCollection() {
					err := fixGC(&nPX, &nQX)
					if err != nil {
						high[hlKey] += 1
						continue
					}
				}
				status = preorderGeneric(state, nPX, pXId, nQX, qXId, isLeft)
				if status == ResultRelated {
					var edge *gTransition = createEdge(&nP, &nQ, &nPX, &nQX, isLeft, edgeLabel, transId, noKPrime, labelsKey)
					edges = append(edges, *edge)
				} else {
					high[hlKey] += 1
				}
			}
		}
		// new version of FINP.2
		if status == ResultRelated {
			var image = getImage(nP.Rho)
			// for each item in range but not in image
			var kPrimes = make(map[int]bool)
			// Create kPrimes - indices of all the registers in nQ that are not empty, but
			// which are not in the image of nP
			for idx := range nQ.Registers.Registers {
				// idx is assumed to always be a non-empty register.
				if _, ok := image[idx]; !ok {
					kPrimes[idx] = false
				}
			}

			// check every kPrime for a matching transition
			if isDebug() {
				fmt.Printf("kPrimes are: %s.\n", fmt.Sprint(kPrimes))
			}
		kPrimesLoop:
			for i := range kPrimes {
				pj := i
				var nPX2, nQX2 FRAConfiguration
				// basically copy INP1 with modifications
				newLabel = pifra.Label{
					Symbol:  pifra.Symbol{Type: pifra.SymbolTypInput, Value: pi},
					Symbol2: pifra.Symbol{Type: pifra.SymbolTypKnown, Value: pj},
				}
				newRho = makeNewRho(nP.Rho, trans.Label.Symbol2.Value, pj)
				nPX2 = FRAConfiguration{
					Process:   pX.Process,
					Registers: pX.Registers,
					Label:     newLabel,
					Rho:       newRho,
					N:         getRegSize(),
				}
				kStatus := ResultNotRelated
				hlPrimeKey := HLKeyFINP{
					Dest:   hlKey.Dest,
					Src:    hlKey.Src,
					Act:    hlKey.Act,
					KPrime: pj,
				}

				if _, ok := highTwo[hlPrimeKey]; !ok {
					highTwo[hlPrimeKey] = 0
				}
				nLk := LabelsKey{pifra.SymbolTypInput, pifra.SymbolTypKnown}
				for idx := highTwo[hlPrimeKey]; idx < len(weakAdjRight[qId][nLk]) && kStatus == ResultNotRelated; idx++ {
					trans2 := weakAdjRight[qId][nLk][idx]
					if isDebug() {
						fmt.Printf("Rule 4.2 trans2 preprocess: %s\n", fmt.Sprint(trans2))
					}
					if trans2.Label.Symbol.Type == pifra.SymbolTypInput &&
						trans2.Label.Symbol2.Type == pifra.SymbolTypKnown &&
						trans2.Label.Symbol.Value == pi &&
						trans2.Label.Symbol2.Value == pj {
						if isDebug() {
							fmt.Printf("Rule 4.2 trans2: %s\n", fmt.Sprint(trans2))
						}
						qXId := trans2.Destination
						qX := rightLts.States[qXId]
						revRho, err := reverseMap(newRho)
						if err != nil {
							highTwo[hlPrimeKey]++
							continue
						}
						nQX2 = FRAConfiguration{
							Process:   qX.Process,
							Registers: qX.Registers,
							Rho:       revRho,
							N:         getRegSize(),
						}
						if enableGarbageCollection() {
							nPX2 = FRAConfiguration{
								Process:   pX.Process,
								Registers: pX.Registers,
								Label:     newLabel,
								Rho:       newRho,
								N:         getRegSize(),
							}
							err := fixGC(&nPX2, &nQX2)
							if err != nil {
								highTwo[hlPrimeKey]++
								continue
							}
						}
						kStatus = preorderGeneric(state, nPX2, pXId, nQX2, qXId, isLeft)
						if status == ResultRelated {
							var edge *gTransition = createEdge(&nP, &nQ, &nPX2, &nQX2, isLeft, edgeLabel, transId, pj, labelsKey)
							edges = append(edges, *edge)
						} else {
							highTwo[hlPrimeKey]++
						}
					}
				}
				if kStatus == ResultNotRelated {
					// the current invocation to processDerivatives failed. Can remove
					status = ResultNotRelated
					break kPrimesLoop
				}
			}
			if status == ResultRelated {
				for ii := range edges {
					state.addTransition(edges[ii])
				}
			}
		}

	} else if trans.Label.Symbol.Type == pifra.SymbolTypOutput &&
		trans.Label.Symbol2.Type == pifra.SymbolTypFreshOutput {
		IC.foutRule++
		// NT rule 6, FOUT
		// find a matching transitions from the pair state
		i := trans.Label.Symbol.Value
		pi := nP.Rho[i]
		var nPX, nQX FRAConfiguration
		nLk := LabelsKey{pifra.SymbolTypOutput, pifra.SymbolTypFreshOutput}
		for idx := high[hlKey]; idx < len(weakAdjRight[qId][nLk]) && status == ResultNotRelated; idx++ {
			trans2 := weakAdjRight[qId][nLk][idx]
			if trans2.Label.Symbol.Type == pifra.SymbolTypOutput &&
				trans2.Label.Symbol2.Type == pifra.SymbolTypFreshOutput &&
				trans2.Label.Symbol.Value == pi {
				k := trans2.Label.Symbol2.Value
				qXId := trans2.Destination
				qX := rightLts.States[qXId]
				newRho = makeNewRho(nP.Rho, trans.Label.Symbol2.Value, k)
				newLabel = pifra.Label{
					Symbol:  pifra.Symbol{Type: pifra.SymbolTypOutput, Value: pi},
					Symbol2: pifra.Symbol{Type: pifra.SymbolTypFreshOutput, Value: k},
				}
				nPX = FRAConfiguration{
					Process:   pX.Process,
					Registers: pX.Registers,
					Label:     newLabel,
					Rho:       newRho,
					N:         getRegSize(),
				}
				revRho, err := reverseMap(newRho)
				if err != nil {
					high[hlKey] += 1
					continue
				}
				nQX = FRAConfiguration{
					Process:   qX.Process,
					Registers: qX.Registers,
					Rho:       revRho,
					N:         getRegSize(),
				}
				if enableGarbageCollection() {
					err := fixGC(&nPX, &nQX)
					if err != nil {
						high[hlKey] += 1
						continue
					}
				}
				status = preorderGeneric(state, nPX, pXId, nQX, qXId, isLeft)
				if status == ResultRelated {
					state.createAndAddTransition(&nP, &nQ, &nPX, &nQX, isLeft, edgeLabel, transId,
						noKPrime, labelsKey)
				} else {
					high[hlKey] += 1
				}
			}
		}

	} // else tau transition
	if isDebug() {
		fmt.Printf("%d. Exit  processDerivativeGeneric with %s, %s, %s, trans:%s. Status: %d\n",
			stackDepth,
			nP.String(), nQ.String(), fmt.Sprint(isLeft), fmt.Sprint(trans), status)
	}
	stackDepth--
	if status == ResultNotRelated {
		IC.failPD++
	}
	return status
}
