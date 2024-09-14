package connectionstatus

import (
	"log"
	"time"
)

const defaultSleepTime = 10

var fibonacciDelays = []int{1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233, 300}

func initSleepTime() func() time.Duration {
	i := 0
	return func() time.Duration {
		delayDuration := 0

		if i < len(fibonacciDelays) {
			delayDuration = fibonacciDelays[i]
		} else {
			delayDuration = fibonacciDelays[len(fibonacciDelays)-1]
		}

		i++

		timeDuration := time.Duration(delayDuration) * 100 * time.Millisecond
		log.Printf("Trying %d times; waiting %v ms before retrying...\n", i, timeDuration.Milliseconds())

		return timeDuration
	}
}

func defaultErrorSleepTime() time.Duration {
	return defaultSleepTime * time.Second
}
