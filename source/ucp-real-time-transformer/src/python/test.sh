
env=$1
bucket=$2

echo "Running unit test for python package"
export TAH_REGION=$(aws configure get region)
source_dir="$(cd $PWD/../../..; pwd -P)"
python3 -m pip install -e ../../../tah_lib --break-system-packages --force-reinstall
rc=$?
if [ $rc != 0 ]; then
    echo "tah_lib install failed"
    exit 1
fi
python3 -m coverage run -m unittest discover
rc=$?
python3 -m coverage xml
cp coverage.xml ../../../z-coverage/tests/coverage-reports/realtime-coverage.coverage.xml
sed -i -e "s,<source>$source_dir,<source>source,g" ../../../z-coverage/tests/coverage-reports/realtime-coverage.coverage.xml
sed -i -e "s,filename=\"$source_dir,filename=\"source,g" ../../../z-coverage/tests/coverage-reports/realtime-coverage.coverage.xml
rm .coverage coverage.xml
if [ $rc != 0 ]; then
    echo "Python unit tests failed"
    exit 1
fi
