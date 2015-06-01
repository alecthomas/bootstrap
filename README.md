# General application utilities

The functionality is split up into "modules", activated with a bitmask.

```go
app := kingpin.New("app", "Description.")
util.Bootstrap(app, util.LoggingModule|util.DebugModule|util.PIDFileModule|util.DaemonizeModule, &util.Options{
		LogToStderrByDefault: true,
		UseSystemPIDFilePath: true,
})
```

## Modules

| Module flag | Provided functionality |
|-------------|------------------------|
| `util.LoggingModule` | Configurable logging via flags, including sinks and levels. |
| `util.DebugModule` | Add a `util.DebugFlag`, also used by some of the other modules. |
| `util.PIDFileModule` | Prevent multiple instances of the application running, via a `--pid-file` flag. |
| `util.DaemonizeModule` | Add `--daemon` flag which daemonizes the process. Usually used in conjunection with `util.PIDFileModule`. |

As a convenience, you can use `util.AllModules` to activate all modules.
