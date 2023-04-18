import { Component, OnInit, Inject, OnDestroy } from '@angular/core';
import * as moment from 'moment';
import { UcpService } from '../service/ucpService';
import { FormGroup, FormControl, Validators } from '@angular/forms';
import { MatDialog, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { ActivatedRoute } from '@angular/router';
import { faUser, faPhone, faEnvelope, faMapMarker, faBriefcase, faBirthdayCake, faExclamationTriangle, faPlane, faMousePointer, faHotel, faUsd, faQuestion } from '@fortawesome/free-solid-svg-icons';
import { PaginationOptions } from '../model/pagination.model'

import { AddressComponent } from './common/address.component'
import { Common } from '../model/common.model'
import { Traveller } from '../model/traveller.model'

@Component({
  selector: 'app-profile',
  templateUrl: './profile.component.html',
  styleUrls: ['./profile.component.css']
})
export class ProfileComponent implements OnInit {
  faUser = faUser
  faPhone = faPhone
  faEnvelope = faEnvelope
  faMapMarker = faMapMarker
  faBriefcase = faBriefcase
  faBirthdayCake = faBirthdayCake
  faExclamationTriangle = faExclamationTriangle
  faQuestion = faQuestion
  objectTypeIcons = {}


  errorsPagination: PaginationOptions = {
    page: 0,
    pageSize: 15
  }
  widgetConfig: any = {}
  traveller: Traveller = new Traveller();
  matches: any = []
  propertyMap: any = {};
  bookStart = 0;
  bookPageSize = 10
  bookEnd = this.bookPageSize;
  totalRevenueGenertated = 0;
  ingestionErrors = []
  DEFAULT_PAGE_SIZE: number = 5
  oldestBirthDate = new Date(1900, 1, 1)

  allPaginationOptions: Map<string, PaginationOptions> = new Map<string, PaginationOptions>()

  constructor(public dialog: MatDialog,
    private route: ActivatedRoute,
    private ucpService: UcpService) {
    this.objectTypeIcons[Common.OBJECT_TYPE_AIR_BOOKING] = { text: "Air Booking", icon: faPlane }
    this.objectTypeIcons[Common.OBJECT_TYPE_CLICKSTREAM] = { text: "Clickstream", icon: faMousePointer }
    this.objectTypeIcons[Common.OBJECT_TYPE_HOTEL_LOYALTY] = { text: "Guest Profiles", icon: faUser }
    this.objectTypeIcons[Common.OBJECT_TYPE_HOTEL_BOOKING] = { text: "Hotel Booking", icon: faHotel }
    this.objectTypeIcons[Common.OBJECT_TYPE_AIR_LOYALTY] = { text: "Passenger Profiles", icon: faUser }
    this.objectTypeIcons[Common.OBJECT_TYPE_STAY_REVENUE] = { text: "Stay Revenue", icon: faUsd }
    this.objectTypeIcons[Common.OBJECT_TYPE_PHONE_HISTORY] = { text: "Email History", icon: faPhone }
    this.objectTypeIcons[Common.OBJECT_TYPE_EMAIL_HISTORY] = { text: "Phone History", icon: faEnvelope }

    this.allPaginationOptions.set(Common.OBJECT_TYPE_AIR_BOOKING, new PaginationOptions(0, this.DEFAULT_PAGE_SIZE, Common.OBJECT_TYPE_AIR_BOOKING))
    this.widgetConfig.objectTypeIcons = this.objectTypeIcons

    this.retreive(this.route.snapshot.params.id)
    this.fetchErrors()
  }

  showSideBySide(matchId: string, matchScore: number) {
    const dialogRef = this.dialog.open(ProfileCompareComponent, {
      height: '80%',
      width: '99%',
      data: {
        id1: this.traveller.connectId,
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

  retreive(id) {
    console.log(id)
    if (id) {
      this.ucpService.retreiveProfile(id, this.allPaginationOptions).subscribe(response => this.afterRetreive(response),
        () => { });
    } else {
      console.log("no ID for guest 360 retreive")
    }
  }

  fetchErrors() {
    this.ucpService.listErrors(this.errorsPagination).subscribe((res: any) => {
      console.log(res)
      this.ingestionErrors = res.ingestionErrors || [];
    })
  }

  onWidgetChange(po: PaginationOptions) {
    this.allPaginationOptions.set(po.objectType, po)
    this.retreive(this.traveller.connectId)

  }


  afterRetreive(response) {
    console.log(response)
    this.matches = response.matches
    this.traveller = response.profiles[0]
  }

}


@Component({
  selector: 'profile-compare',
  templateUrl: 'profile.component-compare.html',
  styleUrls: ['./profile.component.css']
})
export class ProfileCompareComponent {
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

