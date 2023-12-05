package execution

import "sync"

type Subscription[T interface{}] struct {
	//nolint:structcheck // linter bug with generic struct
	channel chan T
	//nolint:structcheck // linter bug with generic struct
	dispatcher *Dispatcher[T]
}

type Dispatcher[T interface{}] struct {
	//nolint:structcheck // linter bug with generic struct
	mutex sync.Mutex
	//nolint:structcheck // linter bug with generic struct
	subscriptions []*Subscription[T]
}

func (d *Dispatcher[T]) Subscribe(capacity int) *Subscription[T] {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	subscription := &Subscription[T]{
		channel:    make(chan T, capacity),
		dispatcher: d,
	}
	d.subscriptions = append(d.subscriptions, subscription)

	return subscription
}

func (s *Subscription[T]) Unsubscribe() {
	if s.dispatcher == nil {
		return
	}

	s.dispatcher.Unsubscribe(s)
}

func (s *Subscription[T]) Channel() <-chan T {
	return s.channel
}

func (d *Dispatcher[T]) Unsubscribe(subscription *Subscription[T]) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if subscription.dispatcher != nil {
		return
	}

	count := len(d.subscriptions)

	for i, s := range d.subscriptions {
		if s == subscription {
			if i < count-1 {
				d.subscriptions[i] = d.subscriptions[count-1]
			}

			d.subscriptions = d.subscriptions[:count-1]
			subscription.dispatcher = nil

			return
		}
	}
}

func (d *Dispatcher[T]) Fire(data T) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	for _, s := range d.subscriptions {
		select {
		case s.channel <- data:
		default:
		}
	}

	return nil
}
