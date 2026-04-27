Feature: Create checkout from a cart
  In order to reserve inventory before payment
  As the store service
  I need to create a pending checkout only from a valid cart with reservable stock

  Scenario: A customer already has a pending checkout
    Given customer "Hannah" already has a pending checkout for product "Lawn Mower Blade"
    When customer "Hannah" starts checkout
    Then the store returns the existing pending checkout unchanged

  Scenario: A customer tries to checkout without an active cart
    Given customer "Alice" does not have a cart yet
    And there is no reserved stock for the cart
    When customer "Alice" starts checkout
    Then the store rejects the request

  Scenario: A customer tries to checkout when a cart product has no stock record
    Given customer "Brian" has a cart containing 1 unit of product "Garden Hose"
    And there is no stock record for product "Garden Hose"
    When customer "Brian" starts checkout
    Then the store rejects the request

  Scenario: A customer tries to checkout when the requested quantity cannot be reserved
    Given customer "Clara" has a cart containing 2 units of product "Pressure Washer"
    And stock for product "Pressure Washer" has 1 available unit and 0 reserved units
    When customer "Clara" starts checkout
    Then the store rejects the request

  Scenario: A customer starts checkout with a valid cart and available stock
    Given customer "Daniel" has a cart containing 2 units of product "Leaf Blower"
    And stock for product "Leaf Blower" has 5 available units and 1 reserved unit
    When customer "Daniel" starts checkout
    Then the store creates a pending checkout for customer "Daniel"
    And the checkout contains 2 units of product "Leaf Blower"
    And stock for product "Leaf Blower" has 3 available units and 3 reserved units
