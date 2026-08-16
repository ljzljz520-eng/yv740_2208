# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
?   	example.com/golabel/transcodewrap/cmd/transcode	[no test files]
--- FAIL: TestCommandStopsWhenResolutionCannotBeApplied (0.00s)
    run_test.go:68: command error = <nil>, output = "progress: 0%\nprogress: 50%\nprogress: 100%\ntask succeeded: /tmp/TestCommandStopsWhenResolutionCannotBeApplied1170632570/001/output.mp4\n"
    run_test.go:71: output file state error = <nil>
    run_test.go:74: command output = "progress: 0%\nprogress: 50%\nprogress: 100%\ntask succeeded: /tmp/TestCommandStopsWhenResolutionCannotBeApplied1170632570/001/output.mp4\n"
FAIL
FAIL	example.com/golabel/transcodewrap/command	0.002s
ok  	example.com/golabel/transcodewrap/transcode	0.002s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/transcode): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/transcode): exit `0`
