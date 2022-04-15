package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yungene/pifra"
)

type HLKey struct {
	Dest string
	Src  string
	Act  string
}

type HLKeyFINP struct {
	Dest   string
	Src    string
	Act    string
	KPrime int
}

const (
	TypeOne int = 1
	TypeTwo int = 2
)

type AKey struct {
	NP        string
	NQ        string
	Type      gLabel
	TransId   int
	KPrime    int
	LabelsKey LabelsKey
}

type CleavelandState struct {
	LeftLts      pifra.Lts
	RightLts     pifra.Lts
	WeakLeftLts  pifra.Lts
	WeakRightLts pifra.Lts
	AdjLeft      AdvAdj
	AdjRight     AdvAdj
	WeakAdjLeft  AdvAdj
	WeakAdjRight AdvAdj
	//NotR     map[string]bool
	States map[uint64]FRAConfiguration
	//Transitions []pifra.Transition
	NextIdnP   uint64
	NextIdnQ   uint64
	RevMap     map[uint64]int
	NStateToId map[string]uint64
	High       map[HLKey]int
	HighTwo    map[HLKeyFINP]int
	Low        map[HLKey]int
	LowTwo     map[HLKeyFINP]int
	//A           []AKey
	G gGraph
}

type ResultType int

const (
	ResultRelated    ResultType = 1
	ResultNotRelated ResultType = 2
)

type gVertex struct {
	A FRAConfiguration
	B FRAConfiguration
}

func newGVertex(a FRAConfiguration, b FRAConfiguration, isLeft bool) gVertex {
	if isLeft {
		return gVertex{a, b}
	} else {
		return gVertex{b, a}
	}
}

func gVertexToString(v *gVertex) string {
	return getFRAPairKey(v.A, v.B)
}

type gGraph struct {
	States map[string]gVertex
	// Transitions []gTransition
	TransitionsSet map[string]*gTransition
	// A map from sourceKey to a vector of gTransitions
	TransitionsSrcMap map[string]map[string]*gTransition
	TransitionsDstMap map[string]map[string]*gTransition
}

func (s *CleavelandState) addTransition(edge gTransition) {
	key := gTransitionKey(edge)
	if _, ok := s.G.TransitionsSet[key]; !ok {
		s.G.TransitionsSet[key] = &edge
	}
	if _, ok := s.G.TransitionsSrcMap[edge.Source][key]; !ok {
		s.G.TransitionsSrcMap[edge.Source] = make(map[string]*gTransition)
	}
	s.G.TransitionsSrcMap[edge.Source][key] = &edge
	if _, ok := s.G.TransitionsDstMap[edge.Destination][key]; !ok {
		s.G.TransitionsDstMap[edge.Destination] = make(map[string]*gTransition)
	}
	s.G.TransitionsDstMap[edge.Destination][key] = &edge

}

