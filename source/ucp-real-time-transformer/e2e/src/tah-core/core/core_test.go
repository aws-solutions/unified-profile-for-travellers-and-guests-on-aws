package core

import (
	"os"
	"strconv"
	"testing"
)

var env = os.Getenv("TAH_CORE_ENV")
var TAH_CORE_REGION = os.Getenv("TAH_CORE_REGION")

/**
* Transactions
**/

func TestLogObfuscation(t *testing.T) {
	tx := NewTransaction("TEST_LOG_OBFUSCATION", "")
	test := `2022/11/06 21:23:42 [tx-2917447081] [ind_connector] [Run] Request from API Gateway: {Resource:/{connector}/{usecase} Path://api/hapi/processBookingFeed HTTPMethod:POST Headers:map[Content-Length:2097 Content-Type:application/json Hapi-Retry-Attempt:0 Host:wnb5esk1q8.execute-api.eu-central-1.amazonaws.com Trace-Id:77907e3e-7ca7-4542-b4cf-c4b614837a94 User-Agent:AHC/2.1 X-Amzn-Trace-Id:Root=1-636825dc-3d27cff370cf6bd11d359083 X-Forwarded-For:52.3.165.169 X-Forwarded-Port:443 X-Forwarded-Proto:https accept:*/* authorization:Bearer eyJraWQiOiJGUXFHMXRRQ2czWFUyYTJYalJCMFhnZkRJbU9ldTFnTnJ3cFpBZ3hhV244PSIsImFsZyI6IlJTMjU2In0.eyJzdWIiOiI0MWNqMHBoNDA3aW04bGRvYmZpZHIxMWRibyIsInRva2VuX3VzZSI6ImFjY2VzcyIsInNjb3BlIjoidXNlcnNcL3Byb2dyYW1tYXRpYyIsImF1dGhfdGltZSI6MTY2Nzc2OTgwMywiaXNzIjoiaHR0cHM6XC9cL2NvZ25pdG8taWRwLmV1LWNlbnRyYWwtMS5hbWF6b25hd3MuY29tXC9ldS1jZW50cmFsLTFfRnhzdUZuYm1vIiwiZXhwIjoxNjY3NzczNDAzLCJpYXQiOjE2Njc3Njk4MDMsInZlcnNpb24iOjIsImp0aSI6ImFlOGYxOTM4LTg4NTMtNDAxMi05MWY5LTMyZDUyNzkyYWVlOCIsImNsaWVudF9pZCI6IjQxY2owcGg0MDdpbThsZG9iZmlkcjExZGJvIn0.l8Mg8NVXAy-YyMo5rnb9RV8pjOU9yG6orG4qh_gY31_lZc2dothACaueIxtzB_6geb7hD7ax61VRTSTmMEti5k9DoVQoK5-ogzGl8LMlr9B2HFpp2TGNE5weZ2AiKd-fG4jWojcfm05W60oQQHpwlPp5R_3n_EVjDMXa--mwp0BSE3SZcnLqqKqECurMaNjOsnqejDrG6yzqEd9NY8jBIdwq-FI4PG96SKaUm0oQCYoDHafm2z-DPpOJ8M36MDmNhvZcgEn-ChFj60_3jwYdr5hB_uBZ9aj8J155orSx-hezKUgijfsgd-IUDUPLxyt5PIolscf5dQ4ZNEWFzxr9SQ duration_ms:PT20.560S] MultiValueHeaders:map[Content-Length:[2097] Content-Type:[application/json] Hapi-Retry-Attempt:[0] Host:[wnb5esk1q8.execute-api.eu-central-1.amazonaws.com] Trace-Id:[77907e3e-7ca7-4542-b4cf-c4b614837a94] User-Agent:[AHC/2.1] X-Amzn-Trace-Id:[Root=1-636825dc-3d27cff370cf6bd11d359083] X-Forwarded-For:[52.3.165.169] X-Forwarded-Port:[443] X-Forwarded-Proto:[https] accept:[*/*] authorization:[Bearer eyJraWQiOiJGUXFHMXRRQ2czWFUyYTJYalJCMFhnZkRJbU9ldTFnTnJ3cFpBZ3hhV244PSIsImFsZyI6IlJTMjU2In0.eyJzdWIiOiI0MWNqMHBoNDA3aW04bGRvYmZpZHIxMWRibyIsInRva2VuX3VzZSI6ImFjY2VzcyIsInNjb3BlIjoidXNlcnNcL3Byb2dyYW1tYXRpYyIsImF1dGhfdGltZSI6MTY2Nzc2OTgwMywiaXNzIjoiaHR0cHM6XC9cL2NvZ25pdG8taWRwLmV1LWNlbnRyYWwtMS5hbWF6b25hd3MuY29tXC9ldS1jZW50cmFsLTFfRnhzdUZuYm1vIiwiZXhwIjoxNjY3NzczNDAzLCJpYXQiOjE2Njc3Njk4MDMsInZlcnNpb24iOjIsImp0aSI6ImFlOGYxOTM4LTg4NTMtNDAxMi05MWY5LTMyZDUyNzkyYWVlOCIsImNsaWVudF9pZCI6IjQxY2owcGg0MDdpbThsZG9iZmlkcjExZGJvIn0.l8Mg8NVXAy-YyMo5rnb9RV8pjOU9yG6orG4qh_gY31_lZc2dothACaueIxtzB_6geb7hD7ax61VRTSTmMEti5k9DoVQoK5-ogzGl8LMlr9B2HFpp2TGNE5weZ2AiKd-fG4jWojcfm05W60oQQHpwlPp5R_3n_EVjDMXa--mwp0BSE3SZcnLqqKqECurMaNjOsnqejDrG6yzqEd9NY8jBIdwq-FI4PG96SKaUm0oQCYoDHafm2z-DPpOJ8M36MDmNhvZcgEn-ChFj60_3jwYdr5hB_uBZ9aj8J155orSx-hezKUgijfsgd-IUDUPLxyt5PIolscf5dQ4ZNEWFzxr9SQ] duration_ms:[PT20.560S]] QueryStringParameters:map[] MultiValueQueryStringParameters:map[] PathParameters:map[connector:hapi usecase:processBookingFeed] StageVariables:map[] RequestContext:{AccountID:9999999999 ResourceID:POST /{connector}/{usecase} OperationName: Stage:api DomainName:wnb5esk1q8.execute-api.eu-central-1.amazonaws.com DomainPrefix:wnb5esk1q8 RequestID:bMraijPZFiAEJxQ= ExtendedRequestID:bMraijPZFiAEJxQ= Protocol:HTTP/1.1 Identity:{CognitoIdentityPoolID: AccountID: CognitoIdentityID: Caller: APIKey: APIKeyID: AccessKey: SourceIP:52.3.165.169 CognitoAuthenticationType: CognitoAuthenticationProvider: UserArn: UserAgent:AHC/2.1 User:} ResourcePath:/{connector}/{usecase} Path:/api/hapi/processBookingFeed Authorizer:map[claims:map[auth_time:1667769803 client_id:41cj0ph407im8ldobfidr11dbo exp:1667773403 iat:1667769803 iss:https://cognito-idp.eu-central-1.amazonaws.com/eu-central-1_FxsuFnbmo jti:ae8f1938-8853-4012-91f9-32d52792aee8 scope:users/programmatic sub:41cj0ph407im8ldobfidr11dbo token_use:access version:2] scopes:<nil>] HTTPMethod:POST RequestTime:06/Nov/2022:21:23:40 +0000 RequestTimeEpoch:1667769820848 APIID:wnb5esk1q8} Body:{"data":{"propertyDetails":{"code":"HHOTEL","chainCode":"HAPI"},"id":"4b9f7a7111534e6a845341804d3873f4","referenceIds":[{"systemId":"HAPI","systemType":"PMS","idType":"RESERVATION_ID","id":"576092"}],"status":"IN_HOUSE","createdDateTime":"2022-09-18T15:31:09","createdBy":"LIANNA","lastModifiedDateTime":"2022-09-18T15:31:09","lastModifiedBy":"LIANNA","arrival":"2022-11-06","departure":"2022-11-09","bookedUnits":[{"start":"2022-11-06","end":"2022-11-09","unitCount":1,"unitTypeCode":"DDBL","unitTypeName":"Deluxe Double"}],"occupiedUnits":[{"start":"2022-11-06","end":"2022-11-09","unitTypeCode":"DDBL","unitTypeName":"Deluxe Double","unitId":"412"}],"occupancyDetails":[{"start":"2022-11-06","end":"2022-11-09","isPerUnit":false,"occupants":[{"type":"ADULT","count":1}]}],"currency":{"code":"USD"},"reservationTotal":{"amountBeforeTax":692.88,"amountAfterTax":741.38},"taxes":[{"code":"All","amount":48.5}],"ratePlans":[{"start":"2022-11-06","end":"2022-11-09","code":"BAR","description":"Best Available Rate"}],"prices":[{"start":"2022-11-06","end":"2022-11-09","ratePlanCode":"BAR","amount":230.96}],"segmentations":[{"start":"2022-11-06","end":"2022-11-09","type":"MARKET","code":"INDLEI","name":"Individual Leisure"}],"guests":[{"isPrimary":true,"guest":{"id":"71adfe9bb26c44b19d2cdd8f3b26d580","type":"GUEST","propertyDetails":{"code":"HHOTEL","chainCode":"HAPI"},"referenceIds":[{"systemId":"HAPI","systemType":"PMS","idType":"GUEST_ID","id":"471112"}],"createdDateTime":"2020-07-30T04:15:03","createdBy":"FRANCISCO","lastModifiedDateTime":"2020-07-30T04:15:03","lastModifiedBy":"KEN","names":[{"givenName":"Antin","surname":"Patron"}],"addresses":[{"type":"Home","line1":"88 Commercial Parkway","postalCode":"22908","city":"Charlottesville","state":{"code":"VA"},"country":{"code":"US"},"isPrimary":false}],"contactDetails":[{"type":"Mobile","value":"314-540-9444","category":"PHONE","isPrimary":true},{"type":"Personal","value":"ksmythin1@bbc.co.uk","category":"EMAIL","isPrimary":true}]}}],"doNotDisplayPrice":false},"receivedDateTime":"2022-11-06T21:23:20.242Z","schemaVersion":"3.0"} IsBase64Encoded:false}`
	expected := `2022/11/06 21:23:42 [tx-2917447081] [ind_connector] [Run] Request from API Gateway: {Resource:/{connector}/{usecase} Path://api/hapi/processBookingFeed HTTPMethod:POST Headers:map[Content-Length:2097 Content-Type:application/json Hapi-Retry-Attempt:0 Host:wnb5esk1q8.execute-api.eu-central-1.amazonaws.com Trace-Id:77907e3e-7ca7-4542-b4cf-c4b614837a94 User-Agent:AHC/2.1 X-Amzn-Trace-Id:Root=1-636825dc-3d27cff370cf6bd11d359083 X-Forwarded-For:52.3.165.169 X-Forwarded-Port:443 X-Forwarded-Proto:https accept:*/* authorization:Bearer eyJraWQiOiJGUXFHMXRRQ2czWFUyYTJYalJCMFhnZkRJbU9ldTFnTnJ3cFpBZ3hhV244PSIsImFsZyI6IlJTMjU2In0.eyJzdWIiOiI0MWNqMHBoNDA3aW04bGRvYmZpZHIxMWRibyIsInRva2VuX3VzZSI6ImFjY2VzcyIsInNjb3BlIjoidXNlcnNcL3Byb2dyYW1tYXRpYyIsImF1dGhfdGltZSI6MTY2Nzc2OTgwMywiaXNzIjoiaHR0cHM6XC9cL2NvZ25pdG8taWRwLmV1LWNlbnRyYWwtMS5hbWF6b25hd3MuY29tXC9ldS1jZW50cmFsLTFfRnhzdUZuYm1vIiwiZXhwIjoxNjY3NzczNDAzLCJpYXQiOjE2Njc3Njk4MDMsInZlcnNpb24iOjIsImp0aSI6ImFlOGYxOTM4LTg4NTMtNDAxMi05MWY5LTMyZDUyNzkyYWVlOCIsImNsaWVudF9pZCI6IjQxY2owcGg0MDdpbThsZG9iZmlkcjExZGJvIn0.l8Mg8NVXAy-YyMo5rnb9RV8pjOU9yG6orG4qh_gY31_lZc2dothACaueIxtzB_6geb7hD7ax61VRTSTmMEti5k9DoVQoK5-ogzGl8LMlr9B2HFpp2TGNE5weZ2AiKd-fG4jWojcfm05W60oQQHpwlPp5R_3n_EVjDMXa--mwp0BSE3SZcnLqqKqECurMaNjOsnqejDrG6yzqEd9NY8jBIdwq-FI4PG96SKaUm0oQCYoDHafm2z-DPpOJ8M36MDmNhvZcgEn-ChFj60_3jwYdr5hB_uBZ9aj8J155orSx-hezKUgijfsgd-IUDUPLxyt5PIolscf5dQ4ZNEWFzxr9SQ duration_ms:PT20.560S] MultiValueHeaders:map[Content-Length:[2097] Content-Type:[application/json] Hapi-Retry-Attempt:[0] Host:[wnb5esk1q8.execute-api.eu-central-1.amazonaws.com] Trace-Id:[77907e3e-7ca7-4542-b4cf-c4b614837a94] User-Agent:[AHC/2.1] X-Amzn-Trace-Id:[Root=1-636825dc-3d27cff370cf6bd11d359083] X-Forwarded-For:[52.3.165.169] X-Forwarded-Port:[443] X-Forwarded-Proto:[https] accept:[*/*] authorization:[Bearer eyJraWQiOiJGUXFHMXRRQ2czWFUyYTJYalJCMFhnZkRJbU9ldTFnTnJ3cFpBZ3hhV244PSIsImFsZyI6IlJTMjU2In0.eyJzdWIiOiI0MWNqMHBoNDA3aW04bGRvYmZpZHIxMWRibyIsInRva2VuX3VzZSI6ImFjY2VzcyIsInNjb3BlIjoidXNlcnNcL3Byb2dyYW1tYXRpYyIsImF1dGhfdGltZSI6MTY2Nzc2OTgwMywiaXNzIjoiaHR0cHM6XC9cL2NvZ25pdG8taWRwLmV1LWNlbnRyYWwtMS5hbWF6b25hd3MuY29tXC9ldS1jZW50cmFsLTFfRnhzdUZuYm1vIiwiZXhwIjoxNjY3NzczNDAzLCJpYXQiOjE2Njc3Njk4MDMsInZlcnNpb24iOjIsImp0aSI6ImFlOGYxOTM4LTg4NTMtNDAxMi05MWY5LTMyZDUyNzkyYWVlOCIsImNsaWVudF9pZCI6IjQxY2owcGg0MDdpbThsZG9iZmlkcjExZGJvIn0.l8Mg8NVXAy-YyMo5rnb9RV8pjOU9yG6orG4qh_gY31_lZc2dothACaueIxtzB_6geb7hD7ax61VRTSTmMEti5k9DoVQoK5-ogzGl8LMlr9B2HFpp2TGNE5weZ2AiKd-fG4jWojcfm05W60oQQHpwlPp5R_3n_EVjDMXa--mwp0BSE3SZcnLqqKqECurMaNjOsnqejDrG6yzqEd9NY8jBIdwq-FI4PG96SKaUm0oQCYoDHafm2z-DPpOJ8M36MDmNhvZcgEn-ChFj60_3jwYdr5hB_uBZ9aj8J155orSx-hezKUgijfsgd-IUDUPLxyt5PIolscf5dQ4ZNEWFzxr9SQ] duration_ms:[PT20.560S]] QueryStringParameters:map[] MultiValueQueryStringParameters:map[] PathParameters:map[connector:hapi usecase:processBookingFeed] StageVariables:map[] RequestContext:{AccountID:9999999999 ResourceID:POST /{connector}/{usecase} OperationName: Stage:api DomainName:wnb5esk1q8.execute-api.eu-central-1.amazonaws.com DomainPrefix:wnb5esk1q8 RequestID:bMraijPZFiAEJxQ= ExtendedRequestID:bMraijPZFiAEJxQ= Protocol:HTTP/1.1 Identity:{CognitoIdentityPoolID: AccountID: CognitoIdentityID: Caller: APIKey: APIKeyID: AccessKey: SourceIP:52.3.165.169 CognitoAuthenticationType: CognitoAuthenticationProvider: UserArn: UserAgent:AHC/2.1 User:} ResourcePath:/{connector}/{usecase} Path:/api/hapi/processBookingFeed Authorizer:map[claims:map[auth_time:1667769803 client_id:41cj0ph407im8ldobfidr11dbo exp:1667773403 iat:1667769803 iss:https://cognito-idp.eu-central-1.amazonaws.com/eu-central-1_FxsuFnbmo jti:ae8f1938-8853-4012-91f9-32d52792aee8 scope:users/programmatic sub:41cj0ph407im8ldobfidr11dbo token_use:access version:2] scopes:<nil>] HTTPMethod:POST RequestTime:06/Nov/2022:21:23:40 +0000 RequestTimeEpoch:1667769820848 APIID:wnb5esk1q8} Body:******************** IsBase64Encoded:false}`

	tx.AddLogObfuscationPattern("Body:", " *{(.|\n)*}", " ")
	obfuscated := tx.Obfuscate(test)
	if obfuscated != expected {
		t.Errorf("[TestLogObfuscation] log oobfuscation failed: %v\n should be %v", obfuscated, expected)
	}
	obfuscated = tx.Obfuscate(LOG_LINE_WITH_LINE_BREAK)
	expected = `[tx-2733684515] [ind_connector] [Run] Request from API Gateway: {Resource:/{connector}/{usecase} Path://api/hapi/processBookingFeed HTTPMethod:POST Headers:map[Content-Length:6249 Content-Type:application/json Host:wnb5esk1q8.execute-api.eu-central-1.amazonaws.com Postman-Token:60660ba3-e76f-47c3-a7c5-3a662081b075 User-Agent:PostmanRuntime/7.29.2 X-Amzn-Trace-Id:Root=1-63691ad1-4cc12c1f22360d7d72754005 X-Forwarded-For:24.62.100.140 X-Forwarded-Port:443 X-Forwarded-Proto:https accept:*/* accept-encoding:gzip, deflate, br authorization:Bearer eyJraWQiOiJGUXFHMXRRQ2czWFUyYTJYalJCMFhnZkRJbU9ldTFnTnJ3cFpBZ3hhV244PSIsImFsZyI6IlJTMjU2In0.eyJzdWIiOiI0MWNqMHBoNDA3aW04bGRvYmZpZHIxMWRibyIsInRva2VuX3VzZSI6ImFjY2VzcyIsInNjb3BlIjoidXNlcnNcL3Byb2dyYW1tYXRpYyIsImF1dGhfdGltZSI6MTY2NzgzMjUyMCwiaXNzIjoiaHR0cHM6XC9cL2NvZ25pdG8taWRwLmV1LWNlbnRyYWwtMS5hbWF6b25hd3MuY29tXC9ldS1jZW50cmFsLTFfRnhzdUZuYm1vIiwiZXhwIjoxNjY3ODM2MTIwLCJpYXQiOjE2Njc4MzI1MjAsInZlcnNpb24iOjIsImp0aSI6IjkyYTNlMTIyLTU2ZWItNGM0Zi1hOTc3LWQzMmNmYWExYjJmZCIsImNsaWVudF9pZCI6IjQxY2owcGg0MDdpbThsZG9iZmlkcjExZGJvIn0.c1_QNasvSRzMpmyoCV4yYxFnJHh9H0z9WlAOLRfqC9GyFvw7zOGcejM77QK5LvrlZQgMyfAlscdFOlGFvhHBfBzQCxli96JDcE8Z67I35kWpSXsejJ3W2pKTi97cOayAPC5Yv591yklhn3cRCFHXcEzgziTZETS60FdCHrw0saGjD7H588Se7079NAudmti2zBOcDmuD5jsuwqmj_l3NUSlfTKidbJu9rzGYy-arOhqJSQkTCG7BSlJavOSjA43Co9taqZ8W9mcUaGuEcBEKSYJPS7E8cEzfQ7yGgaoiHLv1anKe7D006BeUbAiO2Y8jdbmQpQQNPI32HyZMkj2fqQ cache-control:no-cache] MultiValueHeaders:map[Content-Length:[6249] Content-Type:[application/json] Host:[wnb5esk1q8.execute-api.eu-central-1.amazonaws.com] Postman-Token:[60660ba3-e76f-47c3-a7c5-3a662081b075] User-Agent:[PostmanRuntime/7.29.2] X-Amzn-Trace-Id:[Root=1-63691ad1-4cc12c1f22360d7d72754005] X-Forwarded-For:[24.62.100.140] X-Forwarded-Port:[443] X-Forwarded-Proto:[https] accept:[*/*] accept-encoding:[gzip, deflate, br] authorization:[Bearer eyJraWQiOiJGUXFHMXRRQ2czWFUyYTJYalJCMFhnZkRJbU9ldTFnTnJ3cFpBZ3hhV244PSIsImFsZyI6IlJTMjU2In0.eyJzdWIiOiI0MWNqMHBoNDA3aW04bGRvYmZpZHIxMWRibyIsInRva2VuX3VzZSI6ImFjY2VzcyIsInNjb3BlIjoidXNlcnNcL3Byb2dyYW1tYXRpYyIsImF1dGhfdGltZSI6MTY2NzgzMjUyMCwiaXNzIjoiaHR0cHM6XC9cL2NvZ25pdG8taWRwLmV1LWNlbnRyYWwtMS5hbWF6b25hd3MuY29tXC9ldS1jZW50cmFsLTFfRnhzdUZuYm1vIiwiZXhwIjoxNjY3ODM2MTIwLCJpYXQiOjE2Njc4MzI1MjAsInZlcnNpb24iOjIsImp0aSI6IjkyYTNlMTIyLTU2ZWItNGM0Zi1hOTc3LWQzMmNmYWExYjJmZCIsImNsaWVudF9pZCI6IjQxY2owcGg0MDdpbThsZG9iZmlkcjExZGJvIn0.c1_QNasvSRzMpmyoCV4yYxFnJHh9H0z9WlAOLRfqC9GyFvw7zOGcejM77QK5LvrlZQgMyfAlscdFOlGFvhHBfBzQCxli96JDcE8Z67I35kWpSXsejJ3W2pKTi97cOayAPC5Yv591yklhn3cRCFHXcEzgziTZETS60FdCHrw0saGjD7H588Se7079NAudmti2zBOcDmuD5jsuwqmj_l3NUSlfTKidbJu9rzGYy-arOhqJSQkTCG7BSlJavOSjA43Co9taqZ8W9mcUaGuEcBEKSYJPS7E8cEzfQ7yGgaoiHLv1anKe7D006BeUbAiO2Y8jdbmQpQQNPI32HyZMkj2fqQ] cache-control:[no-cache]] QueryStringParameters:map[] MultiValueQueryStringParameters:map[] PathParameters:map[connector:hapi usecase:processBookingFeed] StageVariables:map[] RequestContext:{AccountID:9999999999 ResourceID:POST /{connector}/{usecase} OperationName: Stage:api DomainName:wnb5esk1q8.execute-api.eu-central-1.amazonaws.com DomainPrefix:wnb5esk1q8 RequestID:bPEgyhxRFiAEMsg= ExtendedRequestID:bPEgyhxRFiAEMsg= Protocol:HTTP/1.1 Identity:{CognitoIdentityPoolID: AccountID: CognitoIdentityID: Caller: APIKey: APIKeyID: AccessKey: SourceIP:24.62.100.140 CognitoAuthenticationType: CognitoAuthenticationProvider: UserArn: UserAgent:PostmanRuntime/7.29.2 User:} ResourcePath:/{connector}/{usecase} Path:/api/hapi/processBookingFeed Authorizer:map[claims:map[auth_time:1667832520 client_id:41cj0ph407im8ldobfidr11dbo exp:1667836120 iat:1667832520 iss:https://cognito-idp.eu-central-1.amazonaws.com/eu-central-1_FxsuFnbmo jti:92a3e122-56eb-4c4f-a977-d32cfaa1b2fd scope:users/programmatic sub:41cj0ph407im8ldobfidr11dbo token_use:access version:2] scopes:<nil>] HTTPMethod:POST RequestTime:07/Nov/2022:14:48:49 +0000 RequestTimeEpoch:1667832529604 APIID:wnb5esk1q8} Body:******************** IsBase64Encoded:false}`
	if obfuscated != expected {
		t.Errorf("[TestLogObfuscation] log oobfuscation failed: %v\n should be %v", obfuscated, expected)
	}

}

