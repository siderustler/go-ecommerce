Feature: Create order from checkout
  In order to finalize a successful purchase
  As the store service
  I need to consume reserved stock, finalize checkout, and archive the cart only when the order can be created consistently

  Scenario: A customer cannot create an order from an already inactive cart
    Given customer "Alice" has an inactive cart
    And checkout "checkout-alice" for customer "Alice" is pending
    And the checkout contains no products
    When the store creates an order from checkout "checkout-alice"
    Then the store rejects the request

  Scenario: A customer cannot create an order when reserved stock cannot be fully removed
    Given customer "Brian" has an active cart containing 2 units of product "Garden Hose"
    And customer "Brian" has a pending checkout "checkout-brian" which reserves 2 units of product "Garden Hose"
    And stock for product "Garden Hose" has 10 available units and 1 reserved unit
    When the store creates an order from checkout "checkout-brian"
    Then the store rejects the request

  Scenario: A customer creates an order and discounted pricing is preserved
    Given customer "Daniel" has an active cart containing 1 unit of product "Leaf Blower"
    And customer "Daniel" has pending checkout "checkout-daniel" which reserves 1 unit of product "Leaf Blower"
    And stock for product "Leaf Blower" has 10 available units and 1 reserved unit
    And product "Leaf Blower" has regular price 10 and discount price 7
    When the store creates an order from checkout "checkout-daniel"
    Then the cart becomes inactive
    And the checkout becomes finalized
    And stock for product "Leaf Blower" has 10 available units and 0 reserved units
    And the order contains 1 line for product "Leaf Blower"
    And the order line for product "Leaf Blower" uses discounted price 7
