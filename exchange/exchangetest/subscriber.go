package exchangetest

import "github.com/ParadigmFoundation/go-obm"

type FakeSubscriber struct {
	ch chan *obm.Update
}

func NewFakeSubscriber(ch chan *obm.Update) *FakeSubscriber {
	return &FakeSubscriber{ch: ch}
}

func (s *FakeSubscriber) OnSnapshot(_ string, u *obm.Update) error {
	s.ch <- u
	return nil
}

func (s *FakeSubscriber) OnUpdate(_ string, u *obm.Update) error {
	s.ch <- u
	return nil
}
