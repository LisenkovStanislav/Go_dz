package practice

import (
	"errors"
	"fmt"
)

type Notifier interface {
	Send(message string) error
}

type BaseNotifier struct {
	Channel string
}

func (b BaseNotifier) Label(message string) string {
	return "[" + b.Channel + "] " + message
}

func (b BaseNotifier) Name() string {
	return b.Channel
}

type EmailNotifier struct {
	BaseNotifier
	Address  string
	Messages []string
}

type PushNotifier struct {
	BaseNotifier
	DeviceID string
	Messages []string
}

type namedNotifier interface {
	Name() string
}

func channelName(n Notifier) string {
	if named, ok := n.(namedNotifier); ok {
		return named.Name()
	}
	return fmt.Sprintf("%T", n)
}

var ErrNotifierPanic = errors.New("паника канала")

type SendError struct {
	Channel string
	Reason  string
}

type BrokenNotifier struct {
	BaseNotifier
}

func (n *BrokenNotifier) Send(message string) error {
	panic("драйвер канала повреждён")
}

var ErrEmptyMessage = errors.New("пустое сообщение")

func (e *SendError) Error() string {
	return fmt.Sprintf("канал %s: %s", e.Channel, e.Reason)
}

func (n *EmailNotifier) Send(message string) error {
	if message == "" {
		return ErrEmptyMessage
	}
	if n.Address == "" {
		return &SendError{
			Channel: n.Channel,
			Reason:  "не задан адрес",
		}
	}
	n.Messages = append(n.Messages, n.Label(message))
	return nil
}

func (n *PushNotifier) Send(message string) error {
	if message == "" {
		return ErrEmptyMessage
	}
	if n.DeviceID == "" {
		return &SendError{
			Channel: n.Channel,
			Reason:  "не задан device id",
		}
	}
	n.Messages = append(n.Messages, n.Label(message))
	return nil
}

func Dispatch(all []Notifier, message string) error {
	for _, n := range all {
		if err := n.Send(message); err != nil {
			return fmt.Errorf("канал %s: %w", channelName(n), err)
		}
	}
	return nil
}

func SafeSend(n Notifier, message string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v", ErrNotifierPanic, r)
		}
	}()
	return n.Send(message)
}

func SafeDispatch(all []Notifier, message string) []error {
	var errs []error
	for _, n := range all {
		if err := SafeSend(n, message); err != nil {
			errs = append(errs, fmt.Errorf("канал %s: %w", channelName(n), err))
		}
	}
	return errs
}

func EmailAddress(n Notifier) (string, bool) {
	email, ok := n.(*EmailNotifier)
	if !ok || email == nil {
		return "", false
	}
	return email.Address, true
}

func Describe(n Notifier) string {
	switch v := n.(type) {
	case *EmailNotifier:
		if v == nil {
			return "unknown"
		}
		return "email: " + v.Address
	case *PushNotifier:
		if v == nil {
			return "unknown"
		}
		return "push: " + v.DeviceID
	default:
		return "unknown"
	}
}

var _ Notifier = (*EmailNotifier)(nil)

var _ Notifier = (*PushNotifier)(nil)
