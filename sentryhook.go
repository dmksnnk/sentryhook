// Package sentryhook provides a hook for logrus (https://github.com/sirupsen/logrus) logger
// for sending errors and messages to the Sentry on specific log level.
package sentryhook

import (
	"fmt"

	"github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
)

// SentrySender is interface for sender to Sentry (which raven.Client implements)
type SentrySender interface {
	CaptureMessage(message string, tags map[string]string, interfaces ...raven.Interface) string
	CaptureError(err error, tags map[string]string, interfaces ...raven.Interface) string
	CaptureMessageAndWait(message string, tags map[string]string, interfaces ...raven.Interface) string
	CaptureErrorAndWait(err error, tags map[string]string, interfaces ...raven.Interface) string
}

// SentryHook is a struct which holds levels on which send events, and sender which sends events
type SentryHook struct {
	asyncLevels map[logrus.Level]struct{}
	syncLevels  map[logrus.Level]struct{}
	sender      SentrySender
}

// New creates to Sentry hook. If client is nil, hook will use raven.DefaultClient
func New(client SentrySender) *SentryHook {
	if client == nil {
		client = raven.DefaultClient
	}
	return &SentryHook{
		asyncLevels: make(map[logrus.Level]struct{}),
		syncLevels:  make(map[logrus.Level]struct{}),
		sender:      client,
	}
}

// SetAsync sets hooks for logrus log levels.
// Will send message to the Sentry asynchronously (non-blocking)
func (hook *SentryHook) SetAsync(levels ...logrus.Level) error {
	for _, l := range levels {
		if _, ok := hook.syncLevels[l]; ok {
			return fmt.Errorf("Log level %v already in sync levels", l)
		}
		hook.asyncLevels[l] = struct{}{}
	}
	return nil
}

// SetSync sets hooks for logrus log levels.
// Will send message to the Sentry synchronously (blocking)
func (hook *SentryHook) SetSync(levels ...logrus.Level) error {
	for _, l := range levels {
		if _, ok := hook.asyncLevels[l]; ok {
			return fmt.Errorf("Log level %v already in async levels", l)
		}
		hook.syncLevels[l] = struct{}{}
	}
	return nil
}

// Fire implements Hook interface. Sends error/message to sentry
func (hook *SentryHook) Fire(entry *logrus.Entry) error {
	if _, ok := hook.asyncLevels[entry.Level]; ok {
		hook.sendAsync(entry)
		return nil
	}
	if _, ok := hook.syncLevels[entry.Level]; ok {
		hook.sendSync(entry)
		return nil
	}
	return nil
}

// Levels implements Hook interface. Returns registered levels
func (hook *SentryHook) Levels() []logrus.Level {
	levels := []logrus.Level{}

	for l := range hook.asyncLevels {
		levels = append(levels, l)
	}

	for l := range hook.syncLevels {
		levels = append(levels, l)
	}
	return levels
}

// Send error/message to Sentry asynchronously
func (hook *SentryHook) sendAsync(entry *logrus.Entry) {
	tags := hook.makeTags(entry.Data)
	loggedErr, ok := entry.Data[logrus.ErrorKey].(error)
	if !ok {
		hook.sender.CaptureMessage(entry.Message, tags)
		return
	}
	hook.sender.CaptureError(loggedErr, tags, &raven.Message{Message: entry.Message})
}

// Send error/message to Sentry synchronously
func (hook *SentryHook) sendSync(entry *logrus.Entry) {
	tags := hook.makeTags(entry.Data)
	loggedErr, ok := entry.Data[logrus.ErrorKey].(error)
	if !ok {
		hook.sender.CaptureMessageAndWait(entry.Message, tags)
		return
	}
	hook.sender.CaptureErrorAndWait(loggedErr, tags, &raven.Message{Message: entry.Message})
}

// convert log fields to tags
func (SentryHook) makeTags(fields logrus.Fields) map[string]string {
	tags := map[string]string{}

	for k, v := range fields {
		if k == logrus.ErrorKey {
			continue // omit setting error as tag
		}
		tags[k] = fmt.Sprint(v)
	}
	return tags
}
