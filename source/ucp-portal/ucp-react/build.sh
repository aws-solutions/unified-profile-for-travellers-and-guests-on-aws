env=$1
bucket=$2

echo "**********************************************"
echo "* Build Adminitration Portal '$env' "
echo "***********************************************"
if [ -z "$env" ] || [ -z "$bucket" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh deploy.sh <env> <bucket>"
    exit 1
fi

if [ "$3" = "--skip-tests" ]; then
    echo "Skipping UI tests"
else
    mkdir coverage
    rm -r ../../z-coverage/ucp-portal/ucp-react/*
    npm run test
    rc=$?
    if [ $rc -ne 0 ]; then
        echo "Front end build Failed! Exiting with status $rc" >&2
        exit $rc
    fi
    mkdir -p ../../z-coverage/ucp-portal/ucp-react
    cp -a coverage/. ../../z-coverage/ucp-portal/ucp-react
    rm -rf coverage
fi

echo "1-Build application package for env $env"
rm -rf dist
npx prettier --check './**/*.{ts,tsx}' || echo Problems found
npx eslint --max-warnings=0 . || echo Problems found
npm run build
rc=$?
if [ $rc -ne 0 ]; then
    echo "Build failed with status $rc" >&2
    exit $rc
fi
echo "2-Copying static files to deployment/regional-s3-assets/ui/"
pwd
echo "Deployment folder content:"
ls ../../../deployment/
echo "2.1 created ui directory if not exists"
mkdir -p ../../../deployment/regional-s3-assets
mkdir -p ../../../deployment/regional-s3-assets/ui
rm ../../../deployment/regional-s3-assets/ui/dist.zip
echo "2.3 Copying assets to ui folder"
cd dist
zip -r dist.zip *
cp dist.zip ../../../../deployment/regional-s3-assets/ui/dist.zip
ls ../../../../deployment/regional-s3-assets/ui/
exit 0
