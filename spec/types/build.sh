#!/bin/sh

echo "Make sure your membuffers git repo is updated and pulled before building!"
echo ""

mkdir -p ../../primitives
mkdir -p ./go

#rm -rf lean_helix.mb.go
#rm -rf ./go/*
rm -f ./go/leanhelix/lean_helix.mb.go ./go/primitives/lean_helix_primitives.mb.go
cp -r ../proto/* ./go

### OLD: (uses brew) membufc --go --mock `find . -name "*.proto"`
### NEW: running membufc directly from source to avoid waiting for brew releases
echo "Proto files:"
find . -name "*.proto"

go run $(ls -1 ../../../membuffers/go/membufc/*.go | grep -v _test.go) --go --mock --go-ctx `find . -name "*.proto"`
#rm `find . -name "*.proto"`
rm -f ./go/leanhelix/lean_helix.proto ./go/primitives/lean_helix_primitives.proto

echo ""
echo "Building all packages with go build:"

command 2>&1 go build -a -v ./go/... | grep "orbs-network\|.mb.go"

mv ./go/leanhelix/lean_helix.mb.go ../..
mv ./go/primitives/lean_helix_primitives.mb.go ../../primitives

# TODO Add go fmt to the generated .mb.go files