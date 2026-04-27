Feature: Merge user carts
  In order to preserve shopping progress when a guest customer becomes an identified customer
  As the store service
  I need to merge one customer's cart state into another customer's cart state and clear any checkout reservations made invalid by the merge

  Scenario: A guest customer has no cart worth merging
    Given guest customer "Guest Maya" has no active cart contents
    And signed-in customer "Maya" has an empty cart
    When the store merges customer "Guest Maya" into customer "Maya"
    Then the store completes the merge without changing either cart

  Scenario: A merge fails because guest checkout reservations cannot be fully released
    Given guest customer "Guest Quinn" has an active cart containing 1 unit of product "Lawn Mower Blade"
    And signed-in customer "Quinn" has an active cart
    And guest customer "Guest Quinn" has a pending checkout reserving 2 units of product "Lawn Mower Blade"
    And signed-in customer "Quinn" has a pending checkout with no reserved items
    And stock for product "Lawn Mower Blade" has 10 available units and 1 reserved unit
    When the store merges customer "Guest Quinn" into customer "Quinn"
    Then the store rejects the request

  Scenario: A merge fails because signed-in customer reservations cannot be fully released
    Given guest customer "Guest Ruby" has an active cart containing 1 unit of product "Lawn Mower Blade"
    And signed-in customer "Ruby" has an active cart
    And guest customer "Guest Ruby" has a pending checkout with no reserved items
    And signed-in customer "Ruby" has a pending checkout reserving 2 units of product "Lawn Mower Blade"
    And stock for product "Lawn Mower Blade" has 10 available units and 1 reserved unit
    When the store merges customer "Guest Ruby" into customer "Ruby"
    Then the store rejects the request

  Scenario: A guest cart is merged into an existing active customer cart
    Given guest customer "Guest Sofia" has an active cart containing 1 unit of product "Lawn Mower Blade"
    And signed-in customer "Sofia" has an active cart containing 2 units of product "Lawn Mower Blade"
    And guest customer "Guest Sofia" has a pending checkout reserving 1 unit of product "Lawn Mower Blade"
    And signed-in customer "Sofia" has a pending checkout reserving 2 units of product "Lawn Mower Blade"
    And stock for product "Lawn Mower Blade" has 7 available units and 3 reserved units
    When the store merges customer "Guest Sofia" into customer "Sofia"
    Then signed-in customer "Sofia" has 3 units of product "Lawn Mower Blade" in the merged cart
    And guest customer "Guest Sofia"'s cart becomes inactive
    And both pending checkouts become invalidated
    And stock for product "Lawn Mower Blade" has 10 available units and 0 reserved units

  Scenario: A guest cart is merged into a customer who has no cart yet
    Given guest customer "Guest Theo" has an active cart containing 2 units of product "Lawn Mower Blade"
    And signed-in customer "Theo" does not have a cart yet
    And neither customer has a checkout
    When the store merges customer "Guest Theo" into customer "Theo"
    Then the store creates a new active cart for signed-in customer "Theo"
    And the new cart contains 2 units of product "Lawn Mower Blade"

  Scenario: Two customers both have carts and neither has started checkout yet
    Given guest customer "Guest Uma" has an active cart containing 1 unit of product "Lawn Mower Blade"
    And signed-in customer "Uma" has an active cart containing 2 units of product "Garden Hose"
    And neither customer has a checkout
    When the store merges customer "Guest Uma" into customer "Uma"
    Then the store completes the merge successfully
