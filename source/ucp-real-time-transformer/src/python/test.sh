
env=$1
bucket=$2

echo "Running unit test for python package"

python3 -m unittest discover
if [ $? != 0 ]; then
    echo "Python unit tests failed"
    exit 1
fi
