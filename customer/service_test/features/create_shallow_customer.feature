Feature: Create shallow customer
  In order to attach identity to a shopper early
  As the customer service
  I need to create a shallow customer during login or basket entry

  Scenario: A logged-in customer becomes a shallow customer
    Given customer "login-alice" does not exist yet
    When logged-in customer "login-alice" is recognized for the first time
    Then the customer command succeeds
    And customer "login-alice" can be retrieved as a shallow customer

  Scenario: A basket customer becomes a shallow customer
    Given customer "basket-bob" does not exist yet
    When basket customer "basket-bob" adds an item for the first time
    Then the customer command succeeds
    And customer "basket-bob" can be retrieved as a shallow customer

  Scenario: Creating the same shallow customer twice is rejected
    Given customer "repeat-carol" already exists as a shallow customer
    When logged-in customer "repeat-carol" is recognized for the first time
    Then the customer service rejects the request because the customer already exists
    And customer "repeat-carol" can be retrieved as a shallow customer
