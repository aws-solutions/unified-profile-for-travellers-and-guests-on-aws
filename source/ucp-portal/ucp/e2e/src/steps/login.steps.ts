import { Before, Given, Then, When, setDefaultTimeout } from '@cucumber/cucumber';
import { expect } from 'chai';

import { LoginPage } from '../pages/login.po';
import { HomePage } from '../pages/home.po';
import { Utils } from '../utils/utils';

let loginPage: LoginPage;
let homePage: HomePage;
let domainName: string;

setDefaultTimeout(50000)

Before(() => {
    loginPage = new LoginPage();
    homePage = new HomePage();
    domainName = "test-domain-e2e" + Utils.randomString(6)
});

Given(/^I am on the login page$/, async () => {
    await loginPage.navigateTo();

});
When(/^I do nothing$/, () => { });
Then(/^I should see the title$/, async () => {
    expect(await loginPage.getTitleText()).to.equal('Unified customer profile for travellers and guests');
})

When(/^I enter my temporary password and validate$/, async () => {
    await loginPage.enterLogin()
    await loginPage.enterPassword()
    await loginPage.clickLogin('login-btn')
    await loginPage.waitForFormUpdate('update-btn')
});
Then(/^I should see the form to set a new password$/, async () => {
    expect(await loginPage.getButtonText('update-btn')).to.equal('Update Password');
});
When(/^I type in a new password and confirm$/, async () => {
    await loginPage.enterPassword()
    await loginPage.confirmPassword()
    await loginPage.clickLogin('update-btn')
});
Then(/^I am redirected to the login page$/, async () => {
    expect(await loginPage.getTitleText()).to.equal('Unified customer profile for travellers and guests');
});
When(/^I enter my credentials and validate$/, async () => {
    await loginPage.enterLogin()
    await loginPage.enterPassword()
    await loginPage.clickLogin('login-btn')
    await loginPage.waitForLogin()
});
Then(/^I should be redirected to the home screen$/, async () => {
    expect(await homePage.getHeaderText()).to.equal('Unified Customer Profile for Travellers and Guests');
});
//Creation of Domain
Given(/^I am on the home screen$/, async () => {
    expect(await homePage.getHeaderText()).to.equal('Unified Customer Profile for Travellers and Guests');
});
When(/^I click on the New Domain button$/, async () => {
    await homePage.clickHome('new-domain-button')
    await homePage.waitForFormUpdate('domain-creation-text-box')
});
Then(/^I should see the New Environment screen$/, async () => {
    expect(await homePage.waitForFormUpdate('domain-creation-text-box')).to.equal(true);
});
When(/^I enter the name of my new domain and click create$/, async () => {
    await homePage.fillTextBox(domainName, "domain-creation-text-box")
    await homePage.clickHome('domain-creation-confirm')
    console.log("Domain is being created")
});
Then(/^I should see the new domain$/, async () => {
    expect(await homePage.waitForDomainCreation(domainName)).to.equal(true);
});
Given(/^I have created A new domain$/, async () => {
    expect(await homePage.waitForFormUpdate('domain-component-name-' + domainName)).to.equal(true)
});
When(/^I click on the settings button$/, async () => {
    await homePage.clickHome('settings-button')
});
Then(/^I see the settings window appears with all integrations$/, async () => {
    expect(await homePage.waitForFormUpdate('business-obj-name-air_booking')).to.equal(true);
    expect(await homePage.waitForFormUpdate('business-obj-name-passenger_profile')).to.equal(true);
    expect(await homePage.waitForFormUpdate('business-obj-name-guest_profile')).to.equal(true);
    expect(await homePage.waitForFormUpdate('business-obj-name-clickstream')).to.equal(true);
    expect(await homePage.waitForFormUpdate('business-obj-name-hotel_stay_revenue')).to.equal(true);
    expect(await homePage.waitForFormUpdate('business-obj-name-hotel_booking')).to.equal(true);
});
Given(/^I am on the settings window$/, async () => {
    expect(await homePage.waitForFormUpdateDialog('ucp-settings')).to.equal(true);
});
When(/^I click the Close X button$/, async () => {
    await homePage.clickHome('setting-close-button')
});
Then(/^I should no longer see settings window$/, async () => {
    expect(await homePage.waitForFormUpdateInvisible('business-obj-air_booking')).to.equal(true);
});
Given(/^I have selected the new test domain$/, async () => {
    expect(await homePage.confirmDomainSelected(domainName));
});
When(/^I click on the X in the Domain Component$/, async () => {
    await homePage.clickHome('delete-domain-icon-' + domainName)
});
Then(/^I see the Confirm Deletion Screen$/, async () => {
    expect(await homePage.waitForFormUpdateDialog('delete-confirm')).to.equal(true);
});
When(/^I Click Yes$/, async () => {
    await homePage.clickHome('domain-deletion-confirm')
});
Then(/^I see the Domain Component be Deleted$/, async () => {
    expect(await homePage.waitForFormUpdateInvisible('domain-component-name-' + domainName)).to.equal(true);
});