package main

import (
	"database/sql"
	"log"
)

func SearchCapacity(db *sql.DB) []Locations {
	var locations []Locations

	productDB, err := db.Query("SELECT * FROM locations")
	if err != nil {
		log.Println("Error querying locations:", err)
		return nil
	}
	defer productDB.Close()

	for productDB.Next() {
		var location Locations
		err := productDB.Scan(&location.Id, &location.Location, &location.Capacity, &location.Process)
		if err != nil {
			log.Println("Error scanning location:", err)
			continue
		}
		locations = append(locations, location)
	}

	return locations
}

func DBUpdate(postId int, amount int, location_id int, size string, db *sql.DB) {

	var size_id int

	err := db.QueryRow("SELECT id FROM sizes WHERE size=?", size).Scan(&size_id)
	if err != nil {
		panic(err.Error())
	}

	updateQuery, err := db.Prepare("UPDATE stocks SET quantity=? WHERE product_id=? AND location_id=? AND size_id=?")
	if err != nil {
		panic(err.Error())
	}
	updateQuery.Exec(amount, postId, location_id, size_id)
}

func StockInfo(product_id int, size string, location_id int, db *sql.DB) int {
	var size_id int
	var quantity int

	err := db.QueryRow("SELECT id FROM sizes WHERE size=?", size).Scan(&size_id)
	if err != nil {
		panic(err.Error())
	}

	err = db.QueryRow("SELECT quantity FROM stocks WHERE product_id=? AND size_id=? AND location_id=?", product_id, size_id, location_id).Scan(&quantity)
	if err != nil {
		panic(err.Error())
	}

	return quantity
}

func SearchLocations(productID int, size string, amount int, db *sql.DB) []int {
	var locationIDs []int

	var sizeID int
	err := db.QueryRow("SELECT id FROM sizes WHERE size=?", size).Scan(&sizeID)
	if err != nil {
		log.Println("Error querying size ID:", err)
		return nil
	}

	locationDB, err := db.Query("SELECT location_id, quantity FROM stocks WHERE product_id=? AND size_id=?", productID, sizeID)
	if err != nil {
		log.Println("Error querying stock locations:", err)
		return nil
	}
	defer locationDB.Close()

	for locationDB.Next() {
		var locationID, quantity int
		err := locationDB.Scan(&locationID, &quantity)
		if err != nil {
			log.Println("Error scanning stock locations:", err)
			continue
		}
		if quantity >= amount {
			if !containsInt(locationIDs, locationID) {
				locationIDs = append(locationIDs, locationID)
			}
		}
	}

	return locationIDs
}

func SearchCargoInfo(db *sql.DB) []CargoInfo {
	var cargoInfos []CargoInfo

	cargoDB, err := db.Query("SELECT * FROM orderpriceinfos")
	if err != nil {
		log.Println("Error querying cargo info:", err)
		return nil
	}
	defer cargoDB.Close()

	for cargoDB.Next() {
		var cargoInfo CargoInfo
		var id int
		err := cargoDB.Scan(&id, &cargoInfo.Location_id, &cargoInfo.Cargo_id, &cargoInfo.Price_per_distance, &cargoInfo.Discount_per_piece, &cargoInfo.Order_price)
		if err != nil {
			log.Println("Error scanning cargo info:", err)
			continue
		}
		cargoInfos = append(cargoInfos, cargoInfo)
	}
	return cargoInfos
}

func SearchDistance(city_name string, db *sql.DB) []Distance {
	var distances []Distance

	var distance Distance
	var cityID int

	err := db.QueryRow("SELECT id FROM iller WHERE il_adi=?", city_name).Scan(&cityID)
	if err != nil {
		log.Println("Error querying city ID:", err)
		return nil
	}

	distanceDB, err := db.Query("SELECT location_id, distance, distance_key FROM distances WHERE il_id=?", cityID)
	if err != nil {
		log.Println("Error querying distances:", err)
		return nil
	}
	defer distanceDB.Close()

	for distanceDB.Next() {
		err := distanceDB.Scan(&distance.Location, &distance.Distance, &distance.DistanceKey)
		if err != nil {
			log.Println("Error scanning distance:", err)
			continue
		}
		distance.City = cityID
		distances = append(distances, distance)
	}

	return distances
}
