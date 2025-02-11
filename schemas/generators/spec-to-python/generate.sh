echo "*************************************************"
echo "* AWS Basics for Industry: Python Generator      *"
echo "**************************************************"

echo "0-preping forlders"
mkdir output_dir

echo "1-Setting up environement"
pip install json-schema-codegen

echo "2-Generating Classes"
python generate.py

echo "3-Copying Classes"
cp -r output_dir/* ../../model/python/

echo "4-Cleanup"
rm -rf output_dir

echo "Done"
