#!/bin/bash -xe

{
    echo ""
    echo "***** TESTING LIBRARY *****"
    echo ""
} 2> /dev/null

go test ./...