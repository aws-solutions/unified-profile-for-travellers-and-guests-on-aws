import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders, HttpParams } from '@angular/common/http';
import { Router } from '@angular/router';
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/catch';
import 'rxjs/add/observable/throw';
import { Observable } from 'rxjs/Observable';
import { Configuration } from '../app.constants';
import { SessionService } from '../service/sessionService';
//import { NotificationsService } from 'angular2-notifications';
import { LoaderService } from './loaderService';
import { RestOptions } from '../model/restOptions';
import { Wrapper } from '../model/wrapper.model';
import { AuthContext } from '../model/auth.model';
import * as _ from 'lodash';
import { NotificationService } from './notificationService';

@Injectable()
export class RestFactory {
    public headers: Headers;
    //public notificationSvc: NotificationsService;
    public loader: LoaderService;
    public http: HttpClient;
    public router: Router;
    public configuration: Configuration;
    public sessionService: SessionService;
    public notif: NotificationService;


    constructor(private _http: HttpClient,
        private _configuration: Configuration,
        private _sessionService: SessionService,
        public _notif: NotificationService,
        private _router: Router,
        private _loader: LoaderService) {
        this.http = _http;
        this.configuration = _configuration;
        //this.notificationSvc = _notificationSvc;
        this.router = _router;
        this.loader = _loader;
        this.notif = _notif;
        this.sessionService = _sessionService;

    }

    public buildRestService(endpoint, options, services) {
        return new RestService(endpoint, options, services, this);
    }
}

const ID_URL_TPL = ":id";
const SEP_URL_TPL = "/";

export class RestService {
    public endpoint: string;
    public options: RestOptions;
    public services: Object;
    public rf: RestFactory;

    constructor(private _enpoint: string,
        private _options: RestOptions,
        private _services: Object,
        private _rf: RestFactory) {

        this.endpoint = _enpoint;
        this.options = _options;
        this.services = _services;
        this.rf = _rf;
    }

    private startLoader(options: RestOptions) {
        if (!options.silent) {
            this.rf.loader.start();
        }
    }

    private stopLoader() {
        this.rf.loader.stop();
    }

    private getActionUrl(endpoint) {
        return this.rf.configuration.buildBackEndUrl(this.options.backend) + endpoint;
    }

    private processError(wrapper: Wrapper) {
        this.stopLoader();
        //this.rf.notificationSvc.error(wrapper.error.code, wrapper.error.msg);
        //user not loged in
        console.log(wrapper)
        //this code is configured in the API gateway for 401
        if ((wrapper.error && wrapper.error.code === "STY_001") || wrapper.status === 401) {
            this.rf.sessionService.clear();
            this.rf.router.navigate(['login']);
        } else if (wrapper.error) {
            if (Array.isArray(wrapper.error.errors) && wrapper.error.errors.length > 0) {
                let msg = ""
                if (wrapper.error.transactionId) {
                    msg += "[tx-" + wrapper.error.transactionId + "]"
                }
                msg += wrapper.error.errors[0].message;
                this.rf.notif.showError(msg)
            } else if (wrapper.error.error && wrapper.error.error.msg) {
                let msg = ""
                if (wrapper.error.transactionId) {
                    msg += "[tx-" + wrapper.error.transactionId + "]"
                }
                msg += wrapper.error.error.msg;
                this.rf.notif.showError(msg)
            } else {
                this.rf.notif.showError("An undefined error happened.")
            }
        } else {
            this.rf.notif.showError("An undefined error happened.")

        }
        return wrapper;
    }

    private isSuccess(status: number): boolean {
        return status == 201 || status == 200;
    }

    private handleResponse(response: any) {
        this.stopLoader();
        return <Wrapper>response;
    }



    public getHeaders(): HttpHeaders {
        let ctx: AuthContext = this.rf.sessionService.get();
        let headers = new HttpHeaders();
        headers = headers.set('Content-Type', 'application/json');
        headers = headers.set('Accept', 'application/json');
        if (!!ctx && !!ctx.token) {
            headers = headers.set('Authorization', ctx.token);
        }
        //temp solution until geoloc api can be properly protected
        let domain = this.rf.sessionService.getProfileDomain()
        if (domain) {
            headers = headers.set('customer-profiles-domain', domain);
        }

        return headers;
    }

