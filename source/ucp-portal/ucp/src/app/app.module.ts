import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent, DomainCreationModalComponent } from './app.component';
import { LoginComponent } from './login/login.component';
import { UcpModule } from './home/ucp.module';

import { HttpClientModule } from '@angular/common/http';
import { Configuration } from './app.constants';
import { SessionService } from './service/sessionService';
import { RestFactory } from './service/restFactory';
import { LoaderService } from './service/loaderService';
import { NotificationService } from './service/notificationService';
import { UcpService } from './service/ucpService';
import { UserEngagementService } from './service/userEngagementService';
import { LocalStorageService } from './service/localStorageService';

import { MatTableModule } from '@angular/material/table';
import { MatDialogModule } from '@angular/material/dialog';
import { MatSelectModule } from '@angular/material/select';
import { MatSnackBarModule } from '@angular/material/snack-bar';
import { MatFormFieldModule, } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';

import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ReactiveFormsModule } from '@angular/forms';
import { FontAwesomeModule } from '@fortawesome/angular-fontawesome';
import { ProfileComponent, ProfileCompareComponent } from './profile/profile.component';
import { AddressComponent } from './profile/common/address.component';
import { DataElementComponent } from './profile/common/data-element.component';
import { PaginationComponent } from './profile/common/pagination.component';


import { AirBookingWidget } from './profile/widgets/air-booking/air-booking.widget';
import { AirLoyaltyWidget } from './profile/widgets/air-loyalty/air-loyalty.widget'
import { ClickstreamWidget } from './profile/widgets/clickstream/clickstream.widget'
import { EmailHistoryWidget } from './profile/widgets/email-history/email-history.widget'
import { HotelBookingWidget } from './profile/widgets/hotel-booking/hotel-booking.widget'
import { HotelLoyaltyWidget } from './profile/widgets/hotel-loyalty/hotel-loyalty.widget'
import { PhoneHistoryWidget } from './profile/widgets/phone-history/phone-history.widget'
import { StayRevenueWidget } from './profile/widgets/stay-revenue/stay-revenue.widget'


import { SettingComponent, LinkConnectorComponent, RecordDisplayComponent } from './setting/setting.component';

@NgModule({
  declarations: [
    AppComponent,
    LoginComponent,
    ProfileComponent,
    ProfileCompareComponent,
    SettingComponent,
    LinkConnectorComponent,
    RecordDisplayComponent,
    DomainCreationModalComponent,
    AddressComponent,
    DataElementComponent,
    PaginationComponent,
    AirBookingWidget,
    AirLoyaltyWidget,
    ClickstreamWidget,
    EmailHistoryWidget,
    HotelBookingWidget,
    HotelLoyaltyWidget,
    PhoneHistoryWidget,
    StayRevenueWidget
  ],
  imports: [
    UcpModule,
    ReactiveFormsModule,

    BrowserModule,
    AppRoutingModule,

    HttpClientModule,
    MatTableModule,
    MatDialogModule,
    MatSelectModule,
    MatSnackBarModule,
    BrowserAnimationsModule,
    MatFormFieldModule,
    MatInputModule,
    FontAwesomeModule
  ],
  exports: [ReactiveFormsModule],
  providers: [Configuration, SessionService, LocalStorageService, RestFactory,
    LoaderService,
    NotificationService,
    UcpService,
    UserEngagementService],
  bootstrap: [AppComponent]
})
export class AppModule { }