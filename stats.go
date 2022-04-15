package main

import (
	"fmt"
	"strings"
)

var IC ICounters

type ICounters struct {
	enterToPreorder         int
	fullExecutePreorder     int
	preorderStackDepth      int
	maxPreorderStackDepth   int
	enterProcessDerivatives int
	tauRule                 int
	inp1Rule                int
	inp2Rule                int
	finpRule                int
	outRule                 int
	foutRule                int
	reevalA                 int
	failPD                  int
}

func (ic *ICounters) resetBisim() {
	IC = ICounters{}
}

func (ic *ICounters) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("maxPreorderStackDepth: %d\n", IC.maxPreorderStackDepth))
	sb.WriteString(fmt.Sprintf("enterToPreorder: %d\n", IC.enterToPreorder))
	sb.WriteString(fmt.Sprintf("fullExecutePreorder: %d\n", IC.fullExecutePreorder))
	sb.WriteString(fmt.Sprintf("reevalA: %d\n", IC.reevalA))
	sb.WriteString(fmt.Sprintf("enterProcessDerivatives: %d\n", IC.enterProcessDerivatives))
	sb.WriteString(fmt.Sprintf("failPD: %d\n", IC.failPD))

	sb.WriteString(fmt.Sprintf("Rules stats: \n"))
	sb.WriteString(fmt.Sprintf("\t tauRule: %d\n", IC.tauRule))
	sb.WriteString(fmt.Sprintf("\t inp1Rule: %d\n", IC.inp1Rule))
	sb.WriteString(fmt.Sprintf("\t inp2Rule: %d\n", IC.inp2Rule))
	sb.WriteString(fmt.Sprintf("\t finpRule: %d\n", IC.finpRule))
	sb.WriteString(fmt.Sprintf("\t outRule: %d\n", IC.outRule))
	sb.WriteString(fmt.Sprintf("\t foutRule: %d\n", IC.foutRule))

	return sb.String()
}

//type

func printInternalStats() {
	fmt.Println()
	fmt.Println("Internal counters")
	fmt.Println(IC.String())
}
