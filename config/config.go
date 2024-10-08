package config

import (
	"os"
	"strconv"
	"time"
)

func GetPushGatewayURL() string {
	return os.Getenv("PUSH_GATEWAY_URL")
}

func GetListenNetwork() string {
	return os.Getenv("LISTEN_NETWORK")
}

func GetListenAddress() string {
	return os.Getenv("LISTEN_ADDRESS")
}

func GetRequestPerSecond() int {
	rqsStr := os.Getenv("REQUEST_PER_SECOND")

	rqs, err := strconv.Atoi(rqsStr)
	if err != nil {
		return 1
	}

	return rqs
}

func GetRequestTimeDuration() int {
	rtdStr := os.Getenv("REQUEST_TIME_DURATION")

	rtd, err := strconv.Atoi(rtdStr)
	if err != nil {
		return 1
	}

	return rtd
}

func GetTaskPerWorker() int {
	tpwStr := os.Getenv("TASK_PER_WORKER")

	tpw, err := strconv.Atoi(tpwStr)
	if err != nil {
		return 1
	}

	return tpw
}

func GetWorkerCount() int {
	wcStr := os.Getenv("WORKER_COUNT")

	wc, err := strconv.Atoi(wcStr)
	if err != nil {
		return 1
	}

	return wc
}

func GetWorkerID() string {
	return os.Getenv("WORKER_ID")
}

func GetWorkerIdleTime() int {
	witStr := os.Getenv("WORKER_IDLE_TIME")
	wit, err := strconv.Atoi(witStr)
	if err != nil {
		return 0
	}

	return wit
}

func GetServerTimeout() time.Duration {
	defaultTimeout := 60

	stStr := os.Getenv("SERVER_TIMEOUT")
	st, err := strconv.Atoi(stStr)
	if err != nil {
		return time.Duration(defaultTimeout) * time.Second
	}

	return time.Duration(st) * time.Second
}

func GetClientTimeout() time.Duration {
	defaultTimeout := 60

	ctStr := os.Getenv("CLIENT_TIMEOUT")
	ct, err := strconv.Atoi(ctStr)
	if err != nil {
		return time.Duration(defaultTimeout) * time.Second
	}

	return time.Duration(ct) * time.Second
}

func GetMonitorTimeInterval() time.Duration {
	defaultTimeInterval := 5

	ctStr := os.Getenv("MONITOR_TIME_INTERVAL")
	ct, err := strconv.Atoi(ctStr)
	if err != nil {
		return time.Duration(defaultTimeInterval) * time.Second
	}

	return time.Duration(ct) * time.Second
}

func GetJaegerEndpoint() string {
	return os.Getenv("JAEGER_ENDPOINT")
}
