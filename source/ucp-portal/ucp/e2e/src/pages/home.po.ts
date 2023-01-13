import { browser, by, element } from 'protractor';

export class HomePage {
    navigateTo() {
        console.log("Navigate to home page")
        return browser.get(browser.baseUrl + "/home") as Promise<any>;
    }

    getHeaderText() {
        console.log("Get Home page header")
        return element(by.id('header')).getText() as Promise<string>;
    }



}