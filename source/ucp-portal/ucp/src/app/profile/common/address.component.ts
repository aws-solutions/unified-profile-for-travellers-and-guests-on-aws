import { Component, Input } from '@angular/core';
import { Address } from '../../model/address.model'
@Component({
    selector: 'ucp-address',
    templateUrl: './address.component.html'
})
export class AddressComponent {
    @Input() address: Address;
}
