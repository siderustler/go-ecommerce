Feature: Add shipping address
  In order to proceed with checkout and manage delivery details
  As the customer service
  I need to save a shipping address for an existing customer

  Scenario: A customer adds a missing shipping address during checkout
    Given customer "ship-gina" has saved customer profile with credentials "Gina Buyer" "gina@example.com" "+48100100104" and billing details "5234567890" "Ship Works" "Warsaw" "Market 8" "00-005" "10A"
    When customer "ship-gina" adds shipping address "Gdansk" "Harbor 2" "80-001" "7B"
    Then the customer command succeeds
    And the customer profile for "ship-gina" contains shipping address "Gdansk" "Harbor 2" "80-001" "7B"

  Scenario: A customer changes the shipping address in the Customer Profile
    Given customer "ship-henry" has saved customer profile with credentials "Henry Buyer" "henry@example.com" "+48100100105" and billing details "6234567890" "Henry Shop" "Krakow" "Stone 12" "30-006" "11C"
    And customer "ship-henry" has a shipping address "Katowice" "Coal 4" "40-001" "3D"
    When customer "ship-henry" adds shipping address "Wroclaw" "Bridge 6" "50-002" "14E"
    Then the customer command succeeds
    And the customer profile for "ship-henry" contains shipping address "Wroclaw" "Bridge 6" "50-002" "14E"

  Scenario: A missing customer cannot add a shipping address
    Given customer "ship-ivan" does not exist yet
    When customer "ship-ivan" adds shipping address "Szczecin" "Port 1" "70-003" "5F"
    Then the customer service rejects the request because the customer is missing
