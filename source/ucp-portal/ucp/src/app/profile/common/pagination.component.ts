import { Component, Input, Output, EventEmitter } from '@angular/core';
import { faRefresh, faForward, faBackward } from '@fortawesome/free-solid-svg-icons';

@Component({
    selector: 'ucp-pagination',
    templateUrl: './pagination.component.html',
    styleUrls: ['pagination.component.css']
})
export class PaginationComponent {
    @Input() pageSize: number = 0;
    @Input() nRecords: number = 0;
    @Output() pageChange = new EventEmitter<number>();
    page: number = 0;

    faRefresh = faRefresh
    faForward = faForward
    faBackward = faBackward

    constructor() {
        console.log(this.page)
        console.log(this.pageSize)
    }

    setPage(page: number): void {
        this.page = page;
        this.pageChange.emit(page);
    }

    pageUp() {
        this.setPage(this.page + 1)
    }
    pageDown() {
        if (this.page > 0) {
            this.setPage(this.page - 1)
        }
    }

    refresh() {
        this.page = 0
        this.pageChange.emit(this.page);
    }
}