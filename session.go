package main

import (
	"context"
	"time"
)

type session struct {
	usdd string
	actx context.Context
	acnl context.CancelFunc
	cctx context.Context
	ccnl context.CancelFunc
	date time.Time
	last time.Time
}
