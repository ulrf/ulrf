package models

import (
	"fmt"
	"github.com/fatih/color"
	"time"
)

var (
	maxTime = time.Second
)

func printSince(t time.Time, frmt string, args ...interface{}) {
	if since := time.Since(t); since > maxTime {
		color.Yellow("[%s] %s", since, fmt.Sprintf(frmt, args...))
	}
}

func printNeedSince(t time.Time, need time.Duration, frmt string, args ...interface{}) {
	if since := time.Since(t); since > need {
		color.Yellow("[%s] %s", since, fmt.Sprintf(frmt, args...))
	}
}
