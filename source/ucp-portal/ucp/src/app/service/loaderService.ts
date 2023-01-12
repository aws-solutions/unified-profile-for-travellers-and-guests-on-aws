import { Injectable } from '@angular/core';

@Injectable()
export class LoaderService {
    public loading: boolean = false;
    public count: number = 0;
    constructor() {

    }

    public start() {
        this.count++;
    }

    public stop() {
        if(this.count > 0) {
            this.count--;
        }
    }

    public isLoading() {
        return this.count > 0;
    }

}

