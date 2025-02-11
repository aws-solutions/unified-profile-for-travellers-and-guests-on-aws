#!/bin/sh

# Construct assets/ucp-config.json file from env vars
echo "{\"region\":\"$UCP_REGION\", \"cognitoUserPoolId\":\"$UCP_COGNITO_POOL_ID\", \"cognitoClientId\":\"$UCP_COGNITO_CLIENT_ID\", \"ucpApiUrl\":\"$UCP_API_URL\", \"usePermissionSystem\":\"$USE_PERMISSION_SYSTEM\" }" > /usr/share/nginx/html/assets/ucp-config.json

nginx -g "daemon off;"