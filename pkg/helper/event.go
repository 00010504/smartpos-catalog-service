package helper

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/Shopify/sarama"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func MessageToEvent(message *sarama.ConsumerMessage) cloudevents.Event {
	event := cloudevents.NewEvent()

	for _, header := range message.Headers {
		if x := string(header.Key); x == "ce_id" {
			event.SetID(string(header.Value))
		} else if x == "ce_source" {
			event.SetSource(string(header.Value))
		} else if x == "ce_type" {
			event.SetType(string(header.Value))
		} else if x == "ce_time" {
			t, _ := time.Parse("2006-01-02T15:04:05.999999999Z", string(header.Value))
			event.SetTime(t)
		} else if x == "ce_traceid" {
			event.SetExtension("traceid", string(header.Value))
		} else {
			// fmt.Println("not equal: ", x)
		}
	}

	var m map[string]interface{}
	_ = json.Unmarshal(message.Value, &m)

	_ = event.SetData(cloudevents.ApplicationJSON, m)

	return event
}

func CreateEvent(value interface{}) cloudevents.Event {
	event := cloudevents.NewEvent()
	id, _ := uuid.NewRandom()
	event.SetID(id.String())
	_ = event.SetData(cloudevents.ApplicationJSON, value)
	event.SetType("create")
	event.SetSource("v2.inventory_service_v2")

	return event
}
