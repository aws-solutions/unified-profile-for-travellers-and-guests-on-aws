import { Injectable } from '@angular/core';
import { Configuration } from '../app.constants';

@Injectable()
export class LocalStorageService {

    constructor(private _configuration: Configuration) {

    }

    public set(key: string, obj: Object) {
        return localStorage.setItem(key, JSON.stringify(obj));
    }

    public get(key: string): Object {
        try {
            return JSON.parse(localStorage.getItem(key));
        } catch (ex) {
            return localStorage.getItem(key);
        }

    }

    public clear(key: string) {
        return localStorage.removeItem(key);
    }

}

