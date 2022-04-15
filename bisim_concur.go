package main

// #############################################################################
// ########################### MULTI-THREADING #################################
// #############################################################################
// func worker(linkChan chan map[int]int, wg *sync.WaitGroup,
// 	leftLts pifra.Lts, rightLts pifra.Lts, findAllFlag bool, finish *bool) {
// 	// Decreasing internal counter for wait-group as soon as goroutine finishes
// 	defer wg.Done()

// 	for m := range linkChan {
// 		// Analyze value and do the job here
// 		res, err := cleavelandBisim(leftLts, rightLts, m)
// 		if !findAllFlag && err == nil && res == ResultRelated {
// 			*finish = true
// 			fmt.Println(res)
// 		}
// 	}
// }

// func checkBisimMultiThread(leftLts pifra.Lts, rightLts pifra.Lts, regSizeOverride int,
// 	debugSpecificFlag int, findAllFlag bool) ResultType {
// 	resetBisim()
// 	var res = ResultNotRelated
// 	var finish = false
// 	toFinish := func() bool { return finish }
// 	counter := 0
// 	if regSizeOverride > 0 {
// 		setRegSize(regSizeOverride)
// 	} else {
// 		n1 := getMaxMinRegSize(leftLts)
// 		n2 := getMaxMinRegSize(rightLts)
// 		if n2 > n1 {
// 			setRegSize(n2)
// 		} else {
// 			setRegSize(n1)
// 		}
// 	}

// 	lCh := make(chan map[int]int)
// 	wg := new(sync.WaitGroup)
// 	for i := 0; i < 32; i++ {
// 		wg.Add(1)
// 		go worker(lCh, wg, leftLts, rightLts, findAllFlag, &finish)
// 	}
// 	generateMappings(REG_SIZE, func(m map[int]int) {
// 		if isVerbose() {
// 			fmt.Printf("\n%d.Processing %s.\n", counter, fmt.Sprint(m))
// 		}
// 		if debugSpecificFlag >= 0 && counter == debugSpecificFlag {
// 			debug = true
// 		} else if debugSpecificFlag >= 0 {
// 			debug = false
// 		}
// 		if !toFinish() {
// 			lCh <- m
// 		}
// 		counter++
// 	},
// 		toFinish)
// 	close(lCh)

// 	// Waiting for all goroutines to finish (otherwise they die as main routine dies)
// 	wg.Wait()
// 	fmt.Printf("\nTotal number of callback calls performed for these inputs for N=%d is %d.\n", REG_SIZE, counter)
// 	return res
// }
