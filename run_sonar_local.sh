echo "***************************************"
echo "*  Run Sonacube Analysis Localy       *"
echo "***************************************"
echo "1-Start Sonarcube server"
docker run -d --restart=always --name sonarqube -p 9000:9000 -p 9092:9092 sonarqube
docker start sonarqube
echo "2-Define IP alias fro local machine to allow container access"
sudo ifconfig lo0 alias 123.123.123.123
echo "3-Run sonnar scanner from Container"
docker run --rm -e SONAR_HOST_URL="http://123.123.123.123:9000" -e SONAR_SCANNER_OPTS="-Dsonar.projectKey=tah-ind-conector" -e SONAR_LOGIN="sqp_ec1760095f18ad8e8cc6b80bada9d2404746ba40" -v "/Users/rollatgr/Documents/work/travel/aws_solutions/industry_connector/industry_connector_solution:/usr/src" sonarsource/sonar-scanner-cli