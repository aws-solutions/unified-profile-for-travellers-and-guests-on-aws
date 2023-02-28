Feature: Login
  Display the Login page and enter credentials

  Scenario: Login page successfully load
    Given I am on the login page
    When I do nothing
    Then I should see the title

  Scenario: Login with temporary password
    Given I am on the login page
    When I enter my temporary password and validate
    Then I should see the form to set a new password
    When I type in a new password and confirm
    Then I am redirected to the login page
  
  Scenario: Login with valid credentials
    Given I am on the login page
    When I enter my credentials and validate
    Then I should be redirected to the home screen

  Scenario: Create a new domain
    Given I am on the home screen
    When I click on the New Domain button
    Then I should see the New Environment screen
    When I enter the name of my new domain and click create
    Then I should see the new domain
  
  Scenario: Go to Settings Window
    Given I have created A new domain
    When I click on the settings button
    Then I see the settings window appears with all integrations

  Scenario: Close the Settings Window
    Given I am on the settings window
    When I click the Close X button
    Then I should no longer see settings window

  Scenario: Delete New Domain
    Given I have selected the new test domain
    When I click on the X in the Domain Component
    Then I see the Confirm Deletion Screen
    When I Click Yes
    Then I see the Domain Component be Deleted
