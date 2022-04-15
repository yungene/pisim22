package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/yungene/pifra"
)

func main() {
	startTime := time.Now()

	ltsFileNameFlag := flag.String("lts1", "", "[REQUIRED] A path to the LTS file.")
	ltsFileName2Flag := flag.String("lts2", "", "[REQUIRED] A path to the LTS file.")
	regSizeOverrideFlag := flag.Int("n", -1, "The override for the size of the register.")
	verboseFlag := flag.Bool("v", false, "Whether to be verbose.")
	debugFlag := flag.Bool("d", false, "Whether to print debug information.")
	weakBisimFlag := flag.Bool("w", false, "Whether to do weak bisimulation.")
	internalStatsFlag := flag.Bool("is", false, "Whether to show internal stats.")
	outputGraphFlag := flag.Bool("output-graph", false, "Whether to print the produced graph.")
	maxStatesFlag := flag.Int("max-states", 15000, "Max states in an LTS.")
	closureAlgoFlag := flag.Int("closure-algo", 1, "Choice of closure algorithm. 1 for DFS, 2 ofr Floyd-Warshall.")
	gob1FileNameFlag := flag.String("gob1", "", "A path to the gob file.")
	gob2FileNameFlag := flag.String("gob2", "", "A path to the gob file.")
	outFileNameFlag := flag.String("out", "", "A path to the output files.")
	outBisimFileNameFlag := flag.String("output-bisim", "", "A path to the output bisim lts DOT file.")
	garbageCollectionFlag := flag.Bool("gc", false, "Whether to enable garbage collection.")
	flag.Parse()
	verbose = *verboseFlag
	debug = *debugFlag
	weakBisim = *weakBisimFlag
	internalStats = *internalStatsFlag
	enableGC = *garbageCollectionFlag
	outputGraph = *outputGraphFlag
	outputBisimLtsName = *outBisimFileNameFlag

	if *closureAlgoFlag == 1 || *closureAlgoFlag == 2 {
		closureAlgorithmChoice = *closureAlgoFlag
	}

	// regSize := *regSizeOverrideFlag
	// if *regSizeOverrideFlag == -1 {
	// 	regSize = 1073741824
	// }

	var flags = pifra.Flags{
		MaxStates:    *maxStatesFlag,
		RegisterSize: 1073741824,
		DisableGC:    !enableGarbageCollection(),
		Gob:          true,
		Statistics:   isVerbose(),
	}

	pwd, err := os.Getwd()
	check(err)
	// TODO: Fix this. This is risky as it can delete more files than necessary.
	// Proper temporary files should be used. Or better - no temporary files!s
	outFolder := path.Join(pwd, "tmp")
	defer func() {
		os.RemoveAll(outFolder)
	}()

	pifraTimeStart := time.Now()
	if isVerbose() {
		fmt.Printf("Generating an LTS for lts1.\n")
	}
	// TODO: Fix this to use a proper unique name.
	// Or better, do not store the data structure in a file at all and export it
	// directly from pifra.
	outputPath1 := path.Join(outFolder, "lts1"+".gob")
	if *gob1FileNameFlag == "" {
		opts := flags
		opts.OutputFile = outputPath1
		opts.InputFile = *ltsFileNameFlag
		err = pifra.OutputMode(opts)
		check(err)
	} else {
		if isVerbose() {
			fmt.Println("Gob file 1 override is used. No generation done.")
		}
		outputPath1 = *gob1FileNameFlag
	}
	if isVerbose() {
		fmt.Println()
	}
	if isVerbose() {
		fmt.Printf("Generating an LTS for lts2.\n")
	}
	outputPath2 := path.Join(outFolder, "lts2"+".gob")
	if *gob2FileNameFlag == "" {
		opts := flags
		opts.OutputFile = outputPath2
		opts.InputFile = *ltsFileName2Flag
		err = pifra.OutputMode(opts)
		check(err)
	} else {
		if isVerbose() {
			fmt.Println("Gob file 2 override is used. No generation done.")
		}
		outputPath2 = *gob2FileNameFlag
	}
	if isVerbose() {
		fmt.Println()
	}
	if isVerbose() {
		fmt.Printf("Pifra took in total %s time.\n", time.Since(pifraTimeStart))
	}
	bisimStartTime := time.Now()
	var bisimAlgoStartTime time.Time
	if isWeakBisim() && *outFileNameFlag != "" {
		left, err := decodeLTS(outputPath1)
		check(err)
		weakLeft := doWeakTransform(left)
		data := generateGraphVizFile(weakLeft)
		check(writeFile(*outFileNameFlag+"-out.1"+".dot", data))

		right, err := decodeLTS(outputPath2)
		check(err)
		weakRight := doWeakTransform(right)
		data = generateGraphVizFile(weakRight)
		check(writeFile(*outFileNameFlag+"-out.2"+".dot", data))
	} else if isWeakBisim() {

		left, err := decodeLTS(outputPath1)
		check(err)
		prevTime := time.Now()
		weakLeft := doWeakTransform(left)
		transTime := time.Since(prevTime)
		if isVerbose() {
			fmt.Printf("Left. Originally there were %d states and %d transitions. With weak tranform there are now %d states and %d transitions.\n",
				len(left.States), len(left.Transitions), len(weakLeft.States), len(weakLeft.Transitions))
			fmt.Printf("Left translation took %s.\n", transTime)
		}
		right, err := decodeLTS(outputPath2)
		check(err)
		leftTime := time.Now()
		weakRight := doWeakTransform(right)
		transTime2 := time.Since(leftTime)
		if isVerbose() {
			fmt.Printf("Right. Originally there were %d states and %d transitions. With weak tranform there are now %d states and %d transitions.\n",
				len(right.States), len(right.Transitions), len(weakRight.States), len(weakRight.Transitions))
			fmt.Printf("Right translation took %s.\n", transTime2)
			elapsedTime := time.Since(prevTime)
			fmt.Printf("In total, translation took %s.\n\n", elapsedTime)
		}
		bisimAlgoStartTime = time.Now()
		checkBisim(left, right, weakLeft, weakRight, *regSizeOverrideFlag, -1, false)
	} else {

		left, err := decodeLTS(outputPath1)
		check(err)

		right, err := decodeLTS(outputPath2)
		check(err)
		bisimAlgoStartTime = time.Now()
		checkBisim(left, right, left, right, *regSizeOverrideFlag, -1, false)
	}
	fmt.Printf("Bisimulation algo took: %s.\n", time.Since(bisimAlgoStartTime))
	fmt.Printf("Total bisimulation check took (transformation + bisimulation): %s.\n", time.Since(bisimStartTime))
	fmt.Printf("Total execution time (LTS generation + bisimulation): %s.\n", time.Since(startTime))

	if isInternalStats() {
		printInternalStats()
	}
}

// The below code is borrowed from the original pisim that was authored by Basil L. Contovounesios.

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func closeFile(f *os.File) {
	check(f.Close())
}

func writeFile(name string, data []byte) error {
	dir := filepath.Dir(name)
	os.MkdirAll(dir, os.ModePerm)
	return ioutil.WriteFile(name, data, 0644)
}

func decodeLTS(name string) (lts pifra.Lts, err error) {
	file, err := os.Open(name)
	if err != nil {
		return
	}
	defer closeFile(file)
	dec := gob.NewDecoder(file)
	err = dec.Decode(&lts)
	return
}
