#!/bin/bash -xe

{
    echo ""
    echo "***** TESTING LEAN HELIX LIBRARY *****"
    echo ""
} 2> /dev/null

go test ./... -test.v -test.timeout 3m