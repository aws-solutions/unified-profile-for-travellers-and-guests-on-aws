import { Injectable } from '@angular/core';
import { UcpService } from '../service/ucpService';
import { Router } from '@angular/router';
import { SessionService } from '../service/sessionService';
import { MatDialog, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { DomainCreationModalComponent, UCPProfileDeletionConfirmationComponent } from '../home/ucp.component';
import { BehaviorSubject } from 'rxjs';


@Injectable()

export class DomainService {
    public profiles = [];
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
          this.updateData(this.domains)
          if (res.config.domains.lenth > 0) {
            this.selectDomain(res.config.domains[0].customerProfileDomain)
          }
        })
      }
    
    selectDomain(domain: string) {
        console.log("Selecting domain: ", domain)
        this.profiles = []
        this.session.setProfileDomain(domain)
        //this.showDetail({ "unique_id": "906cd43cae044345b1f2027ad9465fdb" })
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

    createDomain() {
        console.log("Opening dialog to create domain")
        const dialogRef = this.dialog.open(DomainCreationModalComponent, {
          width: '50%',
        });
    
        dialogRef.afterClosed().subscribe((name: any) => {
          console.log('The dialog was closed with value: ', name);
          if (name) {
            this.ucpService.createDomain(name).subscribe((res: any) => {
              console.log(res)
              this.loadDomains()
            })
          }
        });
    }

    getDomains() {
        this.loadDomains()
        console.log("Domain retrival")
        console.log(this.domains)
        this.selectedDomain = this.session.getProfileDomain()
        return this.domains
    }

    deleteDomain(domain: string) {
        const dialogRef = this.dialog.open(UCPProfileDeletionConfirmationComponent, {
          width: '50%',
          data: {
            name: domain
          }
        });
    
        dialogRef.afterClosed().subscribe((confirmed: any) => {
          console.log('The dialog was closed with confirmation: ', confirmed);
          if (confirmed) {
            this.ucpService.deleteDomain(domain).subscribe((res: any) => {
              console.log(res)
              this.session.unsetDomain()
              this.loadDomains()
            })
          }
    
        });
      }

    public updateData(newData: any[]): void {
        this.domainBS.next(newData)
    }

    public updateSelectedData(newData: string): void {
        this.selectedDomainBS.next(newData)
    }
}