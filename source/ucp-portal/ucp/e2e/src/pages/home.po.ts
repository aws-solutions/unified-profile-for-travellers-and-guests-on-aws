import { browser, by, element, ExpectedConditions } from 'protractor';

export class HomePage {
    navigateTo() {
        console.log("Navigate to home page")
        return browser.get(browser.baseUrl + "/home") as Promise<any>;
    }

    getHeaderText() {
        console.log("Get Home page header")
        return element(by.id('header')).getText() as Promise<string>;
    }

    async clickHome(btn_id: string) {
        console.log("Click on " + btn_id + " button")
        let button = element(by.id(btn_id));
        await browser.executeScript("arguments[0].scrollIntoView();", button.getWebElement());
        return button.click()
    }

    waitForFormUpdateDialog(form_element: string) {
        console.log("Waiting for form update")
        return browser.wait(ExpectedConditions.visibilityOf(element(by.css(form_element))), 20000)
    }

    waitForFormUpdate(form_element: string) {
        console.log("Waiting for form update")
        return browser.wait(ExpectedConditions.visibilityOf(element(by.id(form_element))), 5000)
    }

    waitForFormUpdateInvisible(form_element: string) {
        console.log("Waiting for form update")
        return browser.wait(ExpectedConditions.invisibilityOf(element(by.id(form_element))), 5000)
    }

    fillTextBox(text: string, boxId: string) {
        console.log("Entering Text " + text)
        return element(by.id(boxId)).sendKeys(text) as Promise<void>;
    }

    waitForDomainCreation(domain_name: string) {
        console.log("Waiting for the Domain to be Created")
        return browser.wait(async () => {
            let domain_test = element(by.className('domain domain-selected'))
            let domain_innerText = await domain_test.getText()
            let domain_name_created = domain_innerText.toString().split('\n')[0]
            return (domain_name_created === domain_name);
        }, 30000) as Promise<boolean>;
    }

    confirmDomainSelected(domain_name: string) {
        console.log("Confirming selected domain is the test domain")
        return browser.wait(async () => {
            let domain_test = element(by.className('domain domain-selected'))
            let domain_innerText = await domain_test.getText()
            let domain_new = element(by.id('domain-component-name-' + domain_name))
            let domain_new_innerText = await domain_new.getText()
            return (domain_innerText === domain_new_innerText);
        }, 5000) as Promise<boolean>;
    }

}