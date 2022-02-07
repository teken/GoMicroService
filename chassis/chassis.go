package chassis

type EventSourceChassis struct {
	Requests *Requests
	Events   *Events
}

func NewEventSourceChassis() *EventSourceChassis {
	return &EventSourceChassis{
		Requests: &Requests{
			NewRequestManager(nil),
		},
		Events: &Events{
			NewEventManager(nil),
		},
	}
}
