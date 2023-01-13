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