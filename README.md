# Logrus Sentry Hook

It is a hook for [logrus](https://github.com/sirupsen/logrus) logger
for sending errors and messages to the Sentry on specific log level.
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
log.WithError(errors.New("some error")).WithField("BBB", map[string]int{"bb": 111}).Fatal("test")
```

hook will send an error to the Sentry and add log fields as tags.

In a case of log message without an error, hook will send a message with tags.

**Pro tip**

Wrap your errors with `github.com/pkg/errors`, sentry client is able to parse stack trace,
so you will have fancy issues with all trace and context you need.
