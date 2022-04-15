package main

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/yungene/pifra"
)

const BIG_TESTS_FLAG string = "PISIM_BIG_TESTS"

var bisim_files = []string{
	"jev-a1",
	"jev-a2",
	"jev-a3",
	"jev-a4",
	"jev-non-det",
	"jev-finp2-1",
	"jev-gc-2",
	"jev-gc-3",
	"jev-sangiorgi-open-bisim",
	"jev-wc-1",
	"milner-3-7",
	"milner-5-14",
	"sangiorgi-ex-1-4-11",
	"sangiorgi-book-p65",
	"jev-sangiorgi-fig-1-7",
	"jev-vk-fin-st3",
}

var weak_bisim_files = []string{
	"milner-6-14-2-weak",
	"milner-cycler-02",
	"milner-cycler-03",
	"buffer-2x1",
	"buffer-3",
	"cleav-turner-choice",
	"mwb-bool-not",
}

var weak_bisim_big_files = []string{
	"milner-job-shop",
	"milner-cycler-04",
	"milner-cycler-05",
	"handover-no-error",
	//"handover-error-handl",
	//"handover-impl",
	"cleav-abp-jp",
}

// Examples that are neither strongly nor weakly bisimilar
var fully_not_bisim_files = []string{
	"jev-non-det-2",
	"jev-diff-names-1",
	"jev-tau-1",
	"jev-tau-2",
	"jev-vk-fin-st2",
	"jev-ne-1",
	"buffer-2x1-deadlock",
	"milner-3-10",
	"milner-6-12-1",
	"cleav-abp-bv",
}

var flags = pifra.Flags{
	MaxStates:    3000,
	RegisterSize: 1073741824,
	DisableGC:    true,
	Gob:          true,
}

func getPwd(t *testing.T) string {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return pwd
}

// Test for false negatives
func TestBisim(t *testing.T) {
	pwd := getPwd(t)
	// Test directory.
	testFolder := path.Join(pwd, "test", "bisimilar")
	// Output directory.
	outFolder := path.Join(pwd, "test", "bisimilar", "out")
	defer cleanFolder(t, outFolder)

	var testFiles []string = bisim_files
	generateLts(t, testFolder, outFolder, testFiles, flags)
	genericBisimTest(t, testFolder, outFolder, testFiles, ResultRelated, flags, false)
}

func TestStrongBisimImpliesWeakBisim(t *testing.T) {
	pwd := getPwd(t)
	// Test directory.
	testFolder := path.Join(pwd, "test", "bisimilar")
	// Output directory.
	outFolder := path.Join(pwd, "test", "bisimilar", "out")
	defer cleanFolder(t, outFolder)

	var testFiles []string = bisim_files
	generateLts(t, testFolder, outFolder, testFiles, flags)
	genericBisimTest(t, testFolder, outFolder, testFiles, ResultRelated, flags, true)
}

func TestWeakBisim(t *testing.T) {
	pwd := getPwd(t)
	// Test directory.
	testFolder := path.Join(pwd, "test", "weak-bisimilar")
	// Output directory.
	outFolder := path.Join(pwd, "test", "weak-bisimilar", "out")
	defer cleanFolder(t, outFolder)

	var testFiles []string = weak_bisim_files
	generateLts(t, testFolder, outFolder, testFiles, flags)
	genericBisimTest(t, testFolder, outFolder, testFiles, ResultRelated, flags, true)
}

func TestWeakBisimBig(t *testing.T) {
	_, ok := os.LookupEnv(BIG_TESTS_FLAG)
	if ok {
		pwd := getPwd(t)
		// Test directory.
		testFolder := path.Join(pwd, "test", "weak-bisimilar")
		// Output directory.
		outFolder := path.Join(pwd, "test", "weak-bisimilar", "out")
		defer cleanFolder(t, outFolder)

		var testFiles []string = weak_bisim_big_files
		flags_ := flags
		// Can take a while to generate an LTS.
		flags_.MaxStates = 15000
		generateLts(t, testFolder, outFolder, testFiles, flags_)
		genericBisimTest(t, testFolder, outFolder, testFiles, ResultNotRelated, flags_, false)
		genericBisimTest(t, testFolder, outFolder, testFiles, ResultRelated, flags_, true)
	}
}

