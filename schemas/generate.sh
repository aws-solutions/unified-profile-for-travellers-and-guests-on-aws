echo "*************************************************"
echo "* AWS Basics for Industry: Python Generator      *"
echo "**************************************************"

echo "0-Generating Specs"
cd generators/go-to-spec 
sh ./generate.sh

#echo "1-Generate Python"
#cd ../spec-to-python 
#sh ./generate.sh

echo "Done"
