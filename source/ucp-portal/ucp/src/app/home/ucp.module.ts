import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { UCPComponent, UCPDetailComponent, UCPCompareComponent, UCPProfileErrorsComponent, DomainCreationModalComponent, UCPProfileDeletionConfirmationComponent, UCPSettingsComponent, LinkConnectorComponent } from './ucp.component';

import { HttpClientModule } from '@angular/common/http';

import { MatTableModule } from '@angular/material/table';
import { MatDialogModule } from '@angular/material/dialog';
import { MatSelectModule } from '@angular/material/select';
import { MatFormFieldModule, } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { FontAwesomeModule } from '@fortawesome/angular-fontawesome';



import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

@NgModule({
  declarations: [
    UCPComponent,
    UCPDetailComponent,
    UCPCompareComponent,
    UCPProfileErrorsComponent,
    DomainCreationModalComponent,
    UCPProfileDeletionConfirmationComponent,
    UCPSettingsComponent,
    LinkConnectorComponent,
  ],
  imports: [
    ReactiveFormsModule,
    FormsModule,
    BrowserModule,
    HttpClientModule,
    MatTableModule,
    MatDialogModule,
    MatSelectModule,
    BrowserAnimationsModule,
    MatFormFieldModule,
    MatInputModule,
    FontAwesomeModule
  ],
  exports: [],
  providers: [],
  bootstrap: [UCPComponent]
})
export class UcpModule { }
