import { Before, Given, Then, When, setDefaultTimeout } from '@cucumber/cucumber';
import { expect } from 'chai';

import { LoginPage } from '../pages/login.po';
import { HomePage } from '../pages/home.po';

let loginPage: LoginPage;
let homePage: HomePage;

setDefaultTimeout(50000)

Before(() => {
    loginPage = new LoginPage();
    homePage = new HomePage();
});

Given(/^I am on the login page$/, async () => {
    await loginPage.navigateTo();

});
When(/^I do nothing$/, () => { });
Then(/^I should see the title$/, async () => {
    expect(await loginPage.getTitleText()).to.equal('Unified customer profile for travellers and guests');
})

When(/^I enter my credentials and validate$/, async () => {
    await loginPage.enterLogin()
    await loginPage.enterPassword()
    await loginPage.clickLogin()
    await loginPage.waitForLogin()
});
Then(/^I should be redirected to the home screen$/, async () => {
    expect(await homePage.getHeaderText()).to.equal('Unified Customer Profile for Travellers and Guests');
})