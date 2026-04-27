Feature: Add product to cart
  In order to keep a customer's cart consistent with stock and checkout state
  As the store service
  I need to add products only when the request is valid and invalidate stale checkouts when cart contents change

  Scenario: A customer tries to add an out-of-stock product to an active cart
    Given customer "Alice" has an active cart containing no products
    And product "Cordless Trimmer" is not available in stock
    When customer "Alice" tries to add 1 unit of product "Cordless Trimmer" to the cart
    Then the store rejects the request
    And customer "Alice"'s cart remains unchanged

  Scenario: A customer adds the first product
    Given customer "Brian" didnt add any product yet
    And product "Garden Hose" has 10 units available in stock
    When customer "Brian" adds 2 units of product "Garden Hose" to the cart
    Then the store creates a new active cart for customer "Brian"
    And the cart contains 2 units of product "Garden Hose"

  Scenario: A customer adds a product when has a pending checkout
    Given customer "Eva" has an active cart
    And customer "Eva" has a pending checkout reserving 2 units of product "Chainsaw Chain"
    And stock for product "Chainsaw Chain" has 8 available units and 2 reserved units
    When customer "Eva" adds 1 unit of product "Chainsaw Chain" to the cart
    Then the store invalidates the existing pending checkout
    And the store releases the reservation for product "Chainsaw Chain"
    And stock for product "Chainsaw Chain" has 10 available units and 0 reserved units

  Scenario: A customer changes a cart after checkout when one reserved item has been removed entirely from stock
    Given customer "Farah" has an active cart
    And customer "Farah" has a pending checkout containing product "Retired Spool Head"
    And there is no stock record for product "Retired Spool Head"
    And product "Work Gloves" has 10 units available in stock
    When customer "Farah" adds 1 unit of product "Work Gloves" to the cart
    Then the store invalidates the pending checkout
    And the store skips the missing stock record without failing the request
