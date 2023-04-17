import { Component, OnInit, Output, EventEmitter, Input } from '@angular/core';
import { UcpService } from '../../../service/ucpService';
import { MatDialog } from '@angular/material/dialog';
import { ActivatedRoute } from '@angular/router';
import { Traveller, EmailHistory } from '../../../model/traveller.model'
import { PaginationOptions } from '../../../model/pagination.model'

@Component({
    selector: 'email-history-widget',
    templateUrl: './email-history.widget.html',
    styleUrls: ['../../profile.component.css']
})
export class EmailHistoryWidget implements OnInit {
    @Input()
    traveller: Traveller = new Traveller()
    @Input()
    config: any = {}
    @Input()
    objectType: string = ""
    @Output() pageChange = new EventEmitter<PaginationOptions>();


    pageSize = 5;
    page = 0;

    records: EmailHistory[] = []

    constructor(public dialog: MatDialog,
        private route: ActivatedRoute,
        private ucpService: UcpService) {

    }

    ngOnChanges(changes: any) {
        this.updateRecords()
    }

    onPageChange(page) {
        console.log("new page: ", page)
        this.page = page
        let po = new PaginationOptions(page, this.pageSize, this.objectType)
        this.pageChange.emit(po);
    }

    ngOnInit() {

    }

    updateRecords() {
        console.log("[Phone History component] traveler: %v", this.traveller)
        this.records = this.traveller.emailHistoryRecords
        this.records.sort((r1, r2) => {
            if (r1.lastUpdated < r2.lastUpdated) {
                return 1;
            }

            if (r1.lastUpdated > r2.lastUpdated) {
                return -1;
            }

            return 0;
        });
    }
}