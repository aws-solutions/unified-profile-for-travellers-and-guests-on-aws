import { Injectable } from '@angular/core';
import { Configuration } from '../app.constants';
import { AuthContext } from '../model/auth.model';
import { Router } from '@angular/router';
import { LocalStorageService } from './localStorageService';


@Injectable()
export class SessionService {
    public SESSION_KEY: string = "USER_SESSION";
    public DEVICE_TOKEN: string = "DEVICE_TOKEN";

    constructor(private localStorageService: LocalStorageService,
        private router: Router) {

    }

    public create(auth: AuthContext) {
        return this.localStorageService.set(this.SESSION_KEY, auth)

    }

    public setDeviceToken(token: string) {
        return this.localStorageService.set(this.DEVICE_TOKEN, token);
    }
    public getDeviceToken(): string {
        return <string>this.localStorageService.get(this.DEVICE_TOKEN);
    }

    public get(): AuthContext {
        console.log("[SessioinService] Retreiving session stored at ", this.SESSION_KEY)
        return <AuthContext>this.localStorageService.get(this.SESSION_KEY);
    }

    public clear() {
        return this.localStorageService.clear(this.SESSION_KEY);
    }

    setProfileDomain(name: string) {
        this.localStorageService.set("customer-profile-domain", name);
    }

    getProfileDomain(): string {
        return <string>this.localStorageService.get("customer-profile-domain");
    }

    setConnectorData(domain: string, agwUrl: string, tokenEndpoint: string, clientId: string, clientSecret: string, bucketArn: string) {
        this.localStorageService.set("connector-data-" + domain, {
            agwUrl: agwUrl,
            tokenEndpoint: tokenEndpoint,
            clientId: clientId,
            clientSecret: clientSecret,
            bucketArn: bucketArn,
        });
    }

    getConnectorData(domain: string): any {
        return <object>this.localStorageService.get("connector-data-" + domain);
    }
}

