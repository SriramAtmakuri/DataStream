package generator

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"go-producer/internal/kafka"
	"go-producer/internal/metrics"
	"go-producer/internal/models"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
)

type Generator struct {
	producer    *kafka.Producer
	faker       *gofakeit.Faker
	ordersPerSec int
	topicOrders string
	topicDLQ    string
}

func New(producer *kafka.Producer) *Generator {
	rate := 5
	if v := os.Getenv("ORDERS_PER_SECOND"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			rate = n
		}
	}
	return &Generator{
		producer:    producer,
		faker:       gofakeit.New(0),
		ordersPerSec: rate,
		topicOrders: getEnv("KAFKA_TOPIC_ORDERS", "orders"),
		topicDLQ:    getEnv("KAFKA_TOPIC_DLQ", "orders-dlq"),
	}
}

func (g *Generator) Start(ctx context.Context) {
	ticker := time.NewTicker(time.Second / time.Duration(g.ordersPerSec))
	defer ticker.Stop()

	log.Printf("Order generator started: %d orders/sec", g.ordersPerSec)

	for {
		select {
		case <-ctx.Done():
			log.Println("Order generator stopped")
			return
		case <-ticker.C:
			g.emitOrder(ctx)
		}
	}
}

func (g *Generator) emitOrder(ctx context.Context) {
	order := g.generateOrder()

	payload, err := json.Marshal(order)
	if err != nil {
		log.Printf("marshal order: %v", err)
		return
	}

	// ~5% of orders are malformed — simulate by sending to DLQ directly
	if rand.Float64() < 0.05 {
		order.Status = models.StatusFailed
		payload, err = json.Marshal(order)
		if err != nil {
			log.Printf("marshal DLQ order: %v", err)
			return
		}
		if err := g.producer.Publish(ctx, g.topicDLQ, order.ID, payload); err != nil {
			log.Printf("publish to DLQ: %v", err)
		}
		metrics.OrdersFailedTotal.Inc()
		metrics.OrdersDLQTotal.Inc()
		return
	}

	if err := g.producer.Publish(ctx, g.topicOrders, order.ID, payload); err != nil {
		log.Printf("publish order: %v", err)
		metrics.OrdersFailedTotal.Inc()
		return
	}

	metrics.OrdersPublishedTotal.Inc()
	metrics.OrdersRevenue.Add(order.TotalAmount)
}

func (g *Generator) generateOrder() models.Order {
	category := randomElement(models.Categories)
	products := models.Products[category]
	product := randomElement(products)
	region := randomElement(models.Regions)

	quantity := rand.Intn(5) + 1
	price := roundFloat(50+rand.Float64()*950, 2)
	total := roundFloat(price*float64(quantity), 2)

	// Simulate out-of-order events: occasionally backdate the event by up to 30s
	eventTime := time.Now()
	if rand.Float64() < 0.1 {
		eventTime = eventTime.Add(-time.Duration(rand.Intn(30)) * time.Second)
	}

	status := models.StatusCompleted
	if rand.Float64() < 0.08 {
		status = models.StatusFailed
	} else if rand.Float64() < 0.05 {
		status = models.StatusPending
	}

	return models.Order{
		ID:          uuid.New().String(),
		CustomerID:  uuid.New().String(),
		Product:     product,
		Category:    category,
		Quantity:    quantity,
		Price:       price,
		TotalAmount: total,
		Region:      region,
		Status:      status,
		Timestamp:   eventTime,
		EventTime:   eventTime.UnixMilli(),
	}
}

func randomElement(s []string) string {
	return s[rand.Intn(len(s))]
}

func roundFloat(val float64, precision int) float64 {
	pow := 1.0
	for i := 0; i < precision; i++ {
		pow *= 10
	}
	return float64(int(val*pow+0.5)) / pow
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
