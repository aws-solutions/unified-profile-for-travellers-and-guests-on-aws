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
})
Then(/^I am redirected to the login page$/, async () => {
    expect(await loginPage.getTitleText()).to.equal('Unified customer profile for travellers and guests');
})
When(/^I enter my credentials and validate$/, async () => {
    await loginPage.enterLogin()
    await loginPage.enterPassword()
    await loginPage.clickLogin('login-btn')
    await loginPage.waitForLogin()
});
Then(/^I should be redirected to the home screen$/, async () => {
    expect(await homePage.getHeaderText()).to.equal('Unified Customer Profile for Travellers and Guests');
})