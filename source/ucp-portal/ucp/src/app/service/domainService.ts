import { Injectable } from '@angular/core';
import { UcpService } from '../service/ucpService';
import { Router } from '@angular/router';
import { SessionService } from '../service/sessionService';
import { MatDialog, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { BehaviorSubject, Observable } from 'rxjs';


@Injectable()

export class DomainService {

  //Setting up subscriptions, allows each component using the service
  //to thave their own version of "selectedDomain" and "domains"
  //which subscribe to the version in the service
  //More futureproof than grabbing the value directly from the service (that was my view when implementing)
  public selectedDomain: string = "";
  //todo: create a tyype for this
  public selectedDomainData: any = {};
  public domains = []
  private selectedDomainBS: BehaviorSubject<string>;
  public selectedDomainObs: Observable<string>;

  constructor(private ucpService: UcpService, public dialog: MatDialog, private session: SessionService, private router: Router) {
    this.loadDomains()
    this.selectedDomain = this.session.getProfileDomain()
    this.selectedDomainBS = new BehaviorSubject<string>(this.selectedDomain)
    this.selectedDomainObs = this.selectedDomainBS.asObservable();
  }

  getSelectedDomain() {
    console.log("Get selected domain from session storage")
    return this.session.getProfileDomain()
  }

  loadDomains() {
    console.log("Loading all domains and selecting ")
    return new Promise<any>(resolve => {
      this.getDomains().subscribe((res: any) => {
        //if only one domain is returned. we selected it
        if (res.config.domains.length === 1) {
          this.selectDomain(res.config.domains[0].customerProfileDomain)
        }
        resolve(res.config.domains)
      })
    })

  }

  deleteDomain() {
    console.log("Deleting domain ", this.selectedDomain)
    return new Promise<any>(resolve => {
      if (this.selectedDomain === "") {
        console.log("No domain selected")
        resolve("")
        return
      }
      this.ucpService.deleteDomain(this.selectedDomain).subscribe((res: any) => {
        console.log(res)
        this.session.unsetDomain()
        this.loadDomains()
        this.selectedDomain = ""
        this.selectedDomainData = {}
        this.updateSelectedData("")
        resolve("")
      })
    })

  }

  selectDomain(domain: string) {
    console.log("Selecting domain: ", domain)
    this.session.setProfileDomain(domain)
    this.ucpService.getConfig(domain).subscribe((res: any) => {
      console.log(res)
      this.selectedDomain = domain;
      this.selectedDomainData = res.config.domains[0]
      this.updateSelectedData(domain)
    })
  }

  getDomains() {
    return this.ucpService.listDomains()
  }

  public updateSelectedData(newData: string): void {
    this.selectedDomainBS.next(newData)
  }
}