import { Component, OnInit, Inject} from '@angular/core';
import { UcpService } from '../service/ucpService';
import { UserEngagementService } from '../service/userEngagementService';
import { FormGroup, FormArray, FormControl, Validators } from '@angular/forms';
import { MatDialog, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { SessionService } from '../service/sessionService';
import { faCog, faSearch, faTimes, faPlus } from '@fortawesome/free-solid-svg-icons';


import * as moment from 'moment';


@Component({
  selector: 'app-ucp',
  templateUrl: './ucp.component.html',
  styleUrls: ['./ucp.component.css']
})
export class UCPComponent implements OnInit {
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
  domains: any[];

  constructor(private ucpService: UcpService, public dialog: MatDialog, private userEngSvc: UserEngagementService, private session: SessionService) {
    this.loadDomains()

  }

  loadDomains() {
    this.ucpService.listDomains().subscribe((res: any) => {
      console.log(res)
      this.config = res.config;
      this.domains = res.config.domains
      this.selectDomain(res.config.domains[0].customerProfileDomain)
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

  showSettings() {
    const dialogRef = this.dialog.open(UCPSettingsComponent, {
      width: '90%',
      data: {
        selectedDomain: this.selectedDomain
      }
    });

    dialogRef.afterClosed().subscribe((name: any) => {
      console.log('The dialog was closed with value: ', name);

    });
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
          this.loadDomains()
        })
      }

    });
  }


  ngOnInit() {
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

@Component({
  selector: 'ucp-settings',
  templateUrl: './ucp.component-settings.html',
  styleUrls: ['./ucp.component.css']
})
export class UCPSettingsComponent {
  faCog = faCog;
  domain: any;
  industryConnectorSolutions: [];
  industryConnectors: any[] = [
    {
      id: "hapi",
      icon: "https://media-exp1.licdn.com/dms/image/C4E0BAQE38nbk86XEOQ/company-logo_200_200/0/1618338688553?e=2147483647&v=beta&t=32RQL7yl3BxcFrkhVLKiJZEgxrApfj4kJsgC2uhm6Vg",
      description: "Hapi is a Cloud Data Hub that exposes event streams and transactional APIs from hotel systems at scale",
      objectType: "hotel booking, air booking, guest profile, passenger profile, hotel stay revenue",
      deploymentStatus: "Not Deployed"
    },
    {
      id: "tealium",
      icon: "https://miu.sg/wp-content/uploads/tealium.png",
      description: "Tealium connects data so you can connect with your customers",
      objectType: "clickstream",
      deploymentStatus: "Not Deployed"
    }]

  constructor(public dialogRef: MatDialogRef<UCPSettingsComponent>, public dialog: MatDialog,
    @Inject(MAT_DIALOG_DATA) public data: any, private session: SessionService, private ucpService: UcpService) {
    let domain = this.session.getProfileDomain()
    //this.showDetail({ "unique_id": "906cd43cae044345b1f2027ad9465fdb" })
    this.ucpService.getConfig(domain).subscribe((res: any) => {
      console.log(res)
      this.domain = res.config.domains[0];
    })
    this.ucpService.listApplications().subscribe((res: any) => {
      this.industryConnectorSolutions = res;
    })
  }

  public validate() {
    this.dialogRef.close(true)
  }
  public invalidate() {
    this.dialogRef.close(false)
  }
  public openDeployConnectorLink(): void {
    // TODO: replace with Industry Connector solution deployment link
    window.open("https://github.com/aws-solutions/travel-and-hospitality-connectors", "_blank", "noopener, noreferrer")
  }
  public shouldShowLinkButton(): boolean {
    return this.industryConnectorSolutions.length > 0 ? true : false
  }
  showLinkConnector() {
    const dialogRef = this.dialog.open(LinkConnectorComponent, {
      width: '90%',
    });
  }
}

@Component({
  selector: 'link-connector',
  templateUrl: './connector/ucp.component-link-connector.html',
  styleUrls: ['./ucp.component.css']
})
export class LinkConnectorComponent {
  data: any;
  domain: string;
  linkConnectorForm = new FormGroup({
    agwUrl: new FormControl('', Validators.required),
    tokenEndpoint: new FormControl('', Validators.required),
    clientId: new FormControl('', Validators.required),
    clientSecret: new FormControl('', Validators.required),
    bucketArn: new FormControl('', Validators.required),
  });
  buttonDisabled: boolean;

  constructor(public dialogRef: MatDialogRef<LinkConnectorComponent>, private ucpService: UcpService, private session: SessionService, public dialog: MatDialog) {
    this.domain = this.session.getProfileDomain();
    let localData = this.session.getConnectorData(this.domain);
    this.linkConnectorForm.controls['agwUrl'].setValue(localData?.agwUrl ?? "");
    this.linkConnectorForm.controls['tokenEndpoint'].setValue(localData?.tokenEndpoint ?? "");
    this.linkConnectorForm.controls['clientId'].setValue(localData?.clientId ?? "");
    this.linkConnectorForm.controls['clientSecret'].setValue(localData?.clientSecret ?? "");
    this.linkConnectorForm.controls['bucketArn'].setValue(localData?.bucketArn ?? "");
    this.buttonDisabled = true;
  }

  ngOnInit() {
    this.linkConnectorForm.valueChanges.subscribe(() => {
      if (this.linkConnectorForm.valid) {
        this.buttonDisabled = false;
      } else {
        this.buttonDisabled = true;
      }
    });
  }

  public link() {
    this.session.setConnectorData(this.domain, this.linkConnectorForm.value.agwUrl, this.linkConnectorForm.value.tokenEndpoint, this.linkConnectorForm.value.clientId, this.linkConnectorForm.value.clientSecret, this.linkConnectorForm.value.bucketArn);
    this.ucpService.linkIndustryConnector(this.linkConnectorForm.value.agwUrl, this.linkConnectorForm.value.tokenEndpoint, this.linkConnectorForm.value.clientId, this.linkConnectorForm.value.clientSecret, this.linkConnectorForm.value.bucketArn).subscribe((res: any) => {
      this.data = res;
      this.dialogRef.close(null);
      const dialogRef = this.dialog.open(CreateConnectorCrawler, {
        width: '90%',
        data: {
          bucketPolicy: this.data["BucketPolicy"],
          glueRoleArn: this.data["GlueRoleArn"],
          bucketPath: this.linkConnectorForm.controls['bucketArn'].value,
        }
      });
    });
  }

  public cancel() {
    this.dialogRef.close(null)
  }
}

@Component({
  selector: 'create-connector-crawler',
  templateUrl: './connector/ucp.component-create-connector-crawler.html',
  styleUrls: ['./ucp.component.css']
})
export class CreateConnectorCrawler {
  bucketPolicy: string;
  glueRoleArn: string;
  bucketPath: string;
  constructor(public dialogRef: MatDialogRef<CreateConnectorCrawler>, private ucpService: UcpService, @Inject(MAT_DIALOG_DATA) public data: any) {
    this.bucketPolicy = data.bucketPolicy;
    this.glueRoleArn = data.glueRoleArn;
    this.bucketPath = data.bucketPath;
  }

  public link() {
    this.ucpService.createConnectorCrawler(this.glueRoleArn, this.bucketPath).subscribe((res: any) => {
    });
    this.dialogRef.close(null);
  }

  public cancel() {
    this.dialogRef.close(null)
  }
}