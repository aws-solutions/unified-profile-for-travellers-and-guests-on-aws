import { Component, OnInit, OnDestroy } from '@angular/core';
import { FormGroup, FormArray, FormControl, Validators, AbstractControl } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { LoaderService } from './service/loaderService';
import { DomainService } from './service/domainService'
import { MatDialog, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { faSpinner } from '@fortawesome/free-solid-svg-icons';
import { faCog, faSearch, faTimes, faPlus, faHome, faSquareCaretDown, faSignOut } from '@fortawesome/free-solid-svg-icons';
import { Subscription } from 'rxjs';
import { SessionService } from './service/sessionService';
import { UcpService } from './service/ucpService';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css'],
  providers: [DomainService]
})
export class AppComponent implements OnInit, OnDestroy {
  title = 'rep';
  faSpinner = faSpinner
  public itemForm = new FormGroup({})
  faCog = faCog;
  faHome = faHome;
  faSearch = faSearch;
  faTimes = faTimes;
  faPlus = faPlus;
  faSignOut = faSignOut;
  faSqaureCaretDown = faSquareCaretDown
  public domains: any[] = [];
  private domainSubscription: Subscription
  private selectDomainSubscription: Subscription
  public selectedDomain: string = ""
  public isMenuVisible = false;
  public isPageMenuVisible = false;
  currentRoute: string = ""

  constructor(public _loader: LoaderService, public router: Router, public dialog: MatDialog,
    public _domain: DomainService, public session: SessionService, public route: ActivatedRoute, public ucpService: UcpService) {
    this._domain.loadDomains()
    this.domains = this._domain.domains
    this.selectedDomain = this._domain.getSelectedDomain()
  }

  ngOnInit(): void {

  }


  ngOnDestroy(): void {
    this.domainSubscription.unsubscribe()
    this.selectDomainSubscription.unsubscribe()
  }
  public isLoading() {
    return this._loader.isLoading();
  }

  goToSettings() {
    this.isPageMenuVisible = false
    this.router.navigate(["setting"])
  }

  goToHome() {
    this.isPageMenuVisible = false
    this.router.navigate(["home"])
  }

  public toggleNavbar() {
    this._domain.loadDomains().then(domains => {
      this.domains = domains
    })
    this.isMenuVisible = !this.isMenuVisible;
  }

  public togglePageMenu() {
    this.isPageMenuVisible = !this.isPageMenuVisible
  }

  getDisplayStyle() {
    return this.isMenuVisible ? "block" : "none";
  }

  getPageMenuStyle() {
    return this.isPageMenuVisible ? "block" : "none";
  }

  logout() {
    this.session.clear()
    this.router.navigate(["login"])
  }

  showNavBar(): boolean {
    return (this.route.snapshot.routeConfig.path !== "login")
  }

  selectDomain(domain) {
    this._domain.selectDomain(domain)
    this.selectedDomain = this._domain.getSelectedDomain()
    this.isMenuVisible = false;
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
          this._domain.loadDomains()
        })
      }
    });
  }

}

@Component({
  selector: 'delete-confirm',
  templateUrl: './app.component-domain-creation.html',
  styleUrls: ['./app.component.css']
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



