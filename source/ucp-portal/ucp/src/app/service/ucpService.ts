import { Injectable } from '@angular/core';
import 'rxjs/add/operator/map'
import { Configuration } from '../app.constants';
import { RestFactory, RestService } from './restFactory';
import { RestOptions } from '../model/restOptions';
import { PaginationOptions } from '../model/pagination.model';

@Injectable()
export class UcpService {

    private service: RestService;

    constructor(private _restFactory: RestFactory, private _config: Configuration) {
        this.service = _restFactory.buildRestService("/api/ucp", { "backend": _config.BACKEND_UCP }, {
        })
    }

    public searchProfiles(searchRq: any) {
        return this.service.query(searchRq, <RestOptions>{ subEndpoint: "profile" });
    }
    public retreiveProfile(id: string, poMap?: Map<string, PaginationOptions>) {

        return this.service.get(id, this.buildMultiPaginationQueryParams(poMap), <RestOptions>{ subEndpoint: "profile" });
    }

    buildMultiPaginationQueryParams(poMap: Map<string, PaginationOptions>) {
        let objects: string[] = []
        let pages: number[] = []
        let pageSizes: number[] = []
        poMap.forEach((value: PaginationOptions, key: string) => {
            objects.push(key)
            pages.push(value.page)
            pageSizes.push(value.pageSize)
        });
        return { objects: objects, pages: pages, pageSizes: pageSizes }
    }

    public getConfig(domain: string) {
        return this.service.get(domain, null, <RestOptions>{ subEndpoint: "admin" });
    }
    public listDomains() {
        return this.service.query({}, <RestOptions>{ subEndpoint: "admin" });
    }
    public getDataValidationErrors(pagination: PaginationOptions) {
        return this.service.query(pagination, <RestOptions>{ subEndpoint: "data" });
    }
    public createDomain(name: string) {
        return this.service.post({}, {
            "domain": {
                "customerProfileDomain": name
            }
        }, <RestOptions>{ subEndpoint: "admin" });
    }
    public deleteDomain(name: string) {
        return this.service.delete(name, {}, <RestOptions>{ subEndpoint: "admin" });
    }
    public listErrors(pagination: PaginationOptions) {
        return this.service.query(pagination, <RestOptions>{ subEndpoint: "error" });
    }
    public getJobs() {
        return this.service.query({}, <RestOptions>{ subEndpoint: "jobs" });
    }
    public deleteError(id) {
        return this.service.delete(id, {}, <RestOptions>{ subEndpoint: "error" });
    }
    public mergeProfile(id: string, id2: string) {
        return this.service.post({}, { mergeRq: { p1: id, p2: id2 } }, <RestOptions>{ subEndpoint: "merge" });
    }
    public listApplications() {
        return this.service.query({}, <RestOptions>{ subEndpoint: "connector" });
    }
    public linkIndustryConnector(agwUrl: string, tokenEndpoint: string, clientId: string, clientSecret: string, bucketArn: string) {
        return this.service.post({}, {
            AgwUrl: agwUrl,
            TokenEndpoint: tokenEndpoint,
            ClientId: clientId,
            ClientSecret: clientSecret,
            BucketArn: bucketArn,
        }, <RestOptions>{ subEndpoint: "connector/link" });
    }
    public createConnectorCrawler(glueRoleArn: string, bucketPath: string, connectorId: string) {
        return this.service.post({}, {
            GlueRoleArn: glueRoleArn,
            BucketPath: bucketPath,
            ConnectorId: connectorId
        }, <RestOptions>{ subEndpoint: "connector/crawler" });
    }
}