func TestNotStrongBisimButWeakBisim(t *testing.T) {
	pwd := getPwd(t)
	// Test directory.
	testFolder := path.Join(pwd, "test", "weak-bisimilar")
	// Output directory.
	outFolder := path.Join(pwd, "test", "weak-bisimilar", "out")
	defer cleanFolder(t, outFolder)

	var testFiles []string = weak_bisim_files
	generateLts(t, testFolder, outFolder, testFiles, flags)
	genericBisimTest(t, testFolder, outFolder, testFiles, ResultNotRelated, flags, false)
	genericBisimTest(t, testFolder, outFolder, testFiles, ResultRelated, flags, true)
}

func TestFullyNotBisim(t *testing.T) {
	pwd := getPwd(t)
	// Test directory.
	testFolder := path.Join(pwd, "test", "not-bisimilar")
	// Output directory.
	outFolder := path.Join(pwd, "test", "not-bisimilar", "out")
	// Remove output directory when finished.
	defer cleanFolder(t, outFolder)
	var testFiles []string = fully_not_bisim_files
	generateLts(t, testFolder, outFolder, testFiles, flags)
	genericBisimTest(t, testFolder, outFolder, testFiles, ResultNotRelated, flags, false)
	genericBisimTest(t, testFolder, outFolder, testFiles, ResultNotRelated, flags, true)
}

// Test for garbage collection
func TestGC(t *testing.T) {
	enableGC = true
	t.Run("TestBisimGC", TestBisim)
	t.Run("TestStrongBisimImpliesWeakBisimGC", TestStrongBisimImpliesWeakBisim)
	t.Run("TestWeakBisimGC", TestWeakBisim)
	t.Run("TestWeakBisimBigGC", TestWeakBisimBig)
	t.Run("TestNotStrongBisimButWeakBisimGC", TestNotStrongBisimButWeakBisim)
	t.Run("TestFullyNotBisimGC", TestFullyNotBisim)
	enableGC = false
}

func cleanFolder(t *testing.T, outFolder string) {
	if !t.Failed() {
		os.RemoveAll(outFolder)
	}
}

func generateLts(t *testing.T, testFolder string, outFolder string,
	testFiles []string, flags pifra.Flags) {
	for _, testFile := range testFiles {
		for i := 1; i < 3; i++ {
			fileName := fmt.Sprintf("%s.%d", testFile, i)
			outputPath := path.Join(outFolder, fileName+".gob")
			testPath := path.Join(testFolder, fileName+".pi")

			opts := flags
			opts.OutputFile = outputPath
			opts.InputFile = testPath
			if err := pifra.OutputMode(opts); err != nil {
				t.Error(err)
			}
		}
	}
}

func genericBisimTest(t *testing.T, testFolder string, outFolder string,
	testFiles []string, expectedRes ResultType, flags pifra.Flags,
	weakBisim bool) {
	for _, testFile := range testFiles {
		// Do the test
		fmt.Printf("\n%s\n", testFile)
		fileNameLeft := fmt.Sprintf("%s.%d", testFile, 1)
		fileNameRight := fmt.Sprintf("%s.%d", testFile, 2)
		filePath1 := path.Join(outFolder, fileNameLeft+".gob")
		filePath2 := path.Join(outFolder, fileNameRight+".gob")

		left, err := decodeLTS(filePath1)
		if err != nil {
			t.Errorf("Error parsing rights LTS at %s. Error: %s.\n", filePath1, fmt.Sprint(err))
		}
		leftWeak := left
		if weakBisim {
			leftWeak = doWeakTransform(left)
		}

		right, err := decodeLTS(filePath2)
		if err != nil {
			t.Errorf("Error parsing rights LTS at %s. Error: %s.\n", filePath2, fmt.Sprint(err))
		}
		rightWeak := right
		if weakBisim {
			rightWeak = doWeakTransform(right)
		}

		status := checkBisim(left, right, leftWeak, rightWeak, -1, -1, false)
		//printInternalStats()
		if status != expectedRes {
			t.Log(status)
			t.Logf("Bisimilation was not identified correctly for %s. Expected %d, but got %d.\n",
				testFile, expectedRes, status)
			t.Fail()
		}
		// check the symmetry
		statusSym := checkBisim(right, left, rightWeak, leftWeak, -1, -1, false)
		if status != statusSym {
			t.Logf("Symmetry does not hold for %s. Status was %d, but symmetric result was %d",
				testFile, status, statusSym)
			t.Fail()
		}
	}
}
