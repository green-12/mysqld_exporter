// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Scrape `information_schema.innodb_trx`.

package collector

import (
	"context"
	"database/sql"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const innodbTrxQuery = `
	select /* */
		ifnull(sum(case when TIMEDIFF(now(),trx_started) >= '00:00:05' then 1 else 0 end ),0) as "5_sec_count",
		ifnull(sum(case when TIMEDIFF(now(),trx_started) >= '00:00:30' then 1 else 0  end ),0) as "30_sec_count",
		ifnull(sum(case when TIMEDIFF(now(),trx_started) >= '00:01:00' then 1 else 0  end ),0) as "60_sec_count"
	from information_schema.innodb_trx trx
	`

// Metric descriptors.
var (
	infoSchemaTrxCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, informationSchema, "trx_count_per_sec"),
		"Number of transactions performed over (period) seconds.",
		[]string{"period"}, nil,
	)
)

// ScrapeInnodbTrx collects from `information_schema.innodb_trx`.
type ScrapeInnodbTrx struct{}

// Name of the Scraper. Should be unique.
func (ScrapeInnodbTrx) Name() string {
	return informationSchema + ".innodb_trx"
}

// Help describes the role of the Scraper.
func (ScrapeInnodbTrx) Help() string {
	return "Collect metrics from information_schema.innodb_trx"
}

// Version of MySQL from which scraper is available.
func (ScrapeInnodbTrx) Version() float64 {
	return 5.6
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeInnodbTrx) Scrape(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric, logger log.Logger) error {

	informationSchemaInnodbTrxRows, err := db.QueryContext(ctx, innodbTrxQuery)

	if err != nil {
		return err
	}
	defer informationSchemaInnodbTrxRows.Close()

	var (
		period5       string
		period30      string
		period60      string
		trx5SecCount  uint64
		trx30SecCount uint64
		trx60SecCount uint64
	)
	period5 = "5"
	period30 = "30"
	period60 = "60"

	for informationSchemaInnodbTrxRows.Next() {

		err = informationSchemaInnodbTrxRows.Scan(
			&trx5SecCount,
			&trx30SecCount,
			&trx60SecCount,
		)
		if err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			infoSchemaTrxCountDesc, prometheus.GaugeValue, float64(trx5SecCount), period5,
		)
		ch <- prometheus.MustNewConstMetric(
			infoSchemaTrxCountDesc, prometheus.GaugeValue, float64(trx30SecCount), period30,
		)
		ch <- prometheus.MustNewConstMetric(
			infoSchemaTrxCountDesc, prometheus.GaugeValue, float64(trx60SecCount), period60,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeInnodbTrx{}
