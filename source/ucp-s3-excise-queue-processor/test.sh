env=$1
bucket=$2
#update this variable to specify the name of your local env
LOCAL_ENV_NAME=dev

export GLUE_SCHEMA_PATH_ASYNC=$(cd ../tah-common/ && pwd)
export TEST_DATA_FOLDER=$(cd ../test_data/ && pwd)

echo "0-setup virtual environment"
python3 -m venv .venv
source .venv/bin/activate
echo "1-downloading package requirements"
pip3 install -r requirements.txt

echo "2-run testing"
mkdir ../z-coverage/ucp-s3-excise-queue-processor
export TAH_REGION=$(aws configure get region)
source_dir="$(cd $PWD/..; pwd -P)"

echo "Running unit test for python package"
python3 -m coverage run -m unittest discover
# tox
rc=$?
python3 -m coverage xml
cp coverage.xml ../z-coverage/tests/coverage-reports/ucp-s3-excise-queue-processor-coverage.coverage.xml
sed -i -e "s,<source>$source_dir,<source>source,g" ../z-coverage/tests/coverage-reports/ucp-s3-excise-queue-processor-coverage.coverage.xml
rm .coverage coverage.xml

deactivate

if [ $rc != 0 ]; then
    echo "Python unit tests failed"
    exit 1
fi
