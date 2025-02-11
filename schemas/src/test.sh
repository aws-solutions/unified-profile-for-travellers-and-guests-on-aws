env=$1

echo "Running unit tests in env"

go test -v -failfast main_test.go
if [ $? -ne 0 ]; then
    echo "Unit Test Failed"
    exit 1
fi