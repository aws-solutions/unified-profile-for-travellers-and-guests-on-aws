echo "*********************************"
echo "* Glue Schema Generator *"
echo "********************************"

mkdir -p ../../glue_schemas
echo "0-Cleaning directory"
rm -rf ../../glue_schemas/*
echo "1-Compiling typescript"
npm ci
npm run build
echo "2-Generating glue schemas"
node index.js
echo "done"
