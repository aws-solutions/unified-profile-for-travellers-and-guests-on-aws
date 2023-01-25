Feature: Login
  Display the Login page and enter credentials

  Scenario: Login page successfully load
    Given I am on the login page
    When I do nothing
    Then I should see the title

  Scenario: Login with valid credentials
    Given I am on the login page
    When I enter my credentials and validate
    Then I should be redirected to the home screen

  Scenario: Login with temporary password
    Given I am on the login page
    When I enter my credentials and validate
    Then I should see the form to set a new password
    When I type in a new password and confirm
    Then I am redirected to the login page