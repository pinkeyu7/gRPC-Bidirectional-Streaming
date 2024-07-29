package config

import (
	"os"
	"strconv"
)

func GetPushGatewayUrl() string {
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

func GetWorkerId() string {
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

func GetServerTimeout() int {
	stStr := os.Getenv("SERVER_TIMEOUT")
	st, err := strconv.Atoi(stStr)
	if err != nil {
		return 60
	}

	return st
}
