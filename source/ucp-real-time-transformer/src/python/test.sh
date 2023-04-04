
env=$1
bucket=$2

echo "Running unit test for python package"

python3 -m unittest discover
python3 -m coverage run -m unittest discover
python3 -m coverage xml
cp coverage.xml ../../../tests/coverage-reports/realtime-coverage.coverage.xml
rm .coverage coverage.xml
if [ $? != 0 ]; then
    echo "Python unit tests failed"
    exit 1
fi
