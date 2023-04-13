import { Component, OnInit, Inject, OnDestroy, Input } from '@angular/core';
import * as moment from 'moment';
import { UcpService } from '../../../service/ucpService';
import { FormGroup, FormControl, Validators } from '@angular/forms';
import { MatDialog, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { ActivatedRoute } from '@angular/router';
import { faUser, faPhone, faEnvelope, faMapMarker, faBriefcase, faBirthdayCake, faExclamationTriangle, faPlane, faMousePointer, faHotel, faUsd } from '@fortawesome/free-solid-svg-icons';

@Component({
    selector: 'hotel-booking-widget',
    templateUrl: './hotel-booking.widget.html',
    styleUrls: ['../../profile.component.css']
})
export class HotelBookingWidget implements OnInit {
    @Input()
    traveller: any = {}
    @Input()
    config: any = {}
    @Input()
    objectType = ""

    constructor(public dialog: MatDialog,
        private route: ActivatedRoute,
        private ucpService: UcpService) { }

    ngOnInit() {

    }
}