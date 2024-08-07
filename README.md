## Allocator Algorithm

### Features

#### Cargo Information
- **Different City Rates**: Each city has different shipping rates.
- **Distance-based Charges**: Charges are based on distance per kilometer and increase with the number of items.
- **Fixed Order Fee**: A fixed fee is charged per order.
- **Quantity Discounts**: Discounts are applied as the number of items increases.

#### Distance Information
- **Distance Storage**: All distances between the warehouse where the products are located and the city requested by the user are stored in the Distances table.
- **Linked Locations**: This table is linked to locations, i.e., warehouses and cities.
- **Distance Keys**:
  - Less than 300 km: `distanceKey = 1`
  - Between 300 and 600 km: `distanceKey = 2`
  - Between 600 and 900 km: `distanceKey = 3`
  - Between 900 and 1200 km: `distanceKey = 4`
  - Between 1200 and 1500 km: `distanceKey = 5`
  - More than 1500 km: `distanceKey = 6`
- **Proximity Calculation**: If one of the locations is less than 300 km away from the desired city, the closest one is always advantageous, so other distances don't need to be calculated to save computation costs.
- **Close Distance Calculation**: If the closest location's distance is less than 300 km, both locations need to be calculated as the slight difference in distance may be compensated by different city rates.
- **Calculation Logic**: 
  - If the nearest warehouse has a `distanceKey` of 1, calculate for warehouses with `distanceKey` 1 and 2.
  - If the nearest warehouse has a `distanceKey` of 2, calculate for warehouses with `distanceKey` 2 and 3.
  - This approach minimizes errors and unnecessary calculations.

### Algorithm
- **Data Retrieval**: All cargo information, the address where the order is placed, and distances between all warehouses are retrieved from the database.
- **Data Filtering**: The retrieved data is filtered based on:
  - Warehouses with stock.
  - `distanceKey` separations.
  - Shipping information for the desired warehouses' cities.
- **Combination Generation**: All possible combinations of product and warehouse pairs are generated.
- **Cost Calculation**: The cost for each combination and warehouse capacity is calculated.
  - **Cost Formula**: `Cost = OrderPrice + (Distance * PricePerKm * Quantity) * ((100 - (Quantity - 1) * Discount) / 100)`
  - **Cost Score**: `Cost Score = 5 - (Cost / 200)` (The goal is to score out of 5, inversely proportional to cost)
  - **Capacity Score**: `Capacity Score = 5 - (Operations * 5 / Capacity)`
  - **Total Score**: `Total Score = (3 * Cost Score + Capacity Score) / 4`
  - The weight of the capacity score is 1, while the weight of the cost score is 3.
- **Optimal Path Selection**: The path with the highest score is selected and used for shipping.
