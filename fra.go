package main

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/yungene/pifra"
)

// This is a file with FRA structures and related functions

// An extended configuration including the rho register mapping.
type FRAConfiguration struct {
	Process   pifra.Element
	Registers pifra.Registers
	Label     pifra.Label
	Rho       map[int]int
	N         int
}

// An extended LTS based on FRAConfiguration states.
type FRALts struct {
	States      map[int]FRAConfiguration
	Transitions []pifra.Transition
}

// Copied and adapted over from Pifra
// TODO: replace this.
func PrettyPrintRegister(register pifra.Registers, n int) string {
	str := "{"
	//labels := register.Labels()
	reg := register.Registers

	for i := 1; i <= n; i++ {
		//for i, label := range labels {
		if label, ok := reg[i]; ok {
			str = str + "(" + strconv.Itoa(i) + "," + label + "),"

		} else {
			str = str + "(" + strconv.Itoa(i) + "," + "{%}" + "),"
		}
	}
	return str + "}"
}

// Pretty print for FRAConfiguration
func (config FRAConfiguration) String() string {
	return fmt.Sprintf("<%s, %s, rho=%s, n=%d>",
		PrettyPrintRegister(config.Registers, config.N),
		//config.Label.PrettyPrintGraph(),
		pifra.PrettyPrintAst(config.Process),
		fmt.Sprint(config.Rho),
		config.N,
	)
}

// Create an identifying string key for a FRAConfiguration.
func getFRAConfigurationKey(config FRAConfiguration, isLeft bool) string {
	return fmt.Sprintf("<%s,%s,%s,%d,%s>",
		PrettyPrintRegister(config.Registers, config.N),
		pifra.PrettyPrintAst(config.Process),
		fmt.Sprint(config.Rho),
		config.N,
		fmt.Sprint(isLeft),
	)
}

func getFRAPairKey(configLeft FRAConfiguration, configRight FRAConfiguration) string {
	return getFRAConfigurationKey(configLeft, true) + getFRAConfigurationKey(configRight, false)
}

func getConfigurationKey(conf pifra.Configuration) string {
	return pifra.PrettyPrintRegister(conf.Registers) + pifra.PrettyPrintAst(conf.Process)
}

