package config

import (
	"os"
	"strconv"
)

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
