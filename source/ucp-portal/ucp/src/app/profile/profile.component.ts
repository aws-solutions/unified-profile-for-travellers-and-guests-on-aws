import { Component, OnInit, Inject, OnDestroy } from '@angular/core';
import * as moment from 'moment';
import { UcpService } from '../service/ucpService';
import { FormGroup, FormControl, Validators } from '@angular/forms';
import { MatDialog, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { ActivatedRoute } from '@angular/router';
import { faUser } from '@fortawesome/free-solid-svg-icons';


@Component({
  selector: 'app-profile',
  templateUrl: './profile.component.html',
  styleUrls: ['./profile.component.css']
})
export class ProfileComponent implements OnInit {
  faUser = faUser
  guest: any = {};
  matches: any = []
  propertyMap: any = {};
  bookStart = 0;
  bookPageSize = 10
  bookEnd = this.bookPageSize;
  totalRevenueGenertated = 0;
  constructor(public dialog: MatDialog,
    private route: ActivatedRoute,
    private ucpService: UcpService) {
    this.retreive(this.route.snapshot.params.id)

  }

  showSideBySide(matchId: string, matchScore: number) {
    const dialogRef = this.dialog.open(ProfileCompareComponent, {
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

