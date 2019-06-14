# Logrus Sentry Hook

[![Go Report Card](https://goreportcard.com/badge/github.com/dmksnnk/sentryhook)](https://goreportcard.com/report/github.com/dmksnnk/sentryhook)
[![pipeline status](https://gitlab.com/aspidima/sentryhook/badges/master/pipeline.svg)](https://gitlab.com/aspidima/sentryhook/commits/master)
[![coverage report](https://gitlab.com/aspidima/sentryhook/badges/master/coverage.svg)](https://gitlab.com/aspidima/sentryhook/commits/master)
[![GoDoc](https://img.shields.io/badge/GoDoc-referece-blue.svg?style=flat)](https://godoc.org/github.com/dmksnnk/sentryhook)

It is a hook for [logrus](https://github.com/sirupsen/logrus) logger
for sending errors and messages to the [Sentry](https://sentry.io/) on specific log level.
It uses default sentry client, so all you need is to add a hook.

## Usage

```go
err := raven.SetDSN(dsn)
if err != nil {
    log.Fatalf("%+v", errors.Wrap(err, "Can't set up raven"))
}

hook := sentryhook.New(nil) // will use raven.DefaultClient, or provide custom client
hook.SetAsync(logrus.ErrorLevel)                   // async (non-bloking) hook for errors
hook.SetSync(logrus.PanicLevel, logrus.FatalLevel) // sync (blocking) for fatal stuff

logrus.AddHook(hook)
```

Now, when you will make a log statement, like:

```go
log.WithError(errors.New("some error")).WithField("BBB", map[string]int{"bb": 111}).Fatal("This is a fatal message")
```

hook will send an error to the Sentry and add log fields as tags.

In a case of log message without an error, hook will send a message with or without tags:

```go
log.WithField("BBB", map[string]int{"bb": 111}).Error("This is a warning message") // with tags
log.Error("This is a warning message") // without tags, just message
```

**Pro tip**

Wrap your errors with `github.com/pkg/errors`, sentry client is able to parse stack trace,
so you will have fancy issues with all trace and context you need.
