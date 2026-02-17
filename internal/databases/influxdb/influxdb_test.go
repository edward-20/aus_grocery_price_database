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

func compare(a shared.ProductInfo, b map[string]any) bool {
	// this may be incorrect
	return a.ID == b["ID"] && a.Name == b["Name"] && a.Description == b["Description"] && a.Store == b["Store"] && a.Department == b["Department"] && a.Location == b["Location"] && a.PriceCents == b["PriceCents"] && a.WeightGrams == b["WeightGrams"] && a.Timestamp == b["time"]
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

	// create a new measurement(table) [that needs to be taken at the end of the test]
	// fooTags := map[string]string{
	// 	"name":       "Test Product",
	// 	"store":      "Test Store",
	// 	"location":   "Test Location",
	// 	"department": "Test Department",
	// }
	// fooMeasurement := map[string]interface{}{
	// 	"price": 5,
	// }
	// influxdb3.NewPoint(
	// 	"test",
	// 	fooTags,
	// 	fooMeasurement,
	// 	time.Now(),
	// )

	// get the current time
	preWriteTime := time.Now().String()

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
		compare(inputPoints[it], result)
		it++
	}
	if it != 3 {
		t.FailNow()
	}

	// for _, tag := range p.TagList() {
	// 	if want, got := desiredTags[tag.Key], tag.Value; want != got {
	// 		t.Errorf("want %s, got %s", want, got)
	// 	}
	// }

	// for _, field := range p.FieldList() {
	// 	switch field.Key {
	// 	case "cents":
	// 		if want, got := int64(100), field.Value.(int64); want != got {
	// 			t.Errorf("want %v, got %v", want, got)
	// 		}
	// 	case "grams":
	// 		if want, got := int64(1000), field.Value.(int64); want != got {
	// 			t.Errorf("want %v, got %v", want, got)
	// 		}
	// 	case "cents_change":
	// 		t.Errorf("unexpected field %s", field.Key)
	// 	default:
	// 		t.Errorf("unexpected field %s", field.Key)
	// 	}
	// }

	// // Now check the second written point.
	// p = gMock.writtenPoints[1]

	// for _, field := range p.FieldList() {
	// 	switch field.Key {
	// 	case "cents":
	// 		if want, got := int64(101), field.Value.(int64); want != got {
	// 			t.Errorf("want %v, got %v", want, got)
	// 		}
	// 	case "grams":
	// 		continue
	// 	case "cents_change":
	// 		if want, got := int64(1), field.Value.(int64); want != got {
	// 			t.Errorf("want %v, got %v", want, got)
	// 		}
	// 	default:
	// 		t.Errorf("unexpected field %s", field.Key)
	// 	}
	// }

	// // Now check the third written point.
	// p = gMock.writtenPoints[2]

	// for _, field := range p.FieldList() {
	// 	switch field.Key {
	// 	case "cents":
	// 		continue
	// 	case "grams":
	// 		continue
	// 	case "cents_change":
	// 		if want, got := int64(-2), field.Value.(int64); want != got {
	// 			t.Errorf("want %v, got %v", want, got)
	// 		}
	// 	default:
	// 		t.Errorf("unexpected field %s", field.Key)
	// 	}
	// }

}