/*******
* Geohash
*****/

func TestGeohash(t *testing.T) {
	lat := 37.3526713
	lng := -121.9803301
	geohash := GeoHash(lat, lng)
	geohash001 := GeoHashWithPrecision(lat, lng, 0.01) //1km
	geohash1 := GeoHashWithPrecision(lat, lng, 1)      //1km
	geohash10 := GeoHashWithPrecision(lat, lng, 10)    //10km
	geohash100 := GeoHashWithPrecision(lat, lng, 10)   //10km

	if geohash != "9q9k2rbeh69e" {
		t.Errorf("[Geohash] Invalid heohash %+v for ccords (%v,%v)", geohash, lat, lng)
	}
	if geohash001 != "9q9k2r" {
		t.Errorf("[Geohash] Invalid heohash %+v for ccords (%v,%v) with prec 0.01km", geohash001, lat, lng)
	}
	if geohash1 != "9q9k2" {
		t.Errorf("[Geohash] Invalid heohash %+v for ccords (%v,%v) with prec 1km", geohash1, lat, lng)
	}
	if geohash10 != "9q9k" {
		t.Errorf("[Geohash] Invalid heohash %+v for ccords (%v,%v) with prec 10km", geohash10, lat, lng)
	}
	if geohash100 != "9q9k" {
		t.Errorf("[Geohash] Invalid heohash %+v for ccords (%v,%v)with prec 100km", geohash100, lat, lng)
	}
}

