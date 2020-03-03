# Updater

The goals of this library are to provide an updater that:

- Is simple
- Works on all our platforms (at least OS X, Windows, Linux)
- Recovers from non-fatal errors
- Every request or command execution should timeout (nothing blocks)
- Can recover from failures in its environment
- Can run as an unprivileged background service
- Has minimal dependencies
- Is well tested
- Is secure
- Reports failures and activity
- Can notify the user of any non-transient failures

## Packages

The main package is the updater core, there are other support packages:

- command: Executes a command with a timeout
- keybase: Keybase specific behavior for updates
- osx: MacOS specific UI
- process: Utilities to find and terminate Processes
- saltpack: Verify updates with [saltpack](https://saltpack.org/)
- service: Runs the updater as a background service
- sources: Update sources for remote locations (like S3), or locally (for testing)
- test: Test resources
- util: Utilities for updating, such as digests, env, file, http, unzip, etc.
- watchdog: Utility to monitor processes and restart them (like launchd), for use with updater service
- windows: Windows specific UI
