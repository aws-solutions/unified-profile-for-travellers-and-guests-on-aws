import { Component, OnInit, Inject, OnDestroy } from '@angular/core';
import { UcpService } from '../service/ucpService';
import { UserEngagementService } from '../service/userEngagementService';
import { FormGroup, FormControl } from '@angular/forms';
import { Router } from '@angular/router';
import { MatDialog, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { SessionService } from '../service/sessionService';
import { DomainService } from '../service/domainService';
import { faCog, faSearch, faTimes, faPlus } from '@fortawesome/free-solid-svg-icons';
import { Subscription } from 'rxjs';
import { Traveller } from '../model/traveller.model'


@Component({
  selector: 'app-ucp',
  templateUrl: './ucp.component.html',
  styleUrls: ['./ucp.component.css']
})
export class UCPComponent implements OnInit, OnDestroy {
  displayedColumns: string[] = ['id', 'firstName', 'lastName', 'email', 'actions'];
  profiles: Traveller[] = [];
  faCog = faCog;
  faSearch = faSearch;
  faTimes = faTimes;
  faPlus = faPlus;
  searchForm = new FormGroup({
    lastName: new FormControl(null, []),
    email: new FormControl(null, []),
    id: new FormControl(null, []),
    phone: new FormControl(null, []),
    confirmationNumber: new FormControl(null, []),
    loyaltyId: new FormControl(null, []),
  })

  config: any = {};
  ingestionErrors = [];
  totalErrors: number = 0;
  selectedDomain: string;
  private selectDomainSubscription: Subscription

  constructor(private ucpService: UcpService, public dialog: MatDialog, private userEngSvc: UserEngagementService,
    private session: SessionService, private router: Router, public domainService: DomainService) {

  }

  goToSettings() {
    this.router.navigate(["setting"])
  }

  ngOnInit() {
    this.selectDomainSubscription = this.domainService.selectedDomainObs.subscribe((selectedDomain: string) => {
      this.selectedDomain = selectedDomain
    })
  }

  //Unsubscribe from domain subscriptions created in ngoninit, ensure creation / destruction consistancy
  ngOnDestroy() {
    this.selectDomainSubscription.unsubscribe()
  }

  search() {
    let params = <any>{}
    if (this.searchForm.value.id) {
      this.ucpService.retreiveProfile(this.searchForm.value.id).subscribe(res => {
        console.log(res)
        this.profiles = (<any>res).profiles
      })
    } else {
      this.ucpService.searchProfiles(this.searchForm.value).subscribe(res => {
        console.log(res)
        this.profiles = (<any>res).profiles
      })
    }
    this.userEngSvc.logEvent("search_traveler", params)
  }

  viewProfile(profileId) {
    this.router.navigate(["profile/" + profileId])
  }
}






@Component({
  selector: 'delete-confirm',
  templateUrl: './ucp.component-delete-confirmation.html',
  styleUrls: ['./ucp.component.css']
})
export class UCPProfileDeletionConfirmationComponent {
  name: string;
  constructor(public dialogRef: MatDialogRef<UCPProfileDeletionConfirmationComponent>,
    @Inject(MAT_DIALOG_DATA) public data: any) {
    this.name = data.name
  }

  public validate() {
    this.dialogRef.close(true)
  }
  public invalidate() {
    this.dialogRef.close(false)
  }

}

