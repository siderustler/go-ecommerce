Feature: Retrieve customer
  In order to render the right customer state in the shop
  As the customer service
  I need to return the current customer profile details

  Scenario: Retrieving a missing customer
    When the customer service retrieves customer "query-missing"
    Then the customer "query-missing" is missing

  Scenario: Retrieving a shallow customer
    Given customer "query-jane" already exists as a shallow customer
    When the customer service retrieves customer "query-jane"
    Then the retrieved customer "query-jane" is a shallow customer

  Scenario: Retrieving a customer with credentials and billing details
    Given customer "query-karl" has saved customer profile with credentials "Karl Buyer" "karl@example.com" "+48100100106" and billing details "7234567890" "Karl Parts" "Poznan" "Green 11" "60-007" "15G"
    When the customer service retrieves customer "query-karl"
    Then the retrieved customer "query-karl" contains credentials "Karl Buyer" "karl@example.com" "+48100100106" and billing details "7234567890" "Karl Parts" "Poznan" "Green 11" "60-007" "15G"

  Scenario: Retrieving a customer with full profile details
    Given customer "query-lena" has saved customer profile with credentials "Lena Buyer" "lena@example.com" "+48100100107" and billing details "8234567890" "Lena Machines" "Gdynia" "Sea 3" "81-008" "16H"
    And customer "query-lena" has a shipping address "Sopot" "Beach 4" "81-701" "17I"
    When the customer service retrieves customer "query-lena"
    Then the retrieved customer "query-lena" contains credentials "Lena Buyer" "lena@example.com" "+48100100107" billing details "8234567890" "Lena Machines" "Gdynia" "Sea 3" "81-008" "16H" and shipping address "Sopot" "Beach 4" "81-701" "17I"
