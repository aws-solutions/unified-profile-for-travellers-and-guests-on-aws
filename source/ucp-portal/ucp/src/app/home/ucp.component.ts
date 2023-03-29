import { Component, OnInit, Inject, OnDestroy } from '@angular/core';
import { UcpService } from '../service/ucpService';
import { UserEngagementService } from '../service/userEngagementService';
import { FormGroup, FormArray, FormControl, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { MatDialog, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { SessionService } from '../service/sessionService';
import { DomainService } from '../service/domainService';
import { faCog, faSearch, faTimes, faPlus } from '@fortawesome/free-solid-svg-icons';
import { Subscription } from 'rxjs';


import * as moment from 'moment';


@Component({
  selector: 'app-ucp',
  templateUrl: './ucp.component.html',
  styleUrls: ['./ucp.component.css']
})
export class UCPComponent implements OnInit, OnDestroy {
  displayedColumns: string[] = ['id', 'firstName', 'lastName', 'email', 'actions'];
  profiles = [];
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
  domains: any[] = [];
  private domainSubscription: Subscription
  private selectDomainSubscription: Subscription

  constructor(private ucpService: UcpService, public dialog: MatDialog, private userEngSvc: UserEngagementService, 
    private session: SessionService, private router: Router, private domainService: DomainService) {
    this.loadDomains()

  }

  loadDomains() {
    this.domainService.loadDomains()
  }

  selectDomain(domain: string) {
    this.domainService.selectDomain(domain)
  }

  createDomain() {
    this.domainService.createDomain()
  }

  goToSettings() {
    this.router.navigate(["setting"])
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


  ngOnInit() {
    this.domainSubscription = this.domainService.domainObs.subscribe((domains: any[]) => {
      this.domains = domains
    })

    this.selectDomainSubscription = this.domainService.selectedDomainObs.subscribe((selectedDomain: string) => {
      this.selectedDomain = selectedDomain
    })
  }

  ngOnDestroy() {
    this.domainSubscription.unsubscribe()
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

  showDetail(guest): void {
    const dialogRef = this.dialog.open(UCPDetailComponent, {
      height: '80%',
      width: '99%',
      data: {
        id: guest.unique_id
      }
    });

    dialogRef.afterClosed().subscribe(result => {
      console.log('The dialog was closed');
    });
  }


  showErrors(): void {
    const dialogRef = this.dialog.open(UCPProfileErrorsComponent, {
      height: '80%',
      width: '99%',
      data: {
        errors: this.ingestionErrors,
        totalErrors: this.totalErrors
      }
    });

    dialogRef.afterClosed().subscribe(result => {
      console.log('The dialog was closed');
    });
  }





}


@Component({
  selector: 'ucp-detail',
  templateUrl: 'ucp.component-detail.html',
  styleUrls: ['./ucp.component.css']
})
export class UCPDetailComponent {
  guest: any = {};
  matches: any = []
  propertyMap: any = {};
  bookStart = 0;
  bookPageSize = 10
  bookEnd = this.bookPageSize;
  totalRevenueGenertated = 0;
  constructor(
    private ucpService: UcpService,
    public dialog: MatDialog,
    @Inject(MAT_DIALOG_DATA) public data: any) {
    this.retreive(data.id)

  }

  showSideBySide(matchId: string, matchScore: number) {
    const dialogRef = this.dialog.open(UCPCompareComponent, {
      height: '80%',
      width: '99%',
      data: {
        id1: this.guest.unique_id,
        id2: matchId,
        matchScore: matchScore
      }
    });

    dialogRef.afterClosed().subscribe(result => {
      console.log('The dialog was closed');
    });
  }


  ngOnInit() {

  }

  toPropertyName(code: string): string {
    return (this.propertyMap[code] || {}).name || code
  }

  next() {
    this.bookStart += this.bookPageSize
    this.bookEnd += this.bookPageSize
  }
  prev() {
    this.bookStart -= this.bookPageSize
    this.bookEnd -= this.bookPageSize
  }

  retreive(id) {
    console.log(id)
    if (id) {
      this.ucpService.retreiveProfile(id).subscribe(response => this.afterRetreive(response),
        () => { });
    } else {
      console.log("no ID for guest 360 retreive")
    }
  }


  afterRetreive(response) {
    console.log(response)
    this.matches = response.matches
    this.guest = response.profiles[0]
    /*  (this.guest.hotelBookings || []).forEach((booking) => {
        booking.startDate = moment(booking.startDate).format("YYYY-MM-DD")
        this.totalRevenueGenertated += +booking.total_price
      });
    (this.guest.hotelBookings || []).sort((b1, b2) => {
      if (moment(b1.startDate).isAfter(moment(b2.startDate))) {
        return -1
      }
      return 1
    })*/
    let hotelSearches = {}
    let locationSearches = {}
    this.guest.searches.forEach((search) => {
      if (search.location) {
        locationSearches[search.date + search.location] = {
          "date": moment(search.date).format("MMM-YYYY"), "location": search.location
        }
      }
      if (search.hotel) {
        hotelSearches[search.date + search.hotel] = {
          "date": moment(search.date).format("MMM-YYYY"), "hotel": search.hotel
        }
      }
    });
    this.guest.hotelSearches = Object.values(hotelSearches)
    this.guest.locationSearches = Object.values(locationSearches)
  }


}


@Component({
  selector: 'ucp-compare',
  templateUrl: 'ucp.component-compare.html',
  styleUrls: ['./ucp.component.css']
})
export class UCPCompareComponent {
  guest1: any = {};
  guest2: any = {};
  matchScore: number = 0;
  loyaltyDiff: any = {};
  loyaltyArray = [];
  public personalData: FormGroup = new FormGroup({
    lastName: new FormControl("doNothing"),
    firstName: new FormControl("doNothing"),
    middleName: new FormControl("doNothing"),
    pronouns: new FormControl("doNothing"),
    title: new FormControl("doNothing"),
    birthDate: new FormControl("doNothing"),
    gender: new FormControl("doNothing"),
  });

  public loyaltyProfile: FormGroup = new FormGroup({

  });

  constructor(
    private ucpService: UcpService,
    public dialog: MatDialog,
    @Inject(MAT_DIALOG_DATA) public data: any) {
    this.retreive(data.id1, data.id2)
    this.matchScore = data.matchScore
  }

  ngOnInit() {

  }

  retreive(id1, id2) {
    let p1 = new Promise<any>((resolve) => {
      this.ucpService.retreiveProfile(id1).subscribe((response: any) => {
        this.guest1 = response.profiles[0]
        this.addToLoyaltyDiff(this.guest1.loyaltyProfiles, 'main')
        resolve(response)
      });
    })
    let p2 = new Promise<any>((resolve) => {
      this.ucpService.retreiveProfile(id2).subscribe((response: any) => {
        this.guest2 = response.profiles[0]
        this.addToLoyaltyDiff(this.guest2.loyaltyProfiles, 'dupe')
        resolve(response)
      })
    })
    Promise.all([p1, p2]).then(res => {
      console.log("loyalty diff: ", this.loyaltyDiff)
      Object.entries(this.loyaltyDiff).forEach(([key, value]) => {
        (<any>value).id = key
        this.loyaltyArray.push(value)
        var mainVal = ((<any>value).main || {}).id
        var dupeVal = ((<any>value).dupe || {}).id
        let defaulVal = "doNothing"
        if (dupeVal && !mainVal) {
          defaulVal = "add"
        }
        this.loyaltyProfile.addControl("profile-" + key, new FormControl(defaulVal))
      });
    })


  }

  merge() {
    console.log("Merge Config personal data: ", this.personalData.value)
    console.log("Merge Config loyalty: ", this.loyaltyProfile.value)
  }

  addToLoyaltyDiff(profiles, type) {
    (profiles || []).forEach(p => {
      if (!this.loyaltyDiff[p.id]) {
        this.loyaltyDiff[p.id] = {}
      }
      this.loyaltyDiff[p.id][type] = p
    })
  }
}


@Component({
  selector: 'delete-confirm',
  templateUrl: './ucp.component-domain-creation.html',
  styleUrls: ['./ucp.component.css']
})
export class DomainCreationModalComponent {
  name = new FormControl();
  constructor(public dialogRef: MatDialogRef<DomainCreationModalComponent>) { }

  public create() {
    this.dialogRef.close(this.name.value)
  }
  public cancel() {
    this.dialogRef.close(null)
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

@Component({
  selector: 'show-errors',
  templateUrl: './ucp.component-errors.html',
  styleUrls: ['./ucp.component.css']
})
export class UCPProfileErrorsComponent {
  errors: string[];
  totalErrors: number;
  constructor(public dialogRef: MatDialogRef<UCPProfileErrorsComponent>,
    @Inject(MAT_DIALOG_DATA) public data: any) {
    this.errors = data.errors
    this.totalErrors = data.totalErrors
  }

  public close() {
    this.dialogRef.close()
  }
}