func (graph gGraph) String() string {
	var sb strings.Builder

	sb.WriteString("\nFRALts:: \n")
	sb.WriteString(" States: \n")
	keys := make([]string, 0)
	for k := range graph.States {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	//fmt.Println(fmt.Sprint(keys))
	for _, v := range keys {
		// TODO: fix bug
		//fmt.Println(v)
		sb.WriteString(fmt.Sprintf("\t%s : %s\n", v, fmt.Sprint(graph.States[v])))
	}
	if len(keys) == 0 {
		sb.WriteString("\t{}\n")
	}
	sb.WriteString(" Transitions: \n")
	for key := range graph.TransitionsSet {
		var edge *gTransition = graph.TransitionsSet[key]
		sb.WriteString(fmt.Sprintf("\t%s -- %d, %d, %d, %s --> %s \n",
			edge.Source,
			edge.Label,
			edge.TransId,
			edge.KPrime,
			fmt.Sprint(edge.LabelsKey),
			edge.Destination,
		))
	}
	if len(graph.TransitionsSet) == 0 {
		sb.WriteString("\t{}\n")
	}

	return sb.String()
}

type gLabel int

type gTransition struct {
	Source      string
	Destination string
	Label       gLabel
	TransId     int
	KPrime      int
	LabelsKey   LabelsKey
}

func gTransitionKey(trans gTransition) string {
	return fmt.Sprintf("%s,%d,%d,%s,%d,%s",
		trans.Source,
		trans.Label,
		trans.TransId,
		trans.Destination,
		trans.KPrime,
		fmt.Sprint(trans.LabelsKey),
	)
}

const noKPrime int = -999999

const (
	GLabelOne gLabel = 1
	GLabelTwo gLabel = 2
)

// 2^63 - 1
const statesLimit uint64 = 9223372036854775807

// #############################################################################
// ################################ STATE ######################################
// #############################################################################
func NewCleavelandState(leftLts pifra.Lts, rightLts pifra.Lts,
	weakLeftLts pifra.Lts, weakRightLts pifra.Lts) *CleavelandState {
	var state CleavelandState
	state.LeftLts = leftLts
	state.RightLts = rightLts
	state.WeakLeftLts = weakLeftLts
	state.WeakRightLts = weakRightLts
	state.AdjLeft = ToAdvAdjacency(leftLts)
	state.AdjRight = ToAdvAdjacency(rightLts)
	state.WeakAdjLeft = ToAdvAdjacency(weakLeftLts)
	state.WeakAdjRight = ToAdvAdjacency(weakRightLts)

	// A set \hat{R} (notR) that stores all state pairs that have been determined
	// to not be related.
	//state.NotR = make(map[string]bool)
	// Values to build the result LTS
	state.States = make(map[uint64]FRAConfiguration)
	//var transitions []pifra.Transition

	// We have the original leftLts and rightLts. However, the algorithm will use
	// new type of states -- it will use FRAConfigurations. We need to keep a map
	// between the new states and old states, as we need the transitions from the
	// original states/graph.
	// Let the original state from leftLts be P, and the original state from rightLts
	// be Q, then let the new state from newLeftLts be nP, and sim. for newRightLts
	// be nQ. So P is of type Configuration, while NP of type FRAConfiguration.
	// We then have that each nQ* is derived from some Q, so we need to maintain a
	// a mapping. And we also have that each nQ* is paired with nP* corresponding
	// to the proposed bisimulation relation R.
	// Each P* or Q* has an Id which is its key in pifra.Lts.States map.
	// Each nP* or nQ* also has an Id which is its key in states map.
	state.NextIdnP = 0
	// TODO: this limits the number of unique states
	state.NextIdnQ = statesLimit
	// revMap is a map from Id of nP* to corresponding Id of P*.
	state.RevMap = make(map[uint64]int)

	// A map of state keys to ids. Should only be used for nP* and nQ* states.
	state.NStateToId = make(map[string]uint64)

	// A set of pointers high(p', q, a), where
	state.High = make(map[HLKey]int)
	state.Low = make(map[HLKey]int)
	state.HighTwo = make(map[HLKeyFINP]int)
	state.LowTwo = make(map[HLKeyFINP]int)
	// A set A which records the pairs that need to be re-examined.
	//state.A = make(map[AKey]bool)

	state.G = gGraph{}
	state.G.States = make(map[string]gVertex)
	state.G.TransitionsSrcMap = make(map[string]map[string]*gTransition)
	state.G.TransitionsDstMap = make(map[string]map[string]*gTransition)
	state.G.TransitionsSet = make(map[string]*gTransition)

	return &state
}

// pid is the state in original old LTS.
func (s *CleavelandState) addNState(config FRAConfiguration, pid int, isLeft bool) uint64 {
	key := getFRAConfigurationKey(config, isLeft)
	if id, ok := s.NStateToId[key]; !ok {
		var nId uint64
		if isLeft {
			nId = s.NextIdnP
		} else {
			nId = s.NextIdnQ
		}
		s.States[nId] = config
		s.RevMap[nId] = pid
		s.NStateToId[key] = nId
		if isLeft {
			s.NextIdnP = s.NextIdnP + 1
			if s.NextIdnP > statesLimit {
				panic("Too many states for the tool to handle.")
			}
		} else {
			s.NextIdnQ = s.NextIdnQ + 1
			if s.NextIdnQ < statesLimit {
				panic("Too many states for the tool to handle.")
			}
		}
		return nId
	} else {
		return id
	}
}

func (s *CleavelandState) addNQState(config FRAConfiguration, pid int) uint64 {
	return s.addNState(config, pid, false)
}

func (s *CleavelandState) addNPState(config FRAConfiguration, pid int) uint64 {
	return s.addNState(config, pid, true)
}

func (s *CleavelandState) removeEdgeWithKey(edgeKey string, edge *gTransition) {
	delete(s.G.TransitionsSet, edgeKey)
	delete(s.G.TransitionsSrcMap[edge.Source], edgeKey)
	delete(s.G.TransitionsDstMap[edge.Destination], edgeKey)
}

func (s *CleavelandState) removeEdge(edge *gTransition) {
	edgeKey := gTransitionKey(*edge)
	s.removeEdgeWithKey(edgeKey, edge)
}

func (s *CleavelandState) removeIncidentEdges(vertexKey string) {
	for j := range s.G.TransitionsSrcMap[vertexKey] {
		edge := s.G.TransitionsSrcMap[vertexKey][j]
		s.removeEdge(edge)
	}
	for j := range s.G.TransitionsDstMap[vertexKey] {
		edge := s.G.TransitionsDstMap[vertexKey][j]
		s.removeEdge(edge)
	}
}

func (s *CleavelandState) populateA(A []AKey, vertexKey string) []AKey {
	for edgeKey := range s.G.TransitionsSrcMap[vertexKey] {
		edge := s.G.TransitionsSrcMap[vertexKey][edgeKey]
		sourceVertKey := edge.Source
		_, ok1 := s.G.States[sourceVertKey]
		if !ok1 {
			if isDebug() {
				fmt.Printf("ISSUE with sourceVertKey: %s\n", sourceVertKey)
			}
		}
		destVertKey := edge.Destination
		destVert, ok2 := s.G.States[destVertKey]
		if !ok2 {
			if isDebug() {
				fmt.Printf("ISSUE with destVertKey: %s\n", destVertKey)
			}
		}
		if !ok1 || !ok2 {
			s.removeEdgeWithKey(edgeKey, edge)
			continue
		}
		act := edge.Label
		if sourceVertKey == vertexKey {
			A = append(A, AKey{
				NP:        getFRAConfigurationKey(destVert.A, true),
				NQ:        getFRAConfigurationKey(destVert.B, false),
				Type:      act,
				TransId:   edge.TransId,
				KPrime:    edge.KPrime,
				LabelsKey: edge.LabelsKey,
			})
		}
	}
	return A
}

func (s *CleavelandState) createAndAddTransition(
	destL *FRAConfiguration, destR *FRAConfiguration,
	srcL *FRAConfiguration, srcR *FRAConfiguration,
	isLeft bool,
	edgeLabel gLabel,
	transId int,
	kPrime int,
	labelsKey LabelsKey,
) {

	s.addTransition(*createEdge(destL, destR, srcL, srcR, isLeft, edgeLabel, transId, kPrime, labelsKey))
}

func createEdge(
	destL *FRAConfiguration, destR *FRAConfiguration,
	srcL *FRAConfiguration, srcR *FRAConfiguration,
	isLeft bool,
	edgeLabel gLabel,
	transId int,
	kPrime int,
	labelsKey LabelsKey,
) *gTransition {
	var vertexDest gVertex = newGVertex(*destL, *destR, isLeft)
	var vertexDestKey string = gVertexToString(&vertexDest)
	var vertexSrc gVertex = newGVertex(*srcL, *srcR, isLeft)
	var vertexSrcKey string = gVertexToString(&vertexSrc)
	return &gTransition{vertexSrcKey, vertexDestKey, edgeLabel,
		transId, kPrime, labelsKey}
}

func makeNewRho(oldRho map[int]int, idx int, val int) map[int]int {
	newRho := make(map[int]int)
	for kk, vv := range oldRho {
		// make sure it stays bijective
		if vv != val {
			newRho[kk] = vv
		}
	}
	newRho[idx] = val
	return newRho
}

// Garbage collection
func fixGC(nPX *FRAConfiguration, nQX *FRAConfiguration) error {
	var rho map[int]int = make(map[int]int)
	for k, v := range nPX.Rho {
		rho[k] = v
	}
	for lid, rid := range rho {
		// check if either rid or lid points to an empty register.
		if _, ok := nPX.Registers.Registers[lid]; !ok {
			delete(rho, lid)
		}
		if _, ok := nQX.Registers.Registers[rid]; !ok {
			delete(rho, lid)
		}
	}
	nPX.Rho = rho
	revRho, err := reverseMap(rho)
	if err != nil {
		return err
	}
	nQX.Rho = revRho
	return nil
}

// Experimental:
type bisimPair struct {
	LeftState  pifra.Configuration
	RightState pifra.Configuration
}

func pifraStateKey(conf *pifra.Configuration) string {
	return pifra.PrettyPrintRegister(conf.Registers) + " ⊢ " + pifra.PrettyPrintAst(conf.Process)
}

func statePairKey(lstate *pifra.Configuration, rstate *pifra.Configuration) string {
	return pifraStateKey(lstate) + " <---> " + pifraStateKey(rstate)
}

func getBisimilarStates(state *CleavelandState) map[string]bisimPair {
	res := make(map[string]bisimPair)
	for _, v := range state.G.States {
		lstate := pifra.Configuration{
			Process:   v.A.Process,
			Registers: v.A.Registers,
		}
		rstate := pifra.Configuration{
			Process:   v.B.Process,
			Registers: v.B.Registers,
		}

		pairKey := statePairKey(&lstate, &rstate)
		if _, ok := res[pairKey]; !ok {
			res[pairKey] = bisimPair{lstate, rstate}
		}
	}
	return res
}

func bisimilarStatesToString(m map[string]bisimPair) string {
	var sb strings.Builder
	for k, _ := range m {
		sb.WriteString(k + "\n")
	}
	return sb.String()
}
