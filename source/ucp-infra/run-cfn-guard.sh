echo "running  cfn-guard. make sure cfg guard is installed using: https://docs.aws.amazon.com/cfn-guard/latest/ug/getting-started.html"
curl -tlsv1.3 -sSf https://solutions-build-assets.s3.amazonaws.com/cfn-guard-rules/latest/aws-solutions.guard -o aws-solutions.guard
cat aws-solutions.guard
cfn-guard validate --data ../../deployment/global-s3-assets/ucp.template.json --rules aws-solutions.guard
rm aws-solutions.guard