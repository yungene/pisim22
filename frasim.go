package main

import (
	"github.com/yungene/pifra"
)

// GLOBALS
var verbose bool = false

func isVerbose() bool {
	return verbose
}

var debug bool = false

func isDebug() bool {
	return debug
}

var weakBisim bool = false

func isWeakBisim() bool {
	return weakBisim
}

var MAX_ITER int = 30

func getMaxIter() int {
	return MAX_ITER
}

var REG_SIZE int

func getRegSize() int {
	return REG_SIZE
}

func setRegSize(v int) {
	REG_SIZE = v
}

var internalStats bool = false

func isInternalStats() bool {
	return internalStats
}

var enableGC bool = false

func enableGarbageCollection() bool {
	return enableGC
}

var outputGraph bool = false

func isOutputGraph() bool {
	return outputGraph
}

var outputBisimLtsName string = ""

func getOutputBisimLtsName() string {
	return outputBisimLtsName
}

var closureAlgorithmChoice int = 1

func getClosureAlgortihmChoice() int {
	return closureAlgorithmChoice
}

// CONSTANTS
const NULL_REG = 0

// ============================================================================
// ================================= MAIN =====================================
// ============================================================================

func init() {
	pifra.RegisterGobs()
}

// func main() {
// 	startTime := time.Now()
// 	// if len(os.Args) < 2 {
// 	// 	log.Fatalln("Wrong number of arguments")
// 	// }
// 	ltsFileNameFlag := flag.String("lts1", "", "A path to the LTS file.")
// 	ltsFileName2Flag := flag.String("lts2", "", "A path to the LTS file.")
// 	outFileNameFlag := flag.String("out", "", "A path to the output files.")
// 	regSizeOverrideFlag := flag.Int("n", -1, "The override for the size of the register.")
// 	verboseFlag := flag.Bool("v", false, "Whether to be verbose.")
// 	debugFlag := flag.Bool("d", false, "Whether to print debug information.")
// 	weakBisimFlag := flag.Bool("w", false, "Whether to do weak bisimulation.")
// 	debugSpecificFlag := flag.Int("dd", -1, "Debug particular configuration")
// 	findAllFlag := flag.Bool("find_all", false, "Whether to find all bisimulation.")
// 	flag.Parse()
// 	verbose = *verboseFlag
// 	debug = *debugFlag
// 	weakBisim = *weakBisimFlag

// 	if isWeakBisim() && *outFileNameFlag != "" {
// 		left, err := decodeLTS(*ltsFileNameFlag)
// 		check(err)

// 		weakLeft := doWeakTransform(left)
// 		data := generateGraphVizFile(weakLeft)
// 		check(writeFile(*outFileNameFlag+"-out"+".dot", data))
// 	} else if isWeakBisim() {
// 		left, err := decodeLTS(*ltsFileNameFlag)
// 		check(err)
// 		weakLeft := doWeakTransform(left)

// 		fmt.Printf("Left. Originally there were %d states and %d transitions. With weak tranform there are now %d states and %d transitions.\n",
// 			len(left.States), len(left.Transitions), len(weakLeft.States), len(weakLeft.Transitions))
// 		fmt.Printf("Left translation took %s.\n", time.Since(startTime))

// 		right, err := decodeLTS(*ltsFileName2Flag)
// 		check(err)
// 		weakRight := doWeakTransform(right)

// 		fmt.Printf("Right. Originally there were %d states and %d transitions. With weak tranform there are now %d states and %d transitions.\n",
// 			len(right.States), len(right.Transitions), len(weakRight.States), len(weakRight.Transitions))
// 		elapsedTime := time.Since(startTime)
// 		fmt.Printf("Translation took %s.\n", elapsedTime)
// 		status := checkBisim(left, right, weakLeft, weakRight, *regSizeOverrideFlag, *debugSpecificFlag, *findAllFlag)
// 		fmt.Println(status)
// 	} else {

// 		left, err := decodeLTS(*ltsFileNameFlag)
// 		check(err)

// 		right, err := decodeLTS(*ltsFileName2Flag)
// 		check(err)

// 		status := checkBisim(left, right, left, right, *regSizeOverrideFlag, *debugSpecificFlag, *findAllFlag)
// 		fmt.Println(status)
// 	}
// 	elapsedTime := time.Since(startTime)
// 	fmt.Printf("Execution took %s.\n", elapsedTime)

// 	//var res =
// 	//translateLTS(left, right, *regSizeOverrideFlag)
// 	//fmt.Println(res)
// 	// if *outFileNameFlag != "" {
// 	// 	for i := range res {
// 	// 		data := FRALtsGraphViz(res[i])
// 	// 		check(writeFile(*outFileNameFlag+"-out"+strconv.Itoa(i)+".dot", data))
// 	// 	}
// 	// }
// }
