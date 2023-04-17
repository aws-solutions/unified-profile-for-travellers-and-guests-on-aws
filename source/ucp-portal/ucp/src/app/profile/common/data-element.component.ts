import { Component, Input } from '@angular/core';
import { Address } from '../../model/address.model'
import { faExternalLink } from '@fortawesome/free-solid-svg-icons';

@Component({
    selector: 'ucp-data',
    templateUrl: './data-element.component.html'
})
export class DataElementComponent {
    @Input() accpRecord: string;
    @Input() fieldName: string;
    @Input() data: string;
    @Input() config: any = {
        externalUrl: {}
    };
    faExternalLink = faExternalLink;
}
