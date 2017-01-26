package bus

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/InnovaCo/broforce/config"
	"github.com/InnovaCo/broforce/logger"
)

var once sync.Once
var instance *EventsBus
var busConfig = make([]adapterConfig, 0)

type adapterConfig struct {
	EventTypes string
	Condition  *regexp.Regexp
	Adapters   []adapter
}

func registry(eventsTypes []string, adap adapter) {
Events:
	for _, et := range eventsTypes {
		for _, cfg := range busConfig {
			if strings.Compare(cfg.EventTypes, et) == 0 {
				cfg.Adapters = append(cfg.Adapters, adap)
				adap.Run()
				continue Events
			}
		}
		if cond, err := regexp.Compile(et); err == nil {
			busConfig = append(busConfig, adapterConfig{
				EventTypes: et,
				Condition:  cond,
				Adapters:   []adapter{adap}})
			adap.Run()
		}
	}
}

type Event struct {
	Subject string
	Coding  string
	Data    []byte
}

type SafeParams struct {
	Retry int
	Delay time.Duration
}

type Task interface {
	Run(eventBus *EventsBus, cfg config.ConfigData) error
}

type Handler func(e Event) error

func SafeRun(r func(eventBus *EventsBus, cfg config.ConfigData) error, sp SafeParams) func(eventBus *EventsBus, cfg config.ConfigData) error {
	return func(eventBus *EventsBus, cfg config.ConfigData) error {
		for {
			if err := r(eventBus, cfg); err != nil {
				logger.Log.Error(err)
				if sp.Retry <= 0 {
					return err
				} else {
					sp.Retry--
					time.Sleep(sp.Delay)
				}
			} else {
				break
			}
		}
		return nil
	}
}

type adapter interface {
	Run() error
	Publish(e Event) error
	Subscribe(subject string, h Handler)
}

type EventsBus struct {
}

func New() *EventsBus {
	once.Do(func() {
		instance = &EventsBus{}
	})
	return instance
}

func (p *EventsBus) Publish(e Event) error {
	for _, cfg := range busConfig {
		if len(cfg.Condition.FindAllString(e.Subject, -1)) == 1 {
			for _, a := range cfg.Adapters {
				if err := a.Publish(e); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (p *EventsBus) Subscribe(subject string, h Handler) {
	for _, cfg := range busConfig {
		if len(cfg.Condition.FindAllString(subject, -1)) == 1 {
			for _, a := range cfg.Adapters {
				a.Subscribe(subject, h)
			}
		}
	}
}
