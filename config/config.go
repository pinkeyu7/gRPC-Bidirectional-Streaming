package config

import "os"

func GetListenNetwork() string {
	return os.Getenv("LISTEN_NETWORK")
}

func GetListenAddress() string {
	return os.Getenv("LISTEN_ADDRESS")
}
