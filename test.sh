pushd .

echo ""
echo "***** TESTING LIBRARY *****"
echo ""
echo "  Running ./go/test_lib.sh"
echo ""

cd ./go
./test_lib.sh

popd