// Pretty print for FRALts
func (lts FRALts) String() string {
	var sb strings.Builder

	sb.WriteString("\nFRALts:: \n")
	sb.WriteString(" States: \n")
	//sb.WriteString(fmt.Sprint(lts.States))
	keys := make([]int, 0)
	for k := range lts.States {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	fmt.Println(fmt.Sprint(keys))
	for _, v := range keys {
		// TODO: fix bug
		fmt.Println(v)
		sb.WriteString(fmt.Sprintf("\t%d : %s\n", v, fmt.Sprint(lts.States[v])))
	}
	sb.WriteString(" Transitions: \n")
	for i := range lts.Transitions {
		sb.WriteString(fmt.Sprintf("\t%d -- %s --> %d \n",
			lts.Transitions[i].Source,
			lts.Transitions[i].Label.PrettyPrintGraph(),
			lts.Transitions[i].Destination))
	}

	return sb.String()
}

// Generate an adjacency list for graph in pifra.Lts.
func ToAdjacency(lts pifra.Lts) map[int][]pifra.Transition {
	adj := make(map[int][]pifra.Transition)
	for i := range lts.Transitions {
		src := lts.Transitions[i].Source
		//dst := lts.Transitions[i].Destination
		adj[src] = append(adj[src], lts.Transitions[i])
	}
	return adj
}

type LabelsKey struct {
	SymbolType1 pifra.SymbolType
	SymbolType2 pifra.SymbolType
}

func getAdvAdjKey(trans *pifra.Transition) LabelsKey {
	return LabelsKey{
		SymbolType1: trans.Label.Symbol.Type,
		SymbolType2: trans.Label.Symbol2.Type,
	}
}

type AdvAdj = map[int]map[LabelsKey][]pifra.Transition

func ToAdvAdjacency(lts pifra.Lts) AdvAdj {
	adj := make(AdvAdj)
	for i := range lts.Transitions {
		src := lts.Transitions[i].Source
		key := getAdvAdjKey(&lts.Transitions[i])
		if _, ok := adj[src]; !ok {
			adj[src] = make(map[LabelsKey][]pifra.Transition)
		}
		adj[src][key] = append(adj[src][key], lts.Transitions[i])
	}
	return adj
}

func getMaxMinRegSize(lts pifra.Lts) int {
	n := 0
	for i := range lts.States {
		reg_size := 0
		for k := range lts.States[i].Registers.Registers {
			if k > reg_size {
				reg_size = k
			}
		}
		if reg_size > n {
			n = reg_size
		}
	}
	return n
}

// ============================================================================
// ============================== GRAPHING ====================================
// ============================================================================

// FRALts to GraphViz DOT format.
func FRALtsGraphViz(lts FRALts) []byte {
	var buf bytes.Buffer
	type StateTmpl struct {
		State int
		Label string
		Attrs string
	}
	type TransTmpl struct {
		Src   int
		Dest  int
		Label string
	}
	const stmpl = "    {{.State}} [{{.Attrs}}label=\"{{.Label}}\"]\n"
	const ttmpl = "    {{.Src}} -> {{.Dest}} [label=\"{{ .Label}}\"]\n"
	stateTmpl := template.Must(template.New("state").Parse(stmpl))
	transTmpl := template.Must(template.New("trans").Parse(ttmpl))

	var states []int
	for state := range lts.States {
		states = append(states, state)
	}
	sort.Ints(states)
	buf.WriteString("digraph {\n")
	for _, id := range states {
		conf := lts.States[id]
		var label string = PrettyPrintRegister(conf.Registers, conf.N) + ", " + fmt.Sprint(conf.Rho) + " ⊢\n" + pifra.PrettyPrintAst(conf.Process)

		var attrs string
		if id == 0 {
			attrs = attrs + "peripheries=2,"
		}

		node := StateTmpl{State: id, Label: label, Attrs: attrs}
		stateTmpl.Execute(&buf, node)
	}
	buf.WriteRune('\n')
	for _, trans := range lts.Transitions {
		transTmpl.Execute(&buf, TransTmpl{
			Src:   trans.Source,
			Dest:  trans.Destination,
			Label: trans.Label.PrettyPrintGraph(),
		})
	}
	buf.WriteString("}\n")
	return buf.Bytes()
}

// TODO: Copied from pifra, needs to be imported instead.
func generateGraphVizFile(lts pifra.Lts) []byte {
	var buf bytes.Buffer
	type StateTmpl struct {
		State int
		Label string
		Attrs string
	}
	type TransTmpl struct {
		Src   int
		Dest  int
		Label string
	}
	const stmpl = "    {{.State}} [{{.Attrs}}label=\"{{.Label}}\"]\n"
	const ttmpl = "    {{.Src}} -> {{.Dest}} [label=\"{{ .Label}}\"]\n"
	stateTmpl := template.Must(template.New("state").Parse(stmpl))
	transTmpl := template.Must(template.New("trans").Parse(ttmpl))

	var states []int
	for state := range lts.States {
		states = append(states, state)
	}
	sort.Ints(states)
	buf.WriteString("digraph {\n")
	for _, id := range states {
		conf := lts.States[id]
		var label string = pifra.PrettyPrintRegister(conf.Registers) +
			" ⊢\n" + pifra.PrettyPrintAst(conf.Process)

		var attrs string
		if id == 0 {
			attrs = attrs + "peripheries=2,"
		}

		node := StateTmpl{State: id, Label: label, Attrs: attrs}
		stateTmpl.Execute(&buf, node)
	}
	buf.WriteRune('\n')
	for _, trans := range lts.Transitions {
		transTmpl.Execute(&buf, TransTmpl{
			Src:   trans.Source,
			Dest:  trans.Destination,
			Label: trans.Label.PrettyPrintGraph(),
		})
	}
	buf.WriteString("}\n")
	return buf.Bytes()
}

func generateBisimGraphVizFile(ltsLeft pifra.Lts, ltsRight pifra.Lts, m map[string]bisimPair) []byte {
	var buf bytes.Buffer
	buf.WriteString("digraph {\n")

	var stateToIdLeft = make(map[string]int)
	var stateToIdRight = make(map[string]int)

	type StateTmpl struct {
		State int
		Label string
		Attrs string
	}
	type TransTmpl struct {
		Src   int
		Dest  int
		Label string
	}
	const stmpl = "    {{.State}} [{{.Attrs}}label=\"{{.Label}}\"]\n"
	const ttmpl = "    {{.Src}} -> {{.Dest}} [label=\"{{ .Label}}\"]\n"
	stateTmpl := template.Must(template.New("state").Parse(stmpl))
	transTmpl := template.Must(template.New("trans").Parse(ttmpl))

	const ttmp2 = "    {{.Src}} -> {{.Dest}} [style=dotted, arrowhead=\"none\"]\n"
	transTmp2 := template.Must(template.New("trans2").Parse(ttmp2))
	type TransTmp2 struct {
		Src  int
		Dest int
	}

	// LTS LEFT
	var states []int
	for state := range ltsLeft.States {
		states = append(states, state)
	}
	sort.Ints(states)
	for _, id := range states {
		conf := ltsLeft.States[id]
		stateToIdLeft[pifraStateKey(&conf)] = id
		var label string = pifra.PrettyPrintRegister(conf.Registers) +
			" ⊢\n" + pifra.PrettyPrintAst(conf.Process)

		var attrs string
		if id == 0 {
			attrs = attrs + "peripheries=2,"
		}

		node := StateTmpl{State: id, Label: label, Attrs: attrs}
		stateTmpl.Execute(&buf, node)
	}
	buf.WriteRune('\n')
	for _, trans := range ltsLeft.Transitions {
		transTmpl.Execute(&buf, TransTmpl{
			Src:   trans.Source,
			Dest:  trans.Destination,
			Label: trans.Label.PrettyPrintGraph(),
		})
	}

	// LTS RIGHT
	var offset int = 999999
	var states2 []int
	for state := range ltsRight.States {
		states2 = append(states2, state)
	}
	sort.Ints(states2)
	for _, id := range states2 {
		conf := ltsRight.States[id]
		stateToIdRight[pifraStateKey(&conf)] = offset + id
		var label string = pifra.PrettyPrintRegister(conf.Registers) +
			" ⊢\n" + pifra.PrettyPrintAst(conf.Process)

		var attrs string
		if id == 0 {
			attrs = attrs + "peripheries=2,"
		}

		node := StateTmpl{State: offset + id, Label: label, Attrs: attrs}
		stateTmpl.Execute(&buf, node)
	}
	buf.WriteRune('\n')
	for _, trans := range ltsRight.Transitions {
		transTmpl.Execute(&buf, TransTmpl{
			Src:   offset + trans.Source,
			Dest:  offset + trans.Destination,
			Label: trans.Label.PrettyPrintGraph(),
		})
	}

	// BISIM
	for _, v := range m {
		id1 := stateToIdLeft[pifraStateKey(&v.LeftState)]
		id2 := stateToIdRight[pifraStateKey(&v.RightState)]
		transTmp2.Execute(&buf, TransTmp2{
			Src:  id1,
			Dest: id2,
		})
	}

	buf.WriteString("}\n")
	return buf.Bytes()
}

func generateBisimGraphVizTexFile(ltsLeft pifra.Lts, ltsRight pifra.Lts, m map[string]bisimPair) []byte {
	// TODO: create flags for these maybe.
	//gvLayout := "rankdir=LR; margin=100"
	gvLayout := "margin=100"
	outputStateNo := false

	type StateTmpl struct {
		State int
		Label string
		Attrs string
	}
	stateTmpl := template.Must(template.New("trans1").Parse(
		"    {{.State}} [{{.Attrs}}texlbl=\"${{.Label}}$\"]\n"))
	type TransTmpl struct {
		Src   int
		Dest  int
		Label string
	}
	transTmp1 := template.Must(template.New("state").Parse(
		"    {{.Src}} -> {{.Dest}} [label=\"\",texlbl=\"${{.Label}}$\"]\n"))
	type TransTmp2 struct {
		Src  int
		Dest int
	}
	transTmp2 := template.Must(template.New("trans2").Parse(
		"    {{.Src}} -> {{.Dest}} [style=dotted, arrowhead=\"none\"]\n"))

	var stateToIdLeft = make(map[string]int)
	var stateToIdRight = make(map[string]int)

	// Start - LTS LEFT
	var buf bytes.Buffer

	gvl := ""
	if gvLayout != "" {
		gvl = "\n    " + gvLayout + "\n"
	}
	buf.WriteString("digraph {" + gvl + "\n")

	buf.WriteString(`    d2toptions="--format tikz --crop --autosize --nominsize";`)
	buf.WriteString("\n")
	buf.WriteString(`    d2tdocpreamble="\usepackage{amssymb}";`)
	buf.WriteString("\n\n")

	var ids []int
	for id := range ltsLeft.States {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		conf := ltsLeft.States[id]
		stateToIdLeft[pifraStateKey(&conf)] = id
		var config string
		if outputStateNo {
			config = "s_{" + strconv.Itoa(id) + "}"
		} else {
			config = `\begin{matrix} ` +
				pifra.PrettyPrintTexRegister(conf.Registers) +
				` \vdash \\ ` +
				pifra.PrettyPrintTexAst(conf.Process) +
				` \end{matrix}`
		}

		var layout string
		if id == 0 {
			layout = layout + `style="double",`
		}
		if ltsLeft.RegSizeReached[id] {
			layout = layout + `style="thick",`
		}

		vertex := StateTmpl{
			State: id,
			Label: config,
			Attrs: layout,
		}
		stateTmpl.Execute(&buf, vertex)
	}

	buf.WriteString("\n")

	for _, edge := range ltsLeft.Transitions {
		edg := TransTmpl{
			Src:   edge.Source,
			Dest:  edge.Destination,
			Label: pifra.PrettyPrintTexGraphLabel(edge.Label),
		}

		transTmp1.Execute(&buf, edg)
	}
	buf.WriteString("\n")
	// LTS RIGHT
	// TODO: magic number
	var offset int = 999999
	var ids2 []int
	for id := range ltsRight.States {
		ids2 = append(ids2, id)
	}
	sort.Ints(ids2)
	for _, id := range ids2 {
		conf := ltsRight.States[id]
		stateToIdRight[pifraStateKey(&conf)] = offset + id
		var config string
		if outputStateNo {
			config = "s_{" + strconv.Itoa(id) + "}"
		} else {
			config = `\begin{matrix} ` +
				pifra.PrettyPrintTexRegister(conf.Registers) +
				` \vdash \\ ` +
				pifra.PrettyPrintTexAst(conf.Process) +
				` \end{matrix}`
		}

		var layout string
		if id == 0 {
			layout = layout + `style="double",`
		}
		if ltsLeft.RegSizeReached[id] {
			layout = layout + `style="thick",`
		}

		vertex := StateTmpl{
			State: offset + id,
			Label: config,
			Attrs: layout,
		}
		stateTmpl.Execute(&buf, vertex)
	}
	buf.WriteString("\n")

	for _, edge := range ltsRight.Transitions {
		edg := TransTmpl{
			Src:   offset + edge.Source,
			Dest:  offset + edge.Destination,
			Label: pifra.PrettyPrintTexGraphLabel(edge.Label),
		}
		transTmp1.Execute(&buf, edg)
	}
	buf.WriteString("\n")
	buf.WriteString("\n")

	// BISIM
	for _, v := range m {
		id1 := stateToIdLeft[pifraStateKey(&v.LeftState)]
		id2 := stateToIdRight[pifraStateKey(&v.RightState)]
		transTmp2.Execute(&buf, TransTmp2{
			Src:  id1,
			Dest: id2,
		})
	}

	buf.WriteString("}\n")
	var output bytes.Buffer
	buf.WriteTo(&output)
	return output.Bytes()
}
