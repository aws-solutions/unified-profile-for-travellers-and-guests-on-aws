import { Component, OnInit, Inject } from '@angular/core';
import { UcpService } from '../service/ucpService';
import { FormGroup, FormArray, FormControl, Validators } from '@angular/forms';
import { MatDialog, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { SessionService } from '../service/sessionService';
import { faCog, faBackward, faForward, faHome, faRefresh } from '@fortawesome/free-solid-svg-icons';
import { Router } from '@angular/router';
import { PaginationOptions } from '../model/pagination.model'
@Component({
  selector: 'app-setting',
  templateUrl: './setting.component.html',
  styleUrls: ['./setting.component.css']
})
export class SettingComponent implements OnInit {
  faCog = faCog;
  faHome = faHome;
  faBackward = faBackward;
  faForward = faForward;
  faRefresh = faRefresh;
  domain: any = {};
  validationPagination: PaginationOptions = {
    page: 0,
    pageSize: 20
  }
  validationPaginationSeverSide: PaginationOptions = {
    page: 0,
    pageSize: 20
  }
  validationErrorsInView = [];
  industryConnectorSolutions: [];

  selectedDomain: any;
  validationErrors = []
  industryConnectors: any[] = [
    {
      id: "hapi",
      icon: "https://media-exp1.licdn.com/dms/image/C4E0BAQE38nbk86XEOQ/company-logo_200_200/0/1618338688553?e=2147483647&v=beta&t=32RQL7yl3BxcFrkhVLKiJZEgxrApfj4kJsgC2uhm6Vg",
      description: "Hapi is a Cloud Data Hub that exposes event streams and transactional APIs from hotel systems at scale",
      objectType: "hotel booking, guest profile, hotel stay revenue",
      deploymentStatus: "Not Deployed"
    },
    {
      id: "tealium",
      icon: "https://miu.sg/wp-content/uploads/tealium.png",
      description: "Tealium connects data so you can connect with your customers",
      objectType: "clickstream",
      deploymentStatus: "Not Deployed"
    }]

  constructor(public dialog: MatDialog, private session: SessionService, private ucpService: UcpService, private router: Router) {
    this.selectedDomain = this.session.getProfileDomain()
    if (this.selectedDomain) {
      this.ucpService.getConfig(this.selectedDomain).subscribe((res: any) => {
        console.log(res)
        this.domain = res.config.domains[0];
      })
      this.ucpService.listApplications().subscribe((res: any) => {
        this.industryConnectorSolutions = res || [];
      })
    }
    this.fetchValidationErrors()

  }

  fetchValidationErrors() {
    this.ucpService.getDataValidationErrors(this.validationPaginationSeverSide).subscribe((res: any) => {
      console.log(res)
      Array.prototype.push.apply(this.validationErrors, res.dataValidation)
      this.updateValidationTable()
    })
  }

  reset() {
    this.validationErrors = []
    this.validationPaginationSeverSide.page = 0
    this.validationPagination.page = 0
    this.fetchValidationErrors()
  }

  fetchNext() {
    this.validationPaginationSeverSide.page++
    this.fetchValidationErrors()
  }

  isLastPage() {
    return this.validationErrors.length - this.validationPagination.page * this.validationPagination.pageSize <= this.validationPagination.pageSize
  }
  validationPageUp() {
    this.validationPagination.page++
    if (this.isLastPage()) {
      this.fetchNext()
    }
    this.updateValidationTable()
  }
  validationPageDown() {
    this.validationPagination.page--
    if (this.validationPagination.page < 0) {
      this.validationPagination.page = 0
    }
    this.updateValidationTable()
  }
  updateValidationTable() {
    this.validationErrorsInView = [];
    for (let i = 0; i < this.validationPagination.pageSize; i++) {
      this.validationErrorsInView.push(this.validationErrors[i + this.validationPagination.pageSize * this.validationPagination.page])
    }
  }

  ngOnInit(): void {

  }

  public openDeployConnectorLink(): void {
    // TODO: replace with Industry Connector solution deployment link
    window.open("https://github.com/aws-solutions/travel-and-hospitality-connectors", "_blank", "noopener, noreferrer")
  }
  public shouldShowLinkButton(): boolean {
    if (this.industryConnectorSolutions) {
      return this.industryConnectorSolutions.length > 0
    }
    return false;
  }

  showLinkConnector(connectorId: string) {
    const dialogRef = this.dialog.open(LinkConnectorComponent, {
      width: '90%',
      data: {
        connectorId: connectorId,
      }
    });
  }
  public getColorStatus(status: string) {
    let statusColor: string;

    switch (status) {
      case "Active": {
        statusColor = "darkgreen";
        break;
      }
      case "Deleted": {
        statusColor = "firebrick";
        break;
      }
      case "Errored": {
        statusColor = "firebrick";
        break;
      }
      case "Suspended": {
        statusColor = "firebrick";
        break;
      }
      case "Depreciated": {
        statusColor = "goldenrod";
        break;
      }
      case "Draft": {
        statusColor = "goldenrod";
        break;
      }
      default: {
        statusColor = "gray"
        break;
      }
    }
    return statusColor
  }

  goHome() {
    this.router.navigate(["home"])
  }


  public getLastRunColorStatus(status: string) {
    let statusColor: string;
    switch (status) {
      case "Successful": {
        statusColor = "darkgreen";
        break;
      }
      case "Error": {
        statusColor = "darkgreen";
        break;
      }
      case "InProgress": {
        statusColor = "goldenrod";
        break;
      }
      case "": {
        statusColor = "goldenrod"
        break;
      }
      default: {
        statusColor = "gray"
        break;
      }
    }
    return statusColor
  }

}


@Component({
  selector: 'link-connector',
  templateUrl: './connector/setting.component-link-connector.html',
  styleUrls: ['./setting.component.css']
})
export class LinkConnectorComponent {
  response: any;
  domain: string;
  linkConnectorForm = new FormGroup({
    agwUrl: new FormControl('', Validators.required),
    tokenEndpoint: new FormControl('', Validators.required),
    clientId: new FormControl('', Validators.required),
    clientSecret: new FormControl('', Validators.required),
    bucketArn: new FormControl('', Validators.required),
  });
  buttonDisabled: boolean;

  constructor(public dialogRef: MatDialogRef<LinkConnectorComponent>, private ucpService: UcpService, private session: SessionService, public dialog: MatDialog, @Inject(MAT_DIALOG_DATA) public data: any) {
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
      this.response = res;
      this.dialogRef.close(null);
      const dialogRef = this.dialog.open(CreateConnectorCrawler, {
        width: '90%',
        data: {
          bucketPolicy: this.response["BucketPolicy"],
          glueRoleArn: this.response["GlueRoleArn"],
          bucketPath: this.linkConnectorForm.controls['bucketArn'].value,
          connectorId: this.data.connectorId,
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
  templateUrl: './connector/setting.component-create-connector-crawler.html',
  styleUrls: ['./setting.component.css']
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
    this.ucpService.createConnectorCrawler(this.glueRoleArn, this.bucketPath, this.data.connectorId).subscribe((res: any) => { });
    this.dialogRef.close(null);
  }

  public cancel() {
    this.dialogRef.close(null)
  }
}