#!/bin/sh

echo "Make sure your membuffers git repo is updated and pulled before building!"
echo ""

rm -rf lean_helix.mb.go

### OLD: (uses brew) membufc --go --mock `find . -name "*.proto"`
### NEW: running membufc directly from source to avoid waiting for brew releases
go run $(ls -1 ../membuffers/go/membufc/*.go | grep -v _test.go) --go --mock `find . -name "*.proto"`
#rm `find . -name "*.proto"`

echo ""
echo "Building all packages with go build:"

command 2>&1 go build -a -v lean_helix.mb.go | grep "orbs-network\|.mb.go"