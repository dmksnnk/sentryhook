package sentryhook

import (
	"errors"
	"reflect"
	"testing"

	"github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
)

// fakeClient just records how it was called
type fakeClient struct {
	called string
}

func (fc *fakeClient) CaptureMessage(message string, tags map[string]string, interfaces ...raven.Interface) string {
	fc.called = "CaptureMessage"
	return "fake message ID"
}

func (fc *fakeClient) CaptureError(err error, tags map[string]string, interfaces ...raven.Interface) string {
	fc.called = "CaptureError"
	return "fake message ID"
}

func (fc *fakeClient) CaptureMessageAndWait(message string, tags map[string]string, interfaces ...raven.Interface) string {
	fc.called = "CaptureMessageAndWait"
	return "fake message ID"
}

func (fc *fakeClient) CaptureErrorAndWait(err error, tags map[string]string, interfaces ...raven.Interface) string {
	fc.called = "CaptureErrorAndWait"
	return "fake message ID"
}

func TestNew(t *testing.T) {
	fc := fakeClient{}
	type args struct {
		client SentrySender
	}
	tests := []struct {
		name string
		args args
		want *SentryHook
	}{
		{
			name: "nil client",
			args: args{client: nil},
			want: &SentryHook{
				asyncLevels: make(map[logrus.Level]struct{}),
				syncLevels:  make(map[logrus.Level]struct{}),
				sender:      raven.DefaultClient,
			},
		},
		{
			name: "custom client",
			args: args{client: &fc},
			want: &SentryHook{
				asyncLevels: make(map[logrus.Level]struct{}),
				syncLevels:  make(map[logrus.Level]struct{}),
				sender:      &fc,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.client); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSettingLevels(t *testing.T) {
	s := New(nil)

	// no error
	err := s.SetAsync(logrus.ErrorLevel)
	if err != nil {
		t.Errorf("Want error to be nil, but get %v", err)
	}
	err = s.SetSync(logrus.PanicLevel)
	if err != nil {
		t.Errorf("Want error to be nil, but get %v", err)
	}

	wantLevels := map[logrus.Level]struct{}{logrus.ErrorLevel: {}}
	if !reflect.DeepEqual(s.asyncLevels, wantLevels) {
		t.Errorf("Want asyncLevels to be %v, but get %v", wantLevels, s.asyncLevels)
	}
	wantLevels = map[logrus.Level]struct{}{logrus.PanicLevel: {}}
	if !reflect.DeepEqual(s.syncLevels, wantLevels) {
		t.Errorf("Want syncLevels to be %v, but get %v", wantLevels, s.syncLevels)
	}

	// error, trying to set async level which is already in sync levels
	err = s.SetAsync(logrus.PanicLevel)
	wantErr := "Log level panic already in sync levels"
	if err.Error() != wantErr {
		t.Errorf("Want error to be \"%s\", but get %s", wantErr, err.Error())
	}
	// error, trying to set sync level which is already in async levels
	err = s.SetSync(logrus.ErrorLevel)
	wantErr = "Log level error already in async levels"
	if err.Error() != wantErr {
		t.Errorf("Want error to be \"%s\", but get %s", wantErr, err.Error())
	}
}

func TestGetLevels(t *testing.T) {
	sh := New(nil)

	wantLevels := []logrus.Level{}
	if !reflect.DeepEqual(sh.Levels(), wantLevels) {
		t.Errorf("Want levels to be %v, but get %v", wantLevels, sh.Levels())
	}

	_ = sh.SetAsync(logrus.ErrorLevel)
	_ = sh.SetSync(logrus.PanicLevel)

	wantLevels = []logrus.Level{logrus.ErrorLevel, logrus.PanicLevel}
	if !reflect.DeepEqual(sh.Levels(), wantLevels) {
		t.Errorf("Want levels to be %v, but get %v", wantLevels, sh.Levels())
	}
}

func TestMakeTags(t *testing.T) {
	sh := New(nil)

	tags := sh.makeTags(logrus.Fields{
		"AAA":           123,
		logrus.ErrorKey: errors.New("some error"),
		"BBB":           map[string]int{"aaa": 123},
	})

	wantTags := map[string]string{
		"AAA": "123",
		"BBB": "map[aaa:123]",
	}

	if !reflect.DeepEqual(tags, wantTags) {
		t.Errorf("Want tags to be %v, but get %v", wantTags, tags)
	}
}

func TestLogHook(t *testing.T) {
	err := errors.New("some error")

	tests := []struct {
		name    string
		logFunc func(args ...interface{})
		called  string
	}{
		{
			name:    "async error",
			logFunc: logrus.WithError(err).WithField("some key", "some value").Error,
			called:  "CaptureError",
		},
		{
			name:    "sync error",
			logFunc: logrus.WithError(err).WithField("some key", "some value").Info,
			called:  "CaptureErrorAndWait",
		},
		{
			name:    "async message",
			logFunc: logrus.WithField("some key", "some value").Error,
			called:  "CaptureMessage",
		},
		{
			name:    "sync message",
			logFunc: logrus.WithField("some key", "some value").Info,
			called:  "CaptureMessageAndWait",
		},
		{
			name:    "no calling",
			logFunc: logrus.WithField("some key", "some value").Warn,
			called:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := fakeClient{}
			sh := New(&fc)
			_ = sh.SetAsync(logrus.ErrorLevel)
			_ = sh.SetSync(logrus.InfoLevel)
			logrus.AddHook(sh)

			tt.logFunc() // calling log

			// checking fake client
			if !reflect.DeepEqual(fc.called, tt.called) {
				t.Errorf("want to be called = %v, but get %v", tt.called, fc.called)
			}
		})
	}

}
