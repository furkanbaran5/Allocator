package main

import (
	"database/sql"
	"math"
	"strconv"
)

func Allocator(Order Order, db *sql.DB) Info {
	allCombinations := [][]int{}
	var Info Info

	matrixDistances, locations, Ids, CargoInfos, Id := CalculateProductStock(Order, db)
	if Id != "" {
		Info.Ret = 0
		Info.Text = "Sipariş tamamlanamadı. Yeterli stok olmayan ürün: " + Id + " Lütfen ürün adedini düşürünüz."
		return Info
	}
	allCombinations = append(allCombinations, Ids)
	FindCombinations(locations, []int{}, 0, len(locations), &allCombinations)
	locationsData := SearchCapacity(db)
	BestOrder := CalculateAllCombinations(allCombinations, matrixDistances, CargoInfos, locationsData, len(locations))
	WriteInOrder(&Order, BestOrder, allCombinations)
	Info = updateStock(Order, db)

	return Info
}

// Tüm ürünlerin çıkabilecekleri depolar hesaplanıyor
func CalculateProductStock(Order Order, db *sql.DB) ([][]Distance, []LocationsForCombination, []int, []CargoInfo, string) {
	var locations []LocationsForCombination
	var ids []int
	matrixDistances := make([][]Distance, len(Order.Baskets))

	distances := SearchDistance(Order.Address.City, db)
	cargoInfos := SearchCargoInfo(db)

	for index, basketProduct := range Order.Baskets {
		amount, _ := strconv.Atoi(basketProduct.Amount)

		for i := 0; i < amount; i++ {
			id, _ := strconv.Atoi(basketProduct.Id)

			locationIDs := SearchLocations(id, basketProduct.Size, amount, db)

			if locationIDs == nil {
				return nil, nil, nil, nil, basketProduct.Id
			}

			distance := FilterDistances(distances, locationIDs)

			distance = FilterDistancesByLocations(distance)
			matrixDistances[index] = append(matrixDistances[index], distance...)

			locationIDs = locationIDs[:0]

			for _, dist := range distance {
				locationIDs = append(locationIDs, dist.Location)
			}

			locations = append(locations, LocationsForCombination{Location_ids: locationIDs, Product_id: id})

			ids = append(ids, id)
		}
	}
	return matrixDistances, locations, ids, cargoInfos, ""
}

func FilterDistances(distances []Distance, locationIDs []int) []Distance {
	var filtered []Distance
	locationMap := make(map[int]bool)

	for _, id := range locationIDs {
		locationMap[id] = true
	}

	for _, d := range distances {
		if locationMap[d.Location] {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func FilterDistancesByLocations(distances []Distance) []Distance {
	var filtered []Distance
	minKey := float32(6)

	for _, d := range distances {
		if d.DistanceKey < minKey {
			minKey = d.DistanceKey
		}
	}

	for _, d := range distances {
		if minKey == 1 && (d.DistanceKey == 1 || d.DistanceKey == 2) ||
			minKey == 2 && (d.DistanceKey == 2 || d.DistanceKey == 3) ||
			minKey == 3 && (d.DistanceKey == 3 || d.DistanceKey == 4) ||
			minKey == 4 && (d.DistanceKey == 4 || d.DistanceKey == 5) ||
			minKey == 5 && (d.DistanceKey == 5 || d.DistanceKey == 6) ||
			minKey == 6 && d.DistanceKey == 6 {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func FindCombinations(locations []LocationsForCombination, combination []int, start, k int, allCombinations *[][]int) {
	if k == 0 {
		comb := make([]int, len(combination))
		copy(comb, combination)
		*allCombinations = append(*allCombinations, comb)
		return
	}

	for i := start; i <= len(locations)-k; i++ {
		for _, id := range locations[i].Location_ids {
			FindCombinations(locations, append(combination, id), i+1, k-1, allCombinations)
		}
	}
}

func CalculateAllCombinations(allCombinations [][]int, matrixDistances [][]Distance, CargoInfos []CargoInfo, Locations []Locations, productAmount int) BestCombination {

	var bestCombination [6][]int
	var BestOrder BestCombination
	PackagesWay := make([]int, productAmount)
	BestOrder.Point = -5

	for i := 1; i < len(allCombinations); i++ {
		var sameArray [6][]int

		for j, locID := range allCombinations[i] {
			sameArray[locID] = append(sameArray[locID], j)
		}

		var targetDistance Distance
		var cost, capacityPoint, multiplier float32
		var totalWays []Way

		for index, arr := range sameArray {
			if len(arr) == 0 {
				continue
			}
			multiplier++

			for _, d := range matrixDistances {
				for _, distance := range d {
					if distance.Location == index {
						targetDistance = distance
						break
					}
				}
			}
			cargo := FilterCargosByLocationOfSingle(CargoInfos, targetDistance.Location)
			capacityPoint += CalculatePointToCapacity(Locations, targetDistance.Location)
			way := CalculatePointToCargo(targetDistance, cargo, float32(len(arr)))

			cost += way.Amount
			totalWays = append(totalWays, way)
		}

		capacityPoint /= multiplier
		costPoint := 5 - (cost / 200)
		totalPoint := ((3 * costPoint) + capacityPoint) / 4

		if BestOrder.Point < totalPoint {
			BestOrder.Point = totalPoint
			bestCombination = sameArray
			BestOrder.Ways = totalWays
		}

	}
	for index, values := range bestCombination {
		for _, value := range values {
			PackagesWay[value] = index
		}
	}
	BestOrder.Combination = PackagesWay

	return BestOrder
}

func FilterCargosByLocationOfSingle(cargoInfo []CargoInfo, location int) []CargoInfo {
	var filtered []CargoInfo
	for _, d := range cargoInfo {
		if d.Location_id == location {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func CalculatePointToCapacity(Locations []Locations, Location int) float32 {
	for _, L := range Locations {
		if L.Id == Location {
			return 5 - (L.Process * 5 / L.Capacity)
		}
	}
	return 0
}

func CalculatePointToCargo(Distance Distance, CargoInfo []CargoInfo, adet float32) Way {
	var cheapestWay Way
	cheapestWay.Amount = float32(math.MaxFloat32)

	for _, c := range CargoInfo {
		discount := (100 - ((adet - 1) * c.Discount_per_piece)) / 100
		amount := c.Order_price + (Distance.Distance * c.Price_per_distance * adet * discount)

		if amount < cheapestWay.Amount {
			cheapestWay.Amount = amount
			cheapestWay.Cargo_id = c.Cargo_id
			cheapestWay.Location_id = c.Location_id
		}
	}

	return cheapestWay
}

func WriteInOrder(Order *Order, BestOrder BestCombination, allCombinations [][]int) {
	for index, value := range BestOrder.Combination {
		for _, way := range BestOrder.Ways {
			if way.Location_id == value {
				for i := range Order.Baskets {
					Id, _ := strconv.Atoi(Order.Baskets[i].Id)
					if Id == allCombinations[0][index] {
						Order.Baskets[i].Cargo_id = way.Cargo_id
						Order.Baskets[i].Location_id = value
					}
				}
			}
		}
	}
}

func updateStock(Order Order, db *sql.DB) Info {
	for _, basket := range Order.Baskets {
		id, _ := strconv.Atoi(basket.Id)
		amount, _ := strconv.Atoi(basket.Amount)
		quantity := StockInfo(id, basket.Size, basket.Location_id, db)
		DBUpdate(id, quantity-amount, basket.Location_id, basket.Size, db)
	}

	return Info{Ret: 1, Text: "Sipariş tamamlandı"}
}

func containsInt(slice []int, item int) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
