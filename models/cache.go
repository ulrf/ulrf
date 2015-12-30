package models

import (
	"github.com/patrickmn/go-cache"
	"time"
)

var (
	cch = cache.New(time.Hour, time.Minute)
)