    private buildEnpointUrl(endpoint: string, subendpoint: string, withId: boolean): string {
        let components: string[] = [];
        components.push(endpoint)
        if (subendpoint) {
            components.push(subendpoint)
        }
        if (withId) {
            components.push(ID_URL_TPL)
        }
        return components.join(SEP_URL_TPL)
    }




    public get(id: string, params: any, options: RestOptions) {
        params = params || {};
        options = options || new RestOptions();
        params.id = id;
        let enpointWithParams = this.buildEndpoint(this.buildEnpointUrl(this.endpoint, options.subEndpoint, true), params);
        this.startLoader(options);
        return this.rf.http.get(this.getActionUrl(enpointWithParams.endpoint), {
            headers: this.getHeaders(),
            params: this.buildSearchParams(enpointWithParams.unusedParams)
        })
            .map(response => this.handleResponse(response))
            .catch((error: any) => { return Observable.throw(this.processError(<Wrapper>error)) })
    }

    public delete(id: string, params: any, options: RestOptions) {
        params = params || {};
        options = options || new RestOptions();
        params.id = id;
        let enpointWithParams = this.buildEndpoint(this.buildEnpointUrl(this.endpoint, options.subEndpoint, true), params);
        this.startLoader(options);
        return this.rf.http.delete(this.getActionUrl(enpointWithParams.endpoint), {
            headers: this.getHeaders(),
            params: this.buildSearchParams(enpointWithParams.unusedParams)
        })
            .map(response => this.handleResponse(response))
            .catch((error: any) => { return Observable.throw(this.processError(<Wrapper>error)) })
    }

    public query(params: any, options: RestOptions) {
        options = options || new RestOptions();
        let enpointWithParams = this.buildEndpoint(this.buildEnpointUrl(this.endpoint, options.subEndpoint, false), params);
        this.startLoader(options);
        return this.rf.http.get(this.getActionUrl(enpointWithParams.endpoint), {
            headers: this.getHeaders(),
            params: this.buildSearchParams(enpointWithParams.unusedParams)
        })
            .map(response => this.handleResponse(response))
            .catch((error: any) => { return Observable.throw(this.processError(<Wrapper>error)) })
    }


    public post(params: any, body: any, options: RestOptions) {
        options = options || new RestOptions();
        let enpointWithParams = this.buildEndpoint(this.buildEnpointUrl(this.endpoint, options.subEndpoint, false), params);
        this.startLoader(options);
        return this.rf.http.post(this.getActionUrl(enpointWithParams.endpoint), JSON.stringify(body), {
            headers: this.getHeaders(),
            params: this.buildSearchParams(enpointWithParams.unusedParams)
        })
            .map(response => this.handleResponse(response))
            .catch((error: any) => { return Observable.throw(this.processError(<Wrapper>error)) })
    }

    public put(id: string, params: any, body: any, options: RestOptions) {
        params = params || {};
        options = options || new RestOptions();
        params.id = id;
        let enpointWithParams = this.buildEndpoint(this.buildEnpointUrl(this.endpoint, options.subEndpoint, true), params);
        this.startLoader(options);
        return this.rf.http.put(this.getActionUrl(enpointWithParams.endpoint), JSON.stringify(body), {
            headers: this.getHeaders(),
            params: this.buildSearchParams(enpointWithParams.unusedParams)
        })
            .map(response => this.handleResponse(response))
            .catch((error: any) => { return Observable.throw(this.processError(<Wrapper>error)) })
    }

    private buildEndpoint(endpoint: string, params: any) {
        let unusedParams = {};
        _.each(params, function (val, key) {
            var oldEndpoint = endpoint;
            endpoint = endpoint.replace(":" + key, val);
            if (endpoint === oldEndpoint) {
                unusedParams[key] = val
            }
        });
        return { endpoint: endpoint, unusedParams: unusedParams };
    }

    private buildSearchParams(params) {
        let searchParams: HttpParams = new HttpParams();
        _.forOwn(params, function (val, key) {
            if (Array.isArray(val)) {
                if (val.length > 0) {
                    searchParams = searchParams.append(key, val.join(","));
                }
            } else {
                if (val) {
                    searchParams = searchParams.append(key, val);
                }
            }
        });
        return searchParams;
    }
}