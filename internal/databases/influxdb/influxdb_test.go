//go:build integration
// +build integration

package influxdb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
	shared "github.com/tjhowse/aus_grocery_price_database/internal/shared"
)

func dir(envFile string) string {
	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			break
		}

		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			panic(fmt.Errorf("go.mod not found"))
		}
		currentDir = parent
	}

	return filepath.Join(currentDir, envFile)
}

func compareSystemStatusDatapoint(a shared.SystemStatusDatapoint, b map[string]any) bool {
	// this may be incorrect
	return a.HDDBytesFree == b["HDDBytesFree"] && a.ProductsPerSecond == b["ProductsPerSecond"] && a.RAMUtilisationPercent == b["RamUtilisationPerSecond"] && a.TotalProductCount == b["TotalProductCount"]
}

func compareProductInfo(a shared.ProductInfo, b map[string]any) bool {
	// do type conversion
	cents, ok := b["cents"].(int64)
	var centsMatch bool = (ok && a.PriceCents != int(cents)) || (!ok && a.PriceCents == 0 && b["cents"] == nil)

	grams, ok := b["grams"].(int64)
	var gramsMatch bool = (ok && a.WeightGrams != int(grams)) || (!ok && a.WeightGrams == 0 && b["grams"] == nil)

	bTime, ok := b["time"].(time.Time)
	if !ok {
		// the value returned was not time convertible
		if a.Timestamp.Equal(time.Time{}) && b["time"] == nil {
			// ok
		} else {
			return false
		}
	} else {
		// it is time convertible
		if !a.Timestamp.Equal(bTime) {
			return false
		}
	}

	return (a.ID == b["id"]) || (a.ID == "" && b["id"] == nil) && (a.Name == b["name"]) || (a.Name == "" && b["name"] == nil) && (a.Store == b["store"]) || (a.Store == "" && b["store"] == nil) && (a.Department == b["department"]) || (a.Department == "" && b["department"] == nil) && (a.Location == b["location"]) || (a.Location == "" && b["location"] == nil)

}

func compareArbitrarySystemStatusDatapoint(field string, value interface{}, b map[string]any) bool {
	return b[field] == value
}
func TestWriteProductDatapoint(t *testing.T) {
	i := InfluxDB{}
	// read in .env.test
	godotenv.Load(dir(".env.test"))
	influxDBUrl, influxDBToken, influxDBDatabase, influxDBProductTable := os.Getenv("INFLUXDB_URL"), os.Getenv("INFLUXDB_TOKEN"), os.Getenv("INFLUXDB_DATABASE"), os.Getenv("INFLUXDB_PRODUCT_TABLE")
	err := i.Init(influxDBUrl, influxDBToken, influxDBDatabase) // have to make an influxdb3 instance for testing
	/*
		install influxdb3 needs to be done manually
		./influxdb3 serve
			--node-id=node0 \
			--cluster-id=cluster0 \
			--object-store=file \
			--data-dir=./data
		install the cli manually
		use the cli to healthcheck

		init a database (perhaps needing to tear down an existing database from previous test run)
	*/

	if err != nil {
		t.Errorf("Could not init database client: %s", err.Error())
		t.FailNow()
	}
	desiredTags := map[string]string{
		"name":       "Test Product",
		"store":      "Test Store",
		"location":   "Test Location",
		"department": "Test Department",
	}

	// get the current time
	preWriteTime := time.Now().Format(time.RFC3339Nano)

	// make the input product points
	inputPoints := make([]shared.ProductInfo, 3)
	inputPoints[0] = shared.ProductInfo{
		Name:               desiredTags["name"],
		Store:              desiredTags["store"],
		Location:           desiredTags["location"],
		Department:         desiredTags["department"],
		PriceCents:         100,
		PreviousPriceCents: 0,
		WeightGrams:        1000,
		Timestamp:          time.Now(),
	}
	inputPoints[1] = shared.ProductInfo{
		Name:               desiredTags["name"],
		Store:              desiredTags["store"],
		Location:           desiredTags["location"],
		Department:         desiredTags["department"],
		PriceCents:         101,
		PreviousPriceCents: 100,
		WeightGrams:        1000,
		Timestamp:          time.Now(),
	}
	inputPoints[2] = shared.ProductInfo{
		Name:               desiredTags["name"],
		Store:              desiredTags["store"],
		Location:           desiredTags["location"],
		Department:         desiredTags["department"],
		PriceCents:         99,
		PreviousPriceCents: 101,
		WeightGrams:        1000,
		Timestamp:          time.Now(),
	}
	// write the input product points
	for _, v := range inputPoints {
		i.WriteProductDatapoint(v)
	}

	// sanity testing: check that only the measurements we wrote exist after preWriteTime (cardinality)
	ctx := context.Background()
	query := fmt.Sprintf("SELECT * FROM %s WHERE time >= TIMESTAMP '%s' ORDER BY time;", influxDBProductTable, preWriteTime)
	iterator, err := i.db.Query(ctx, query) // not using public interface of InfluxDB, this is not black box testing but in order to keep the interface of the package consistent it's acceptable
	if err != nil {
		t.Errorf("couldn't get query: %s", err.Error())
		t.FailNow()
	}
	var it int = 0
	for iterator.Next() {
		// compare the values to what we wrote and ensure that only 3 exist
		result := iterator.Value()
		if it > 2 {
			t.Errorf("cardinality of query didn't match what was expected: %d", it)
		}
		if !compareProductInfo(inputPoints[it], result) {
			t.Errorf("data points don't match what was expected")
			t.FailNow()
		}
		it++
	}
	if it != 3 {
		t.Errorf("cardinality didn't match what was expected: %d", it)
		t.FailNow()
	}

}

