package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/IBM/sarama"
)

// IMicroservice is interface for centralized service management
type IMicroservice interface {
	Start() error
	Stop()
	Cleanup() error
	Log(message string)

	// Consumer Services
	Consume(servers string, topic string, groupID string, h ServiceHandleFunc) error
}

type KafkaConsumer struct {
	consumerClient sarama.ConsumerGroup
	servers        string
	groupID        string
}

// Microservice is the centralized service management
type Microservice struct {
	exitChannel chan bool
	kafkaClient KafkaConsumer
}

// IContext is the context for service
type IContext interface {
	Log(message string)
	Param(name string) string
	Response(responseCode int, responseData interface{})
	ReadInput(any) error
}

// ServiceHandleFunc is the handler for each Microservice
type ServiceHandleFunc func(ctx IContext) error

// NewMicroservice is the constructor function of Microservice
func NewMicroservice() *Microservice {
	fmt.Println("Microservice started")
	return &Microservice{}
}

func (ms *Microservice) ConsumerConfig() *sarama.Config {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Version = sarama.V2_5_0_0 // Set to Kafka version used
	return config
}

func (ms *Microservice) StartConsume(servers, groupID string) error {
	config := ms.ConsumerConfig()

	client, err := sarama.NewConsumerGroup([]string{servers}, groupID, config)
	if err != nil {
		ms.Log("Consumer", fmt.Sprintf("Error creating consumer group: %v", err))
		return err
	}
	// defer client.Close()
	ms.kafkaClient = KafkaConsumer{
		consumerClient: client,
		servers:        servers,
		groupID:        groupID,
	}
	return nil
}

func Producer(topic string, payload any) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Version = sarama.V2_5_0_0

	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, config)
	if err != nil {
		fmt.Println("Error creating producer: ", err)
		return
	}
	defer producer.Close()

	b, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshalling payload: ", err)
		return
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(b),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		fmt.Println("Error sending message: ", err)
		return
	}

	fmt.Printf("Message sent to partition %d at offset %d\n", partition, offset)
}

// Consume registers a consumer for the service
func (ms *Microservice) Consume(topic string, h ServiceHandleFunc) error {
	go func() {

		handler := &ConsumerGroupHandler{
			ms:    ms,
			h:     h,
			topic: topic,
		}

		if ms.kafkaClient.consumerClient == nil {
			if ms.kafkaClient.servers != "" && ms.kafkaClient.groupID != "" {
				err := ms.StartConsume(ms.kafkaClient.servers, ms.kafkaClient.groupID)
				if err != nil {
					ms.Log("Consumer", fmt.Sprintf("Error starting consumer: %v", err))
					return
				}
			} else {
				ms.Log("Consumer", "No consumer client found")
				return
			}
		}
		ctx := context.Background()
		for {
			if err := ms.kafkaClient.consumerClient.Consume(ctx, []string{topic}, handler); err != nil {
				ms.Log("Consumer", fmt.Sprintf("Error during consume: %v", err))
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()
	return nil
}

// Start starts the microservice
func (ms *Microservice) Start() error {
	osQuit := make(chan os.Signal, 1)
	ms.exitChannel = make(chan bool, 1)
	signal.Notify(osQuit, syscall.SIGTERM, syscall.SIGINT)
	exit := false
	for {
		if exit {
			break
		}
		select {
		case <-osQuit:
			exit = true
		case <-ms.exitChannel:
			exit = true
		}
	}
	return nil
}

// Stop stops the microservice
func (ms *Microservice) Stop() {
	if ms.exitChannel == nil {
		return
	}
	ms.exitChannel <- true
}

// Cleanup performs cleanup before exit
func (ms *Microservice) Cleanup() error {
	if ms.kafkaClient.consumerClient != nil {
		ms.kafkaClient.consumerClient.Close()
	}

	return nil
}

// Log logs a message to the console
func (ms *Microservice) Log(tag string, message string) {
	fmt.Println(tag+":", message)
}

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler
type ConsumerGroupHandler struct {
	ms    *Microservice
	h     ServiceHandleFunc
	topic string
}

// Setup is run at the beginning of a new session
func (handler *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session
func (handler *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from a claim
func (handler *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		handler.ms.Log("Consumer", fmt.Sprintf("Message received: %s", string(msg.Value)))
		ctx := NewConsumerContext(handler.ms, string(msg.Value))
		if err := handler.h(ctx); err != nil {
			handler.ms.Log("Consumer", fmt.Sprintf("Handler error: %v", err))
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

// ConsumerContext implements IContext
type ConsumerContext struct {
	ms      *Microservice
	message string
}

// NewConsumerContext is the constructor function for ConsumerContext
func NewConsumerContext(ms *Microservice, message string) *ConsumerContext {
	return &ConsumerContext{
		ms:      ms,
		message: message,
	}
}

// Log logs a message
func (ctx *ConsumerContext) Log(message string) {
	fmt.Println("Consumer:", message)
}

// Param returns a parameter by name (not used in this example)
func (ctx *ConsumerContext) Param(name string) string {
	return ""
}

// ReadInput returns the message
func (ctx *ConsumerContext) ReadInput(data any) error {

	val := reflect.ValueOf(&data)
	switch val.Kind() {
	case reflect.Struct:
		if err := json.Unmarshal([]byte(ctx.message), &data); err != nil {
			return fmt.Errorf("%s, payload: %s", err.Error(), ctx.message)
		}
		return nil
	case reflect.Ptr, reflect.Interface:
		if err := json.Unmarshal([]byte(ctx.message), data); err != nil {
			return fmt.Errorf("%s, payload: %s", err.Error(), ctx.message)
		}
		return nil
	case reflect.Map, reflect.Slice:
		if err := json.Unmarshal([]byte(ctx.message), &data); err != nil {
			return fmt.Errorf("%s, payload: %s", err.Error(), ctx.message)
		}
		return nil
	default:
		if err := json.Unmarshal([]byte(ctx.message), &data); err != nil {
			// return fmt.Errorf("unsupported type %v", val.Kind())
			return fmt.Errorf("%s, payload: %s", err.Error(), ctx.message)
		}
		return nil
	}

}

// Response returns a response to the client (not used in this example)
func (ctx *ConsumerContext) Response(responseCode int, responseData interface{}) {

}
