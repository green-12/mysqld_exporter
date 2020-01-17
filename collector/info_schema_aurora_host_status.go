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

// Scrape `information_schema.replica_host_status`.

package collector

import (
	"context"
	"database/sql"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

const auroraHostStatQuery = `
		select
		  server_id,
		  cpu,
		  replica_lag_in_milliseconds as replica_lag
		from information_schema.replica_host_status
		where server_id = @@aurora_server_id
		`
// Metric descriptors.
var (
	infoSchemaAuroraCPUUsageDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, informationSchema, "aurora_status_cpu_usage"),
		"The cpu usage of aurora instance.",
		[]string{"server_id"}, nil,
	)
	infoSchemaAuroraReplicaLagDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, informationSchema, "aurora_status_replica_lag_ms"),
		"The mili-seconds of repica lag.",
		[]string{"server_id"}, nil,
	)
)

// ScrapeAuroraHostStatus collects from `information_schema.replica_host_status`.
type ScrapeAuroraHostStatus struct{}

// Name of the Scraper. Should be unique.
func (ScrapeAuroraHostStatus) Name() string {
	return "info_schema.aurora_stats"
}

// Help describes the role of the Scraper.
func (ScrapeAuroraHostStatus) Help() string {
	return "Only works on aurora instance"
}

// Version of MySQL from which scraper is available.
func (ScrapeAuroraHostStatus) Version() float64 {
	return 5.6
}

// Scrape collects data from database connection and sends it over channel as prometheus metric.
func (ScrapeAuroraHostStatus) Scrape(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric, logger log.Logger) error {

	informationSchemaReplicaHostStatusRows, err := db.QueryContext(ctx, auroraHostStatQuery)
	if err != nil {
		return err
	}
	defer informationSchemaReplicaHostStatusRows.Close()

	var (
		auroraServerID string
		cpu            float64
		replicaLag     float64
	)

	for informationSchemaReplicaHostStatusRows.Next() {
		err = informationSchemaReplicaHostStatusRows.Scan(
			&auroraServerID,
			&cpu,
			&replicaLag,
		)
		if err != nil {
			return err
		}
		ch <- prometheus.MustNewConstMetric(
			infoSchemaAuroraCPUUsageDesc, prometheus.GaugeValue, float64(cpu),
			auroraServerID,
		)
		ch <- prometheus.MustNewConstMetric(
			infoSchemaAuroraReplicaLagDesc, prometheus.GaugeValue, float64(replicaLag),
			auroraServerID,
		)
	}
	return nil
}

// check interface
var _ Scraper = ScrapeAuroraHostStatus{}