// func TestWriteSystemDatapoint(t *testing.T) {
// 	i := InfluxDB{}
// 	// read in .env.test
// 	godotenv.Load(dir(".env.test"))
// 	influxDBUrl, influxDBToken, influxDBDatabase, influxDBSystemTable := os.Getenv("INFLUXDB_URL"), os.Getenv("INFLUXDB_TOKEN"), os.Getenv("INFLUXDB_DATABASE"), os.Getenv("INFLUXDB_SYSTEM_TABLE")
// 	err := i.Init(influxDBUrl, influxDBToken, influxDBDatabase) // have to make an influxdb3 instance for testing
// 	/*
// 		install influxdb3 needs to be done manually
// 		./influxdb3 serve
// 			--node-id=node0 \
// 			--cluster-id=cluster0 \
// 			--object-store=file \
// 			--data-dir=./data
// 		install the cli manually
// 		use the cli to healthcheck
//
// 		init a database (perhaps needing to tear down an existing database from previous test run)
// 	*/
//
// 	if err != nil {
// 		t.Errorf("Could not init database client: %s", err.Error())
// 		t.FailNow()
// 	}
//
// 	// get the current time
// 	preWriteTime := time.Now().String()
//
// 	// make the input system status points
// 	inputPoints := make([]shared.SystemStatusDatapoint, 3)
// 	inputPoints[0] = shared.SystemStatusDatapoint{
// 		RAMUtilisationPercent: 35.3,
// 		ProductsPerSecond:     0.05,
// 		HDDBytesFree:          12,
// 		TotalProductCount:     12,
// 	}
// 	inputPoints[1] = shared.SystemStatusDatapoint{
// 		RAMUtilisationPercent: 37.30,
// 		ProductsPerSecond:     0.15,
// 		HDDBytesFree:          10,
// 		TotalProductCount:     8,
// 	}
// 	inputPoints[2] = shared.SystemStatusDatapoint{
// 		RAMUtilisationPercent: 12.15,
// 		ProductsPerSecond:     9.63,
// 		HDDBytesFree:          10,
// 		TotalProductCount:     8,
// 	}
// 	// write the input product points
// 	for _, v := range inputPoints {
// 		i.WriteSystemDatapoint(v)
// 	}
//
// 	// sanity testing: check that only the measurements we wrote exist after preWriteTime (cardinality)
// 	ctx := context.Background()
// 	query := fmt.Sprintf("SELECT * FROM %s WHERE time >= TIMESTAMP '%s' ORDER BY time;", influxDBSystemTable, preWriteTime)
// 	iterator, err := i.db.Query(ctx, query) // not using public interface of InfluxDB, this is not black box testing but in order to keep the interface of the package consistent it's acceptable
// 	if err != nil {
// 		t.Errorf("couldn't get query: %s", err.Error())
// 		t.FailNow()
// 	}
// 	var it int = 0
// 	for iterator.Next() {
// 		// compare the values to what we wrote and ensure that only 3 exist
// 		result := iterator.Value()
// 		compareSystemStatusDatapoint(inputPoints[it], result)
// 		it++
// 	}
// 	if it != 3 {
// 		t.FailNow()
// 	}
//
// }
//
// func TestWriteArbitrarySystemDatapoint(t *testing.T) {
// 	i := InfluxDB{}
// 	// read in .env.test
// 	godotenv.Load(dir(".env.test"))
// 	influxDBUrl, influxDBToken, influxDBDatabase, influxDBSystemTable := os.Getenv("INFLUXDB_URL"), os.Getenv("INFLUXDB_TOKEN"), os.Getenv("INFLUXDB_DATABASE"), os.Getenv("INFLUXDB_SYSTEM_TABLE")
// 	err := i.Init(influxDBUrl, influxDBToken, influxDBDatabase) // have to make an influxdb3 instance for testing
// 	/*
// 		install influxdb3 needs to be done manually
// 		./influxdb3 serve
// 			--node-id=node0 \
// 			--cluster-id=cluster0 \
// 			--object-store=file \
// 			--data-dir=./data
// 		install the cli manually
// 		use the cli to healthcheck
//
// 		init a database (perhaps needing to tear down an existing database from previous test run)
// 	*/
//
// 	if err != nil {
// 		t.Errorf("Could not init database client: %s", err.Error())
// 		t.FailNow()
// 	}
//
// 	// get the current time
// 	preWriteTime := time.Now().String()
//
// 	// write the arbitrary system points
// 	i.WriteArbitrarySystemDatapoint("colour", "grey")
// 	i.WriteArbitrarySystemDatapoint("number", 42)
// 	i.WriteArbitrarySystemDatapoint("metres", 1.5)
//
// 	// sanity testing: check that only the measurements we wrote exist after preWriteTime (cardinality)
// 	ctx := context.Background()
// 	// idk how this is even queried, this links to the original question, are they all being written to the same table
// 	query := fmt.Sprintf("SELECT * FROM %s WHERE time >= TIMESTAMP '%s' ORDER BY time;", influxDBSystemTable, preWriteTime)
// 	iterator, err := i.db.Query(ctx, query) // not using public interface of InfluxDB, this is not black box testing but in order to keep the interface of the package consistent it's acceptable
// 	if err != nil {
// 		t.Errorf("couldn't get query: %s", err.Error())
// 		t.FailNow()
// 	}
// 	var it int = 0
// 	for iterator.Next() {
// 		// compare the values to what we wrote and ensure that only 3 exist
// 		result := iterator.Value()
// 		switch it {
// 		case 0:
// 			compareArbitrarySystemStatusDatapoint("colour", "grey", result)
// 		case 1:
// 			compareArbitrarySystemStatusDatapoint("number", 42, result)
// 		case 2:
// 			compareArbitrarySystemStatusDatapoint("metres", 1.5, result)
// 		}
// 		it++
// 	}
// 	if it != 3 {
// 		t.FailNow()
// 	}
//
// }
//
