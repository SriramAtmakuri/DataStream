package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go-producer/internal/clickhouse"
	kfk "go-producer/internal/kafka"
	"go-producer/internal/metrics"
	"go-producer/internal/models"
)

type Handler struct {
	producer *kfk.Producer
	ch       *clickhouse.Client
	topicOrders string
	topicDLQ    string
}

func New(producer *kfk.Producer) *Handler {
	return &Handler{
		producer:    producer,
		ch:          clickhouse.NewClient(),
		topicOrders: getEnv("KAFKA_TOPIC_ORDERS", "orders"),
		topicDLQ:    getEnv("KAFKA_TOPIC_DLQ", "orders-dlq"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if order.ID == "" {
		order.ID = uuid.New().String()
	}
	if order.CustomerID == "" {
		order.CustomerID = uuid.New().String()
	}
	order.Timestamp = time.Now()
	order.EventTime = order.Timestamp.UnixMilli()
	if order.Status == "" {
		order.Status = models.StatusCompleted
	}
	order.TotalAmount = order.Price * float64(order.Quantity)

	payload, err := json.Marshal(order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "serialization failed"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.producer.Publish(ctx, h.topicOrders, order.ID, payload); err != nil {
		metrics.OrdersFailedTotal.Inc()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish order"})
		return
	}

	metrics.OrdersPublishedTotal.Inc()
	metrics.OrdersRevenue.Add(order.TotalAmount)
	c.JSON(http.StatusCreated, order)
}

func (h *Handler) GetStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "running",
		"timestamp": time.Now().UTC(),
	})
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	})
}

func (h *Handler) GetOrdersPerMinute(c *gin.Context) {
	rows, err := h.ch.Query(`
		SELECT
			toStartOfMinute(minute) AS minute,
			order_count,
			total_revenue,
			failed_count
		FROM ecommerce.orders_per_minute
		WHERE minute >= now() - INTERVAL 60 MINUTE
		ORDER BY minute ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, rows)
}

func (h *Handler) GetRevenueByRegion(c *gin.Context) {
	rows, err := h.ch.Query(`
		SELECT
			region,
			sum(total_revenue) AS revenue,
			sum(order_count)   AS orders
		FROM ecommerce.revenue_by_region
		WHERE window_start >= now() - INTERVAL 1 HOUR
		GROUP BY region
		ORDER BY revenue DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, rows)
}

func (h *Handler) GetTopProducts(c *gin.Context) {
	rows, err := h.ch.Query(`
		SELECT
			product,
			category,
			sum(quantity_sold) AS quantity,
			sum(total_revenue) AS revenue
		FROM ecommerce.top_products
		WHERE window_start >= now() - INTERVAL 1 HOUR
		GROUP BY product, category
		ORDER BY revenue DESC
		LIMIT 10
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rows == nil {
		rows = []map[string]interface{}{}
	}
	c.JSON(http.StatusOK, rows)
}

func (h *Handler) GetErrorRate(c *gin.Context) {
	rows, err := h.ch.Query(`
		SELECT
			count(*)                                        AS total,
			countIf(status = 'failed')                      AS failed,
			round(countIf(status = 'failed') * 100.0 / count(*), 2) AS error_rate
		FROM ecommerce.orders
		WHERE timestamp >= now() - INTERVAL 5 MINUTE
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(rows) == 0 {
		c.JSON(http.StatusOK, gin.H{"total": 0, "failed": 0, "error_rate": 0})
		return
	}
	c.JSON(http.StatusOK, rows[0])
}
