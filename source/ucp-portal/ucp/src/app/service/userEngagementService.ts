import { Injectable } from '@angular/core';
import 'rxjs/add/operator/map'

import * as _ from 'lodash';


@Injectable()
export class UserEngagementService {

    constructor() {

    }

    logEvent(eventName: string, params: any) {
        console.log("[EVENT_LOGGED]: ", eventName, " with params ", params)
    }


}