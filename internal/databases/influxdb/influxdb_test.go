//go:build integration
// +build integration

package influxdb

import (
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	shared "github.com/tjhowse/aus_grocery_price_database/internal/shared"
)

func TestWriteProductDatapoint(t *testing.T) {
	i := InfluxDB{}
	// read in .env.test
	godotenv.Load(".env.test")
	influxDBUrl, influxDBToken, influxDBDatabase := os.Getenv("URL"), os.Getenv("TOKEN"), os.Getenv("DATABASE")
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
		t.FailNow()
	}

	desiredTags := map[string]string{
		"name":       "Test Product",
		"store":      "Test Store",
		"location":   "Test Location",
		"department": "Test Department",
	}

	i.WriteProductDatapoint(shared.ProductInfo{
		Name:               desiredTags["name"],
		Store:              desiredTags["store"],
		Location:           desiredTags["location"],
		Department:         desiredTags["department"],
		PriceCents:         100,
		PreviousPriceCents: 0,
		WeightGrams:        1000,
		Timestamp:          time.Now(),
	})
	i.WriteProductDatapoint(shared.ProductInfo{
		Name:               desiredTags["name"],
		Store:              desiredTags["store"],
		Location:           desiredTags["location"],
		Department:         desiredTags["department"],
		PriceCents:         101,
		PreviousPriceCents: 100,
		WeightGrams:        1000,
		Timestamp:          time.Now(),
	})
	i.WriteProductDatapoint(shared.ProductInfo{
		Name:               desiredTags["name"],
		Store:              desiredTags["store"],
		Location:           desiredTags["location"],
		Department:         desiredTags["department"],
		PriceCents:         99,
		PreviousPriceCents: 101,
		WeightGrams:        1000,
		Timestamp:          time.Now(),
	})

	if want, got := 3, len(gMock.writtenPoints); want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	// Check the first written point.
	p := gMock.writtenPoints[0]
	if want, got := "product", p.Name(); want != got {
		t.Errorf("want %s, got %s", want, got)
	}

	for _, tag := range p.TagList() {
		if want, got := desiredTags[tag.Key], tag.Value; want != got {
			t.Errorf("want %s, got %s", want, got)
		}
	}

	for _, field := range p.FieldList() {
		switch field.Key {
		case "cents":
			if want, got := int64(100), field.Value.(int64); want != got {
				t.Errorf("want %v, got %v", want, got)
			}
		case "grams":
			if want, got := int64(1000), field.Value.(int64); want != got {
				t.Errorf("want %v, got %v", want, got)
			}
		case "cents_change":
			t.Errorf("unexpected field %s", field.Key)
		default:
			t.Errorf("unexpected field %s", field.Key)
		}
	}

	// Now check the second written point.
	p = gMock.writtenPoints[1]

	for _, field := range p.FieldList() {
		switch field.Key {
		case "cents":
			if want, got := int64(101), field.Value.(int64); want != got {
				t.Errorf("want %v, got %v", want, got)
			}
		case "grams":
			continue
		case "cents_change":
			if want, got := int64(1), field.Value.(int64); want != got {
				t.Errorf("want %v, got %v", want, got)
			}
		default:
			t.Errorf("unexpected field %s", field.Key)
		}
	}

	// Now check the third written point.
	p = gMock.writtenPoints[2]

	for _, field := range p.FieldList() {
		switch field.Key {
		case "cents":
			continue
		case "grams":
			continue
		case "cents_change":
			if want, got := int64(-2), field.Value.(int64); want != got {
				t.Errorf("want %v, got %v", want, got)
			}
		default:
			t.Errorf("unexpected field %s", field.Key)
		}
	}

}
