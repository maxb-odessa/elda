package events

import (
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"elda/action"
	"elda/def"
	"elda/log"
	"elda/source"
)

type evAction struct {
	action *action.Action
	data   string
}

type Event struct {
	pattern *regexp.Regexp
	source  *source.Source
	actions []*evAction
}

func Run(sources map[string]*source.Source, actions map[string]*action.Action, events []*Event) {

	doneChan := make(chan bool, 1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range sigChan {
			log.Info("got signal %d\n", sig)
			doneChan <- true
		}
	}()

	srcChan := make(chan *def.ChanMsg)

	for _, src := range sources {
		ch := src.GetChan()
		go func(c chan *def.ChanMsg) {
			for msg := range c {
				srcChan <- msg
			}
		}(ch)
	}

	for {

		select {
		case msg, ok := <-srcChan:
			if ok {
				process(msg, events)
			}
		case <-doneChan:
			for _, ac := range actions {
				ac.Done()
			}
			for _, sc := range sources {
				sc.Done()
			}
			// close(srcChan)
			return
		}

	}

	return
}

func process(srcMsg *def.ChanMsg, events []*Event) {

	// examine all configured events
	for _, ev := range events {

		// chose event from this source only
		if ev.source.Name() != srcMsg.Name {
			continue
		}

		// match event by pattern
		log.Debug("event: matching source data '%s' to pattern '%v'\n", srcMsg.Data, ev.pattern)
		if ok := ev.pattern.MatchString(srcMsg.Data); !ok {
			log.Debug("event: not matched\n")
			continue
		} else {
			log.Debug("event: matched!\n")
		}

		// send message to all actions
		for _, ea := range ev.actions {

			// replace regex submatches
			data := ev.pattern.ReplaceAllString(srcMsg.Data, ea.data)
			log.Debug("event: data after subs replacing: <%s>\n", data)

			// send modified event data to action
			actMsg := &def.ChanMsg{Name: srcMsg.Name, Data: data}
			select {
			case ea.action.GetChan() <- actMsg:
				log.Debug("event sending '%+v' to action '%s'\n", actMsg, ea.action.Name())
			default:
				log.Warn("action '%s' channel is full\n", ea.action.Name())
			}
		}

	}
}

func New() *Event {
	return new(Event)
}

func (self *Event) SetSource(src *source.Source, pattern string) (err error) {

	self.pattern, err = regexp.Compile(pattern)
	if err != nil {
		return
	}

	self.source = src

	return nil
}

func (self *Event) AddAction(act *action.Action, data string) (err error) {

	eac := &evAction{
		action: act,
		data:   data,
	}
	self.actions = append(self.actions, eac)

	return nil
}
