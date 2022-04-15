# Useful commands

Build the code:
```
go build
```

Generate the gob files:
```
./pifra.exe ../../../pifra/test/jev-a4.pi --output-gob --output jev-a4.gob -d
```

Translate the LTS:
```
./pisim22.exe -lts1 ../basil-pifra/pifra/pifra/jev-a4.gob -n 2 -v -out test/jev-a4
```

Generate the PDF:
```
dot -Tpdf test/jev-a4-out0.dot > test/jev-a4-out0.pdf
```

Test:
```
go clean -testcache
go test
```

Test big files (Note: LTS generation might take a while):
```
export PISIM_BIG_TESTS=1
PISIM_BIG_TESTS=1
go clean -testcache
go test --timeout 30m
unset PISIM_BIG_TESTS
```