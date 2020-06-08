package stats

import "fmt"

const (
	ProcessedOk    = "ok"
	ProcessedNok   = "nok"
	TotalToProcess = "total"
)

type EventData struct {
	Topic string
	Data  interface{}
}

type Subscriber chan EventData

type EventBus struct {
	subscribers map[string][]Subscriber
}

func NewEventBus() *EventBus {
	return &EventBus{subscribers: map[string][]Subscriber{}}
}

func (bus *EventBus) Publish(event EventData) {
	if subscribers, onList := bus.subscribers[event.Topic]; onList {
		for _, subscriber := range subscribers {
			go func(sub Subscriber, event EventData) {
				sub <- event
			}(subscriber, event)
		}

		return
	}

	// TODO Use logger here
	fmt.Printf("No subscriber for %v", event.Topic)
}

func (bus *EventBus) Subscribe(topic string, subscriber Subscriber) {
	if subscribers, onList := bus.subscribers[topic]; onList {
		bus.subscribers[topic] = append(subscribers, subscriber)
	} else {
		bus.subscribers[topic] = []Subscriber{subscriber}
	}
}
