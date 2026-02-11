package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Client handles RabbitMQ operations
type Client struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	exchange  string
	confirms  chan amqp.Confirmation
	logger    *zap.Logger
	mu        sync.RWMutex
	connected bool
}

// Config for RabbitMQ client
type Config struct {
	Host              string
	Port              int
	User              string
	Password          string
	Exchange          string
	PublisherConfirms bool
}

// NewClient creates a new RabbitMQ client
func NewClient(cfg Config, logger *zap.Logger) (*Client, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}
	
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("open channel: %w", err)
	}
	
	// Declare exchange
	err = channel.ExchangeDeclare(
		cfg.Exchange, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}
	
	client := &Client{
		conn:      conn,
		channel:   channel,
		exchange:  cfg.Exchange,
		logger:    logger,
		connected: true,
	}
	
	// Enable publisher confirms if requested
	if cfg.PublisherConfirms {
		if err := channel.Confirm(false); err != nil {
			client.Close()
			return nil, fmt.Errorf("enable confirms: %w", err)
		}
		client.confirms = channel.NotifyPublish(make(chan amqp.Confirmation, 100))
	}
	
	// Monitor connection
	go client.monitorConnection()
	
	return client, nil
}

// Close closes the RabbitMQ connection
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.connected = false
	
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Publish publishes a message to the exchange
func (c *Client) Publish(ctx context.Context, routingKey string, payload interface{}) error {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return fmt.Errorf("not connected to rabbitmq")
	}
	c.mu.RUnlock()
	
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}
	
	// Extract trace context from ctx if available
	// traceID := extractTraceID(ctx)
	
	msg := amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		Timestamp:    time.Now(),
		DeliveryMode: amqp.Persistent, // Persistent messages
	}
	
	// Publish with context
	err = c.channel.PublishWithContext(
		ctx,
		c.exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		msg,
	)
	
	if err != nil {
		return fmt.Errorf("publish message: %w", err)
	}
	
	// Wait for publisher confirmation if enabled
	if c.confirms != nil {
		select {
		case confirm := <-c.confirms:
			if !confirm.Ack {
				return fmt.Errorf("message not acknowledged by broker")
			}
		case <-time.After(5 * time.Second):
			return fmt.Errorf("confirmation timeout")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	
	c.logger.Debug("published message",
		zap.String("routing_key", routingKey),
		zap.Int("body_size", len(body)))
	
	return nil
}

// DeclareQueue declares a queue and binds it to the exchange
func (c *Client) DeclareQueue(name string, bindingKeys []string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if !c.connected {
		return fmt.Errorf("not connected to rabbitmq")
	}
	
	// Declare queue
	_, err := c.channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-dead-letter-exchange": c.exchange + ".dlx",
		},
	)
	if err != nil {
		return fmt.Errorf("declare queue: %w", err)
	}
	
	// Bind to exchange with routing keys
	for _, key := range bindingKeys {
		err = c.channel.QueueBind(
			name,       // queue name
			key,        // routing key
			c.exchange, // exchange
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("bind queue: %w", err)
		}
	}
	
	c.logger.Info("declared queue",
		zap.String("queue", name),
		zap.Strings("binding_keys", bindingKeys))
	
	return nil
}

// Consume starts consuming messages from a queue
func (c *Client) Consume(queueName string, handler func([]byte) error) error {
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return fmt.Errorf("not connected to rabbitmq")
	}
	c.mu.RUnlock()
	
	// Set QoS
	err := c.channel.Qos(
		20,    // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("set qos: %w", err)
	}
	
	msgs, err := c.channel.Consume(
		queueName,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("start consuming: %w", err)
	}
	
	go func() {
		for msg := range msgs {
			err := handler(msg.Body)
			if err != nil {
				c.logger.Error("handler error",
					zap.Error(err),
					zap.String("queue", queueName))
				msg.Nack(false, true) // Requeue on error
			} else {
				msg.Ack(false)
			}
		}
	}()
	
	c.logger.Info("started consuming", zap.String("queue", queueName))
	return nil
}

// monitorConnection watches for connection issues
func (c *Client) monitorConnection() {
	closeChan := make(chan *amqp.Error)
	c.conn.NotifyClose(closeChan)
	
	err := <-closeChan
	if err != nil {
		c.logger.Error("connection closed", zap.Error(err))
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
	}
}

// IsConnected returns connection status
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}
