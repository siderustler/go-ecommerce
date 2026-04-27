Feature: Remove product from cart
  In order to keep cart contents and checkout reservations consistent
  As the store service
  I need to remove products safely and invalidate stale checkouts after cart changes

  Scenario: A customer removes one unit while more quantity remains
    Given customer "Brian" has an active cart with 2 units of product "Garden Hose"
    And customer "Brian" has no checkout yet
    When customer "Brian" removes 1 unit of product "Garden Hose" from the cart
    Then the cart contains 1 unit of product "Garden Hose"

  Scenario: A customer removes the last unit of a product from the cart
    Given customer "Clara" has an active cart with 1 unit of product "Pressure Washer"
    And customer "Clara" has no checkout yet
    When customer "Clara" removes 1 unit of product "Pressure Washer" from the cart
    Then product "Pressure Washer" is removed from the cart

  Scenario: A customer removes a product that is not in the cart
    Given customer "Daniel" has an active cart with no units of product "Leaf Blower"
    And customer "Daniel" has no checkout yet
    When customer "Daniel" tries to remove 1 unit of product "Leaf Blower" from the cart
    Then the store rejects the request

  Scenario: A customer tries to remove more units than the cart contains
    Given customer "Emma" has an active cart with 1 unit of product "Hedge Trimmer Blade"
    And customer "Emma" has no checkout yet
    When customer "Emma" tries to remove 2 units of product "Hedge Trimmer Blade" from the cart
    Then the store rejects the request
    And the cart still contains 1 unit of product "Hedge Trimmer Blade"

  Scenario: A customer changes the cart but checkout reservations cannot be fully released
    Given customer "Grace" has an active cart with 2 units of product "Work Gloves"
    And customer "Grace" has a pending checkout reserving 2 units of product "Work Gloves"
    And stock for product "Work Gloves" has 9 available units and 1 reserved unit
    When customer "Grace" removes 1 unit of product "Work Gloves" from the cart
    Then the store rejects the request
    And the cart keeps the already applied quantity change

  Scenario: A customer removes a product after a valid pending checkout exists
    Given customer "Hector" has an active cart with 2 units of product "Pruning Shears"
    And customer "Hector" has a pending checkout reserving 2 units of product "Pruning Shears"
    And stock for product "Pruning Shears" has 8 available units and 2 reserved units
    When customer "Hector" removes 1 unit of product "Pruning Shears" from the cart
    Then the cart contains 1 unit of product "Pruning Shears"
    And the store invalidates the pending checkout
    And the reservation for product "Pruning Shears" is released

  Scenario: A customer removes a product after checkout when one reserved stock record is missing
    Given customer "Iris" has an active cart with 2 units of product "Water Pump Filter"
    And customer "Iris" has a pending checkout reserving 1 unit of product "Discontinued Nozzle"
    And there is no stock record for product "Discontinued Nozzle"
    When customer "Iris" removes 1 unit of product "Water Pump Filter" from the cart
    Then the cart contains 1 unit of product "Water Pump Filter"
    And the store invalidates the pending checkout
    And the missing stock record does not block the request
