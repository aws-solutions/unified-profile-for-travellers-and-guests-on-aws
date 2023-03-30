import { Injectable } from '@angular/core';
import { UcpService } from '../service/ucpService';
import { Router } from '@angular/router';
import { SessionService } from '../service/sessionService';
import { MatDialog, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { BehaviorSubject } from 'rxjs';


@Injectable()

export class DomainService {
    public config: any = {};
    public ingestionErrors = [];
    public totalErrors: number = 0;

    public selectedDomain: string = "Select a Domain";
    private selectedDomainBS = new BehaviorSubject<string>(this.selectedDomain)
    public selectedDomainObs = this.selectedDomainBS.asObservable();

    public domains: any[] = [];
    private domainBS = new BehaviorSubject<any[]>(this.domains)
    public domainObs = this.domainBS.asObservable();

    constructor(private ucpService: UcpService, public dialog: MatDialog, private session: SessionService, private router: Router) {
        this.loadDomains()
        this.selectedDomain = this.session.getProfileDomain()
    }

    loadDomains() {
        this.ucpService.listDomains().subscribe((res: any) => {
          console.log(res)
          this.config = res.config;
          this.domains = res.config.domains
          this.updateDomainData(this.domains)
          if (res.config.domains.lenth > 0) {
            this.selectDomain(res.config.domains[0].customerProfileDomain)
          }
        })
      }
    
    selectDomain(domain: string) {
        console.log("Selecting domain: ", domain)
        this.session.setProfileDomain(domain)
        this.ucpService.getConfig(domain).subscribe((res: any) => {
          console.log(res)
          this.config = res.config.domains[0];
          this.selectedDomain = domain;
          this.updateSelectedData(domain)
        })
        this.ucpService.listErrors().subscribe((res: any) => {
          console.log(res)
          this.ingestionErrors = res.ingestionErrors || [];
          this.totalErrors = res.totalErrors || 0;
        })
      }

    getDomains() {
        this.loadDomains()
        console.log("Domain retrival")
        console.log(this.domains)
        this.selectedDomain = this.session.getProfileDomain()
        return this.domains
    }

    public updateDomainData(newData: any[]): void {
        this.domainBS.next(newData)
    }

    public updateSelectedData(newData: string): void {
        this.selectedDomainBS.next(newData)
    }
}