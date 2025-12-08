package email

import (
	"errors"
	"sync"
)

type TestSender struct {
	emails sync.Map
}

func (t *TestSender) Send(to string, c Composer) (err error) {
	t.emails.Store(to, c)
	return nil
}

func LoadTestEmail(to string) (c Composer, err error) {
	if driver == nil {
		err = errors.New("email driver is nil")
		return
	}

	sender, ok := driver.(*TestSender)
	if !ok {
		err = errors.New("email driver is not TestSender")
		return
	}

	v, ok := sender.emails.Load(to)
	if !ok {
		err = errors.New("email for " + to + " not exist")
		return
	}

	return v.(Composer), nil
}
