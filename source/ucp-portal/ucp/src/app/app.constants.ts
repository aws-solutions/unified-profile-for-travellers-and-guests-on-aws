import { Injectable } from '@angular/core';
import infraConfig from './ucp-config.json';

@Injectable()
export class Configuration {
	public BACKEND_UCP = "ucp";
	public BackEnds: any = {
		ucp: {
			server: infraConfig.ucpApiUrl,
			path: ""
		},

	};
	public AWS_CLIENT_ID = infraConfig.cognitoClientId;
	public AWS_USER_POOL_ID = infraConfig.cognitoUserPoolId;
	public AWS_COGNITO_REGION = infraConfig.region;
	public AWS_REGION = infraConfig.region

	public permissions = {
	};

	public buildBackEndUrl(backEnd: string) {
		let beConfig = this.BackEnds[backEnd || this.BACKEND_UCP];
		if (beConfig.port) {
			return beConfig.server + ":" + beConfig.port + "/" + beConfig.path;
		}
		if (beConfig.path) {
			return beConfig.server + "/" + beConfig.path;
		}
		return beConfig.server;
	}
}
