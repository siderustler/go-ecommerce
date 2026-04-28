Feature: Save customer profile
  In order to continue buying products and manage profile details
  As the customer service
  I need to save customer credentials and billing details consistently

  Scenario: A customer saves checkout credentials and billing details for the first time
    When customer "checkout-daria" saves customer profile with credentials "Daria Buyer" "daria@example.com" "+48100100100" and billing details "1234567890" "Acme Buyer" "Warsaw" "Main 1" "00-001" "2A"
    Then the customer command succeeds
    And the customer profile for "checkout-daria" contains credentials "Daria Buyer" "daria@example.com" "+48100100100" and billing details "1234567890" "Acme Buyer" "Warsaw" "Main 1" "00-001" "2A"

  Scenario: A customer adds credentials and billing details in the Customer Profile
    Given customer "profile-edith" already exists as a shallow customer
    When customer "profile-edith" saves customer profile with credentials "Edith Account" "edith@example.com" "+48100100101" and billing details "2234567890" "Profile Works" "Krakow" "River 9" "30-002" "4B"
    Then the customer command succeeds
    And the customer profile for "profile-edith" contains credentials "Edith Account" "edith@example.com" "+48100100101" and billing details "2234567890" "Profile Works" "Krakow" "River 9" "30-002" "4B"

  Scenario: Saving the customer profile again overwrites credentials and billing details
    Given customer "rewrite-frank" has saved customer profile with credentials "Frank Old" "frank-old@example.com" "+48100100102" and billing details "3234567890" "Old Parts" "Lodz" "Old 5" "90-003" "8C"
    When customer "rewrite-frank" saves customer profile with credentials "Frank New" "frank-new@example.com" "+48100100103" and billing details "4234567890" "New Parts" "Poznan" "New 7" "60-004" "9D"
    Then the customer command succeeds
    And the customer profile for "rewrite-frank" contains credentials "Frank New" "frank-new@example.com" "+48100100103" and billing details "4234567890" "New Parts" "Poznan" "New 7" "60-004" "9D"
