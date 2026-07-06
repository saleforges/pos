package port

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, eventName string, payload interface{}) error
}
