package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	OrdersPublishedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orders_published_total",
		Help: "Total number of orders successfully published to Kafka",
	})

	OrdersFailedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orders_failed_total",
		Help: "Total number of orders that failed to publish",
	})

	OrdersDLQTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orders_dlq_total",
		Help: "Total number of orders sent to dead letter queue",
	})

	OrdersRevenue = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orders_revenue_total",
		Help: "Cumulative revenue from all published orders",
	})
)
