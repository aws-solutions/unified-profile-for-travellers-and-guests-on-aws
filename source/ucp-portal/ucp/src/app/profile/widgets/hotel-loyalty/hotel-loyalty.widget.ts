import { Component, OnInit, Output, EventEmitter, Input } from '@angular/core';
import { UcpService } from '../../../service/ucpService';
import { MatDialog } from '@angular/material/dialog';
import { ActivatedRoute } from '@angular/router';
import { Traveller, AirLoyalty } from '../../../model/traveller.model'
import { PaginationOptions } from '../../../model/pagination.model'

@Component({
    selector: 'hotel-loyalty-widget',
    templateUrl: './hotel-loyalty.widget.html',
    styleUrls: ['../../profile.component.css']
})
export class HotelLoyaltyWidget implements OnInit {
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