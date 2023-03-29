import { Component, OnInit, OnDestroy } from '@angular/core';
import { FormGroup, FormArray, FormControl, Validators, AbstractControl } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { LoaderService } from './service/loaderService';
import { DomainService } from './service/domainService'
import { MatDialog, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { faSpinner } from '@fortawesome/free-solid-svg-icons';
import { faCog, faSearch, faTimes, faPlus, faHome, faSquareCaretDown } from '@fortawesome/free-solid-svg-icons';
import { Subscription } from 'rxjs';
import { SessionService } from './service/sessionService';

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
  faSqaureCaretDown = faSquareCaretDown
  public domains: any[] = [];
  private domainSubscription: Subscription
  private selectDomainSubscription: Subscription
  public selectedDomain: string = "Please select a domain"
  public isMenuVisible = false;
  public isPageMenuVisible = false;
  currentRoute: string = ""

  constructor(public _loader: LoaderService, public router: Router, public dialog: MatDialog, 
    public _domain: DomainService, public session: SessionService, public route: ActivatedRoute ) {
    this._domain.loadDomains()
    this.domains = this._domain.domains
    this.selectedDomain = this._domain.selectedDomain
  }

  ngOnInit(): void {
    this.domainSubscription = this._domain.domainObs.subscribe((domains: any[]) => {
      this.domains = domains
    })

    this.selectDomainSubscription = this._domain.selectedDomainObs.subscribe((selectedDomain: string) => {
      this.selectedDomain = selectedDomain
    })
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

  createDomain() {
    this.domains = this._domain.domains
    this._domain.createDomain()
    return
  }

  selectDomain(domainName: string) {
    this._domain.selectDomain(domainName)
    console.log(this.selectedDomain)
    this.isMenuVisible = false
  }

  deleteDomain(domainName: string) {
    this._domain.deleteDomain(domainName)
  }

  public toggleNavbar() {
    this._domain.loadDomains()
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

}
