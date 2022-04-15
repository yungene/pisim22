# Pisim22

The pisim22 is a tool for equivalence checking of pi-calculus models. Specifically, the tool supports for weak and strong early bisimulation checking. 

It uses [Fresh-Register Automata](http://www.cs.ox.ac.uk/people/nikos.tzevelekos/FRA_11.pdf) by [Nikos Tzevelekos](http://www.tzevelekos.org/) as an intermediate representation. It uses [pifra](https://github.com/sengleung/pifra) by Seng Leung for translation from pi-calculus into the intermediate representation. It then uses a custom on-the-fly algorithm to check for n-bisimulation, which corresponds to normal bisimulation in pi-calculus.

This work was carried out as part of a master's dissertation with [Dr. Vasileios Koutavas](https://www.scss.tcd.ie/Vasileios.Koutavas/) at Trinity College Dublin in 2021/2022.

## How to use

The tool supports checking two systems specified in pi-calculus for bisimulation. The mainline is in `pisim22.go`. See it for some _currently_ hardcoded limits. The "old" mainline that took gob files as inputs is in `frasim.go`.

To use the tool use the following commands:
- Build the code:
```
go build
```
- Do strong bisimulation check:
```
./pisim22 -lts1 test/bisimilar/jev-a2.1.pi -lts2 test/bisimilar/jev-a2.2.pi

./pisim22 -lts1 test/not-bisimilar/jev-diff-names-1.1.pi -lts2 test/not-bisimilar/jev-diff-names-1.2.pi 
```
- Do weak bisimulation check (use `-w` flag):
```
./pisim22 -lts1 test/bisimilar/jev-a2.1.pi -lts2 test/bisimilar/jev-a2.2.pi -w
```

Key supported flags (or use `./pisim22 --help`):
- `lts1` -- path to the first LTS.
- `lts2` -- path to the second LTS.
- `w` -- enable weak bisimulation.
- `gc` -- enable garbage collection (in both pifra and pisim22).
- `n` -- override for the register size.
- `v` -- whether to be more verbose. E.g. to see the names in the registers of the starting states.
- `d` -- debug mode. Only useful for very small systems.
- `max-states` - max states in a generated LTS. Supplied to pifra when generating lts1 and/or lts2.
- `is` -- whether to print out internal statistics.
- `output-graph` -- whether to print out the bisimulation graph as defined in the algorithm.
- `output-bisim` -- if specified then path for the generated bisimulation LTS. See further for details.

### Writing pi-calculus

#### Notes on Pifra

Few reminders about Pifra:
 - It does not support polyadic pi-calculus.
 - It does not support distinctions, so be extremely careful with constants. See handover test case for an example of how this can be overcome.
 - Be careful of the equality (inequality) conditions and their scoping. The parser is defined as `[a=a]P`, so always surround the conditional expression with brackets. E.g. always have `([a=a]P)`. Otherwise you might get the following: `i(x).([x=a]P + [x=b]Q)` will be expanded as `i(x).( [x=a](P + [x=b]Q) )`, which is likely not be as expected.

### Unit testing

The repository also includes some unit tests. To run them execute the following commands:
```
go build
go clean -testcache
go test
```

The systems used for testing can be inspected under the `test/bisimilar` and `test/not-bisimilar` directories.

There is also an option to run "big" tests that will require O(minutes) to finish. To run them use the following, adjusting the timeout if necessary:
```
PISIM_BIG_TESTS=1
export PISIM_BIG_TESTS=1
go clean -testcache
go test --timeout 30m
unset PISIM_BIG_TESTS
```

## Repository structure

### pisim22.go

The mainline that combines Pifra together with FRA LTS bisimulation check.

### frasim.go

"Old" mainline. However, `frasim_test.go` still contains the main tests.

### fra.go

Helper code and data structures related to FRA.

### utils.go

General helper code.

### weak_bisim.go

Code for transformation of a strong LTS into a weak LTS.

### bisim.go

A file with all the main logic for bisimulation checking.

### bisim_ds.go

Data structures and helper code for `bisim.go`.

## Extra features

### Generate bisimulation LTS

It is possible to generate a merged LTS for the specified input models with all the bisimulation states linked with dotted arcs. The generated graph is in GraphViz DOT format and only makes sense when bisimulation is true, otherwise the linked states are only equivalent up-to the information processed by the algorithm. To use this feature, use the `output-bisim` flag:

```
./pisim22 -lts1 test/bisimilar/jev-a2.1.pi -lts2 test/bisimilar/jev-a2.2.pi -output-bisim jev-a2
```

The above command will generate a `jev-a2.bisim.dot` file. WIll also generate a TeX file `jev-a2.bisim.tex.dot`.

To create a PDF from TeX use the following:
```
dot2tex -o jev-a2.bisim.tex jev-a2.bisim.tex.dot && pdflatex jev-a2.bisim.tex -output-directory .
```

### Change the algorithm used for calculating transitive closure

The algorithm can be explicitly switched via flag `closure-algo`, and the following values are accepted:

 - 1 -- for DFS-based closure algorithm.
 - 2 -- for Floyd-Warshall algorithm.

The choice of algorithm can slightly affect the performance of the translation depending on the sparsity of the graph. The first is better for sparse graphs, while Floyd-Warshall might be better for dense graphs.

### Avoid generating LTS

It is possible to avoid generating the LTS, and instead to pass the `gob` files directly. The `gob` files however still need to be in the format that pifra generates them. Might be useful when pifra takes long to generate the LTS. To use `gob` files, do:

```
./pisim22 -gob1 test/bisimilar/jev-a2.1.gob -gob2 test/bisimilar/jev-a2.2.gob
```

### Output the weakly transformed LTSs

It is possible to just output the produced weakly transformed LTSs and not do equivalence checking. Might be useful for debugging and testing. To do that use `-out` flag, supplying the prefix path for the created files.
