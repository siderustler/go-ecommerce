Feature: Product service pagination
  In order to return the correct page of products to the caller
  As the product service
  I need to convert page-based queries into repository offsets
  So that the repository fetches the correct slice of catalog data

  Scenario: Products uses zero offset for the first page
    Given a products query with page 1 and limit 12
    When the product service prepares the repository request
    Then the repository offset should be 0

  Scenario: Products calculates offset for later pages
    Given a products query with page 4 and limit 12
    When the product service prepares the repository request
    Then the repository offset should be 36

  Scenario: Promotions normalizes page zero to the first page
    Given a promotions query with page 0 and page size 8
    When the product service prepares the repository request
    Then the page should be assigned to 1
    Then the repository offset should be 0

  Scenario: Promotions normalizes negative pages to the first page
    Given a promotions query with page -3 and page size 8
    When the product service prepares the repository request
    Then the page should be assigned to 1
    Then the repository offset should be 0

  Scenario: Promotions calculates offset for later pages
    Given a promotions query with page 5 and page size 8
    When the product service prepares the repository request
    Then the repository offset should be 32
