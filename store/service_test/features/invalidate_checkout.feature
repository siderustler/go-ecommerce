Feature: Invalidate checkout
  In order to prevent customers from paying against stale cart or stock state
  As the store service
  I need to invalidate pending checkouts and release their reservations safely

  Scenario: A checkout is already invalidated before another invalidation request arrives
    Given checkout "checkout-alice" for customer "Alice" is already invalidated
    And the checkout reserves no products
    When the store invalidates checkout "checkout-alice"
    Then the store rejects the request

  Scenario: A checkout cannot release more reserved stock than exists
    Given checkout "checkout-brian" for customer "Brian" is pending and reserves 2 units of product "Garden Hose"
    And stock for product "Garden Hose" has 10 available units and 1 reserved unit
    When the store invalidates checkout "checkout-brian"
    Then the store rejects the request

  Scenario: A checkout contains a product that no longer has a stock record
    Given checkout "checkout-clara" for customer "Clara" is pending and reserves 1 unit of product "Retired Spool Head"
    And there is no stock record for product "Retired Spool Head"
    When the store invalidates checkout "checkout-clara"
    Then the checkout becomes invalidated
    And the missing stock record does not block invalidation

  Scenario: A valid pending checkout is invalidated and its reservation is released
    Given checkout "checkout-daniel" for customer "Daniel" is pending and reserves 2 units of product "Leaf Blower"
    And stock for product "Leaf Blower" has 8 available units and 2 reserved units
    When the store invalidates checkout "checkout-daniel"
    Then the checkout becomes invalidated
    And stock for product "Leaf Blower" has 10 available units and 0 reserved units
