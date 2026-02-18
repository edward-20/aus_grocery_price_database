package influxdb

import (
	"log/slog"

	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
	shared "github.com/tjhowse/aus_grocery_price_database/internal/shared"
)

type InfluxDB struct {
	db           *influxdb3.Client
	productTable string
	systemTable  string
}

func (i *InfluxDB) Init(url, token, database string) error {
	slog.Info("Initialising InfluxDB", "url", url, "database", database)
	client, err := influxdb3.New(influxdb3.ClientConfig{
		Host:     url,
		Token:    token,
		Database: database,
	})
	if err != nil {
		// handle error
		return err
	}
	i.db = client
	return nil
}

func (i *InfluxDB) WriteProductDatapoint(info shared.ProductInfo) {
	/*
		(shared.ProductInfo) -> in influxdb we will have:
			fields:
				"cents"
				"grams"
				"cents_change"
			tags:
				"id"
				"name"
				"store"
				"location"
				"department"
			timestamp
	*/
}

func (i *InfluxDB) WriteArbitrarySystemDatapoint(field string, value interface{}) {
	/*
		(field, value) -> in influxdb we will have:
			fields:
				"field": value
			tags:
				"service": shared.SYSTEM_SERVICE_NAME
			timestamp
	*/
}

func (i *InfluxDB) WriteSystemDatapoint(data shared.SystemStatusDatapoint) {
	/*
		(shared.SystemStatusDatapoint) -> in influxdb we will have:
			fields:
				shared.SYSTEM_RAM_UTILISATION_PERCENT_FIELD: data.RAMUtilisationPercent,
				shared.SYSTEM_PRODUCTS_PER_SECOND_FIELD:     data.ProductsPerSecond,
				shared.SYSTEM_HDD_BYTES_FREE_FIELD:          data.HDDBytesFree,
				shared.SYSTEM_TOTAL_PRODUCT_COUNT_FIELD:     data.TotalProductCount,
				"grams"
				"cents_change"
			timestamp
	*/
}

// WriteWorker writes ProductInfo to InfluxDB
// Note that the underlying library automatically batches writes
// so we don't need to worry about that here.
func (i *InfluxDB) WriteWorker(input <-chan shared.ProductInfo) {
	// for info := range input {
	// 	i.WriteProductDatapoint(info)
	// }
}

func (i *InfluxDB) Close() {
	// i.groceryWriteAPI.Flush()
	// i.systemWriteAPI.Flush()
	// i.db.Close()
}
