package metrics

import (
	"expvar"
	"io/ioutil"
	"syscall"
)

func getFDLimit() (uint64, error) {
	var rlimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlimit); err != nil {
		return 0, err
	}
	return rlimit.Cur, nil
}

func getFDUsage() (uint64, error) {
	fds, err := ioutil.ReadDir("/proc/self/fd")
	if err != nil {
		return 0, err
	}
	return uint64(len(fds)), nil
}

func init() {
	expvar.Publish("fds", expvar.Func(func() interface{} {
		open, _ := getFDUsage()
		max, _ := getFDLimit()
		return map[string]uint64{
			"Max":  max,
			"Open": open,
		}
	}))
}
