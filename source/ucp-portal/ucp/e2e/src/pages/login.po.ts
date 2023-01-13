import { browser, by, element } from 'protractor';
import creds = require('../creds.json')

export class LoginPage {
    navigateTo() {
        console.log("Navigate to login page")
        return browser.get(browser.baseUrl) as Promise<any>;
    }

    getTitleText() {
        console.log("Get login page title")
        return element(by.css('app-root h2')).getText() as Promise<string>;
    }

    enterLogin() {
        let email = (<any>creds).email
        console.log("Enter login ", email)
        return element(by.id('login')).sendKeys(email) as Promise<void>;
    }

    enterPassword() {
        let pwd = (<any>creds).pwd
        console.log("Enter password ", pwd)
        return element(by.id('pwd')).sendKeys(pwd) as Promise<void>;
    }

    async clickLogin() {
        console.log("Click on login button")
        let button = element(by.id('login-btn'));
        await browser.executeScript("arguments[0].scrollIntoView();", button.getWebElement());
        return button.click()
    }

    waitForLogin() {
        console.log("Waiting for redirection after login")
        return browser.wait(async () => {
            let newUrl = await browser.getCurrentUrl();
            console.log("Current URL:", newUrl)
            return (newUrl === 'http://localhost:4200/home');
        }, 10000) as Promise<boolean>;
    }


}