/*******
* UTILS
*****/

func TestChunkArray(t *testing.T) {
	chunked := ChunkArray(10, 3)
	lenCheck := len(chunked[0]) == 3 && len(chunked[1]) == 3 && len(chunked[2]) == 3 && len(chunked[3]) == 1
	if !lenCheck {
		t.Errorf("[ChunkArray] Invalid size for chunked array: %+v", chunked)
	}
}

/*******
* CORE Model
*****/
func TestErrorManagement(t *testing.T) {
	respError := ValidationError(93, "test error")
	if respError.Code != VALIDATION_ERROR_PREFIX+"_"+strconv.Itoa(93) {
		t.Errorf("[ErrorManagement] Error while generating ersponse error: %+v ", respError)
	}
}

func TestMapCode(t *testing.T) {
	errString := "VALIDATION_ERROR_123-this is a validation error with -1 as a negative numebr"
	code := MapCode(errString)
	if code != "VALIDATION_ERROR_123" {
		t.Errorf("[ErrorManagement] Error building ResError from '%v': invalid error code: '%v' ", errString, code)
	}
}

var LOG_LINE_WITH_LINE_BREAK = `[tx-2733684515] [ind_connector] [Run] Request from API Gateway: {Resource:/{connector}/{usecase} Path://api/hapi/processBookingFeed HTTPMethod:POST Headers:map[Content-Length:6249 Content-Type:application/json Host:wnb5esk1q8.execute-api.eu-central-1.amazonaws.com Postman-Token:60660ba3-e76f-47c3-a7c5-3a662081b075 User-Agent:PostmanRuntime/7.29.2 X-Amzn-Trace-Id:Root=1-63691ad1-4cc12c1f22360d7d72754005 X-Forwarded-For:24.62.100.140 X-Forwarded-Port:443 X-Forwarded-Proto:https accept:*/* accept-encoding:gzip, deflate, br authorization:Bearer eyJraWQiOiJGUXFHMXRRQ2czWFUyYTJYalJCMFhnZkRJbU9ldTFnTnJ3cFpBZ3hhV244PSIsImFsZyI6IlJTMjU2In0.eyJzdWIiOiI0MWNqMHBoNDA3aW04bGRvYmZpZHIxMWRibyIsInRva2VuX3VzZSI6ImFjY2VzcyIsInNjb3BlIjoidXNlcnNcL3Byb2dyYW1tYXRpYyIsImF1dGhfdGltZSI6MTY2NzgzMjUyMCwiaXNzIjoiaHR0cHM6XC9cL2NvZ25pdG8taWRwLmV1LWNlbnRyYWwtMS5hbWF6b25hd3MuY29tXC9ldS1jZW50cmFsLTFfRnhzdUZuYm1vIiwiZXhwIjoxNjY3ODM2MTIwLCJpYXQiOjE2Njc4MzI1MjAsInZlcnNpb24iOjIsImp0aSI6IjkyYTNlMTIyLTU2ZWItNGM0Zi1hOTc3LWQzMmNmYWExYjJmZCIsImNsaWVudF9pZCI6IjQxY2owcGg0MDdpbThsZG9iZmlkcjExZGJvIn0.c1_QNasvSRzMpmyoCV4yYxFnJHh9H0z9WlAOLRfqC9GyFvw7zOGcejM77QK5LvrlZQgMyfAlscdFOlGFvhHBfBzQCxli96JDcE8Z67I35kWpSXsejJ3W2pKTi97cOayAPC5Yv591yklhn3cRCFHXcEzgziTZETS60FdCHrw0saGjD7H588Se7079NAudmti2zBOcDmuD5jsuwqmj_l3NUSlfTKidbJu9rzGYy-arOhqJSQkTCG7BSlJavOSjA43Co9taqZ8W9mcUaGuEcBEKSYJPS7E8cEzfQ7yGgaoiHLv1anKe7D006BeUbAiO2Y8jdbmQpQQNPI32HyZMkj2fqQ cache-control:no-cache] MultiValueHeaders:map[Content-Length:[6249] Content-Type:[application/json] Host:[wnb5esk1q8.execute-api.eu-central-1.amazonaws.com] Postman-Token:[60660ba3-e76f-47c3-a7c5-3a662081b075] User-Agent:[PostmanRuntime/7.29.2] X-Amzn-Trace-Id:[Root=1-63691ad1-4cc12c1f22360d7d72754005] X-Forwarded-For:[24.62.100.140] X-Forwarded-Port:[443] X-Forwarded-Proto:[https] accept:[*/*] accept-encoding:[gzip, deflate, br] authorization:[Bearer eyJraWQiOiJGUXFHMXRRQ2czWFUyYTJYalJCMFhnZkRJbU9ldTFnTnJ3cFpBZ3hhV244PSIsImFsZyI6IlJTMjU2In0.eyJzdWIiOiI0MWNqMHBoNDA3aW04bGRvYmZpZHIxMWRibyIsInRva2VuX3VzZSI6ImFjY2VzcyIsInNjb3BlIjoidXNlcnNcL3Byb2dyYW1tYXRpYyIsImF1dGhfdGltZSI6MTY2NzgzMjUyMCwiaXNzIjoiaHR0cHM6XC9cL2NvZ25pdG8taWRwLmV1LWNlbnRyYWwtMS5hbWF6b25hd3MuY29tXC9ldS1jZW50cmFsLTFfRnhzdUZuYm1vIiwiZXhwIjoxNjY3ODM2MTIwLCJpYXQiOjE2Njc4MzI1MjAsInZlcnNpb24iOjIsImp0aSI6IjkyYTNlMTIyLTU2ZWItNGM0Zi1hOTc3LWQzMmNmYWExYjJmZCIsImNsaWVudF9pZCI6IjQxY2owcGg0MDdpbThsZG9iZmlkcjExZGJvIn0.c1_QNasvSRzMpmyoCV4yYxFnJHh9H0z9WlAOLRfqC9GyFvw7zOGcejM77QK5LvrlZQgMyfAlscdFOlGFvhHBfBzQCxli96JDcE8Z67I35kWpSXsejJ3W2pKTi97cOayAPC5Yv591yklhn3cRCFHXcEzgziTZETS60FdCHrw0saGjD7H588Se7079NAudmti2zBOcDmuD5jsuwqmj_l3NUSlfTKidbJu9rzGYy-arOhqJSQkTCG7BSlJavOSjA43Co9taqZ8W9mcUaGuEcBEKSYJPS7E8cEzfQ7yGgaoiHLv1anKe7D006BeUbAiO2Y8jdbmQpQQNPI32HyZMkj2fqQ] cache-control:[no-cache]] QueryStringParameters:map[] MultiValueQueryStringParameters:map[] PathParameters:map[connector:hapi usecase:processBookingFeed] StageVariables:map[] RequestContext:{AccountID:9999999999 ResourceID:POST /{connector}/{usecase} OperationName: Stage:api DomainName:wnb5esk1q8.execute-api.eu-central-1.amazonaws.com DomainPrefix:wnb5esk1q8 RequestID:bPEgyhxRFiAEMsg= ExtendedRequestID:bPEgyhxRFiAEMsg= Protocol:HTTP/1.1 Identity:{CognitoIdentityPoolID: AccountID: CognitoIdentityID: Caller: APIKey: APIKeyID: AccessKey: SourceIP:24.62.100.140 CognitoAuthenticationType: CognitoAuthenticationProvider: UserArn: UserAgent:PostmanRuntime/7.29.2 User:} ResourcePath:/{connector}/{usecase} Path:/api/hapi/processBookingFeed Authorizer:map[claims:map[auth_time:1667832520 client_id:41cj0ph407im8ldobfidr11dbo exp:1667836120 iat:1667832520 iss:https://cognito-idp.eu-central-1.amazonaws.com/eu-central-1_FxsuFnbmo jti:92a3e122-56eb-4c4f-a977-d32cfaa1b2fd scope:users/programmatic sub:41cj0ph407im8ldobfidr11dbo token_use:access version:2] scopes:<nil>] HTTPMethod:POST RequestTime:07/Nov/2022:14:48:49 +0000 RequestTimeEpoch:1667832529604 APIID:wnb5esk1q8} Body:        {
"data": {
"propertyDetails": {
"code": "HHOTEL",
"chainCode": "HAPI"
},
"id": "f514cd52c85745fd910897f88ca9e057",
"referenceIds": [
{
"systemId": "HAPI",
"systemType": "PMS",
"idType": "RESERVATION_ID",
"id": "504430"
}
],
"status": "IN_HOUSE",
"createdDateTime": "2021-12-20T09:56:29",
"createdBy": "ELIOT",
"lastModifiedDateTime": "2021-12-20T09:56:29",
"lastModifiedBy": "ELIOT",
"arrival": "2022-01-20",
"departure": "2022-01-24",
"bookedUnits": [
{
"start": "2022-01-20",
"end": "2022-01-24",
"unitCount": 1,
"unitTypeCode": "DKNG",
"unitTypeName": "Deluxe King"
}
],
"occupiedUnits": [
{
"start": "2022-01-20",
"end": "2022-01-24",
"unitTypeCode": "DKNG",
"unitTypeName": "Deluxe King",
"unitId": "470"
}
],
"occupancyDetails": [
{
"start": "2022-01-20",
"end": "2022-01-24",
"isPerUnit": false,
"occupants": [
{
"type": "ADULT",
"count": 1
}
]
}
],
"currency": {
"code": "USD"
},
"reservationTotal": {
"amountBeforeTax": 504.32,
"amountAfterTax": 539.62
},
"taxes": [
{
"code": "All",
"amount": 35.3
}
],
"ratePlans": [
{
"start": "2022-01-20",
"end": "2022-01-24",
"code": "BAR",
"description": "Best Available Rate"
}
],
"prices": [
{
"start": "2022-01-20",
"end": "2022-01-24",
"ratePlanCode": "BAR",
"amount": 126.08
}
],
"segmentations": [
{
"start": "2022-01-20",
"end": "2022-01-24",
"type": "MARKET",
"code": "INDLEI",
"name": "Individual Leisure"
}
],
"guests": [
{
"isPrimary": true,
"guest": {
"id": "1b6a2e486c784c11b1822bcbd2881911",
"type": "GUEST",
"propertyDetails": {
"code": "HHOTEL",
"chainCode": "HAPI"
},
"referenceIds": [
{
"systemId": "HAPI",
"systemType": "PMS",
"idType": "GUEST_ID",
"id": "841589"
}
],
"createdDateTime": "2020-07-24T10:15:21",
"createdBy": "MICHELE",
"lastModifiedDateTime": "2020-07-24T10:15:21",
"lastModifiedBy": "BARTHOLEMY",
"names": [
{
"givenName": "Hannah",
"surname": "Jurasz"
}
],
"addresses": [
{
"type": "Home",
"line1": "771 Independence Terrace",
"postalCode": "46805",
"city": "Fort Wayne",
"state": {
"code": "IN"
},
"country": {
"code": "US"
},
"isPrimary": true
}
],
"contactDetails": [
{
"type": "Mobile",
"value": "992-240-4709",
"category": "PHONE",
"isPrimary": true
},
{
"type": "Personal",
"value": "bshillum0@pagesperso-orange.fr",
"category": "EMAIL",
"isPrimary": true
}
]
}
}
],
"doNotDisplayPrice": false
},
"receivedDateTime": "2022-01-20T13:24:58.182Z",
"createdDateTime": "2022-01-17T15:01:41.037Z",
"lastModifiedDateTime": "2022-10-06T22:58:05.166Z"
} IsBase64Encoded:false}`
