package kafka

import (
	"context"
	"sync"
	"time"

	// "go_boilerplate/pkg/logger"

	"github.com/Invan2/invan_catalog_service/config"
	"github.com/Invan2/invan_catalog_service/pkg/logger"
	"github.com/Shopify/sarama"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
)

type Kafka struct {
	ctx           context.Context
	log           logger.Logger
	cfg           config.Config
	consumers     map[string]*Consumer
	publishers    map[string]*Publisher
	saramaConfig  *sarama.Config
	consumerGroup sarama.ConsumerGroup
	ready         chan struct{}
	wg            *sync.WaitGroup
}

type KafkaI interface {
	RunConsumers()
	AddConsumer(topic string, handler HandlerFunc)
	Push(topic string, e cloudevents.Event) error
	AddPublisher(topic string)
	Shutdown() error
}

func NewKafka(ctx context.Context, cfg config.Config, log logger.Logger) (KafkaI, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V3_2_0_0
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Group.Heartbeat.Interval = time.Second * 30
	saramaConfig.Consumer.Group.Session.Timeout = time.Second * 90
	saramaConfig.Consumer.Group.Rebalance.Timeout = time.Second * 90 * 3
	saramaConfig.Producer.MaxMessageBytes = 1024 * 1024 * 40
	saramaConfig.Consumer.MaxProcessingTime = time.Second * 60
	saramaConfig.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup([]string{cfg.KafkaUrl}, config.ConsumerGroupID, saramaConfig)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				consumerGroup.Close()
				return
			default:
				err = <-consumerGroup.Errors()
				if err != nil {
					log.Error("error on kafka", logger.Error(err))
				}
			}
		}
	}()

	kafka := &Kafka{
		ctx:           ctx,
		log:           log,
		cfg:           cfg,
		consumers:     make(map[string]*Consumer),
		publishers:    make(map[string]*Publisher),
		saramaConfig:  saramaConfig,
		ready:         make(chan struct{}),
		wg:            &sync.WaitGroup{},
		consumerGroup: consumerGroup,
	}

	return kafka, nil
}

// RunConsumers ...
func (kafka *Kafka) RunConsumers() {
	topics := []string{}

	for _, consumer := range kafka.consumers {
		topics = append(topics, consumer.topic)
		// fmt.Println("Key:", consumer.topic, "=>", "consumer:", consumer)
	}
	kafka.log.Info("topics:", logger.Any("topics:", topics))

	kafka.wg.Add(1)
	go func() {
		defer kafka.wg.Done()
		for {
			if err := kafka.consumerGroup.Consume(kafka.ctx, topics, kafka); err != nil {
				kafka.log.Error("error while consuming", logger.Error(err))
			}
			if kafka.ctx.Err() != nil {
				return
			}
			kafka.ready = make(chan struct{})
		}
	}()

	<-kafka.ready
	kafka.log.Warn("consumer group started")
}

func CreateEvent(t, s string, v interface{}) (cloudevents.Event, error) {
	event := cloudevents.NewEvent()
	id, err := uuid.NewRandom()
	if err != nil {
		return event, err
	}
	event.SetType(t)
	event.SetSource(s)
	event.SetID(id.String())
	err = event.SetData(cloudevents.ApplicationJSON, v)
	return event, err
}

func (kafka *Kafka) Shutdown() error {
	kafka.log.Warn("shutting down pub-sub server")
	select {
	case <-kafka.ctx.Done():
		kafka.log.Warn("terminating: context cancelled")
	default:
	}
	kafka.wg.Wait()
	kafka.consumerGroup.Close()

	for _, publisher := range kafka.publishers {
		if err := publisher.sender.Close(context.Background()); err != nil {
			kafka.log.Error("could not close sender", logger.Any("topic", publisher.topic), logger.Error(err))
		}
	}

	return nil
}
