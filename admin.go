package main

import (
	"fmt"
	"time"
)

func NewAdminModule(subs *EventSubs, stats *StatTracker) *AdminModule {
	return &AdminModule{
		LogSubs: subs,
		Stats:   stats,
	}
}

type AdminModule struct {
	UnimplementedAdminServer
	LogSubs *EventSubs
	Stats   *StatTracker
}

func (adm *AdminModule) Logging(_ *Nothing, outStream Admin_LoggingServer) error {
	eventCh := adm.LogSubs.Subscribe(outStream)
	for {
		select {
		case event := <-eventCh:
			if err := outStream.Send(event); err != nil {
				return fmt.Errorf("error sending event: %w", err)
			}
		case <-outStream.Context().Done():
			adm.LogSubs.Unsubscribe(outStream)
			return nil
		}
	}
}

func (adm *AdminModule) Statistics(interval *StatInterval, outStream Admin_StatisticsServer) error {
	intervalDuration := time.Duration(interval.IntervalSeconds) * time.Second
	timer := time.NewTimer(intervalDuration)
	defer timer.Stop()

	adm.Stats.Subscribe(outStream)
	for {
		select {
		case <-timer.C:
			stat, err := adm.Stats.Pull(outStream)
			if err != nil {
				return fmt.Errorf("error pulling stat: %w", err)
			}
			if err := outStream.Send(stat); err != nil {
				return fmt.Errorf("error sending stat: %w", err)
			}
			timer.Reset(intervalDuration)
		case <-outStream.Context().Done():
			adm.Stats.Unsubscribe(outStream)
			return nil
		}
	}
}
