
env=$1
bucket=$2

echo "Running unit test for python package"

python3 -m unittest discover
source_dir="$(cd $PWD/../../..; pwd -P)"
python3 -m coverage run -m unittest discover
python3 -m coverage xml
cp coverage.xml ../../../tests/coverage-reports/realtime-coverage.coverage.xml
sed -i -e "s,<source>$source_dir,<source>source,g" ../../../tests/coverage-reports/realtime-coverage.coverage.xml
rm .coverage coverage.xml
if [ $? != 0 ]; then
    echo "Python unit tests failed"
    exit 1
fi
