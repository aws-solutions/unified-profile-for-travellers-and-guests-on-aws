import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
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

@NgModule({
  declarations: [
    AppComponent,
    LoginComponent
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