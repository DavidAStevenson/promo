package promo

import (
	"errors"
	"fmt"
	"time"

	"github.com/benbjohnson/clock"
)

var timeclock = clock.New() // Real clock, unless modified during testing

func setClock(c clock.Clock) {
	timeclock = c
}

// Status indicates the current state of a promo
type Status int

// Status may be Inactive, Active, or Expired
const (
	Inactive Status = iota
	Active
	Expired
)

func (s Status) String() string {
	return [...]string{"Inactive", "Active", "Expired"}[s]
}

// Banner encapsulates the behaviour of when the promo is to be displayed
type Banner struct {
	// associated promotion?
	name   string
	start  time.Time
	end    time.Time
	status Status
}

var promos map[string]*Banner

func reset() {
	promos = nil
}

func (b *Banner) changeStatus() {
	// TODO need mutex around here for thread-safety?
	if b.status == Inactive {
		b.status = Active

		b.waitForNextTransition(b.end)
	} else if b.status == Active {
		b.status = Expired

		delete(promos, b.name)
	}
	fmt.Printf("%v Banner status changed(%v)\n", timeclock.Now(), b.status)
}

func (b *Banner) waitForNextTransition(t time.Time) {
	now := timeclock.Now()
	interval := t.Sub(now)

	// start timer ticking for the next promo state transition
	go func() {
		fmt.Printf("%v Banner waiting for %v.\n", timeclock.Now(), interval)
		timer := timeclock.Timer(interval)
		<-timer.C
		b.changeStatus()
	}()
}

// New returns a new Banner that will be active during the period between startT and endT.
func New(name string, startT time.Time, endT time.Time) (*Banner, error) {
	b := &Banner{
		name:  name,
		start: startT,
		end:   endT,
	}

	if promos == nil {
		promos = make(map[string]*Banner)
	}

	_, ok := promos[name]
	if ok {
		return nil, errors.New("can't make new promo with an already used name")
	}

	now := timeclock.Now()
	t := b.start
	if now.Equal(b.start) || (now.After(b.start) && now.Before(b.end)) {
		b.status = Active
		t = b.end
	} else if now.After(b.end) {
		b.status = Expired
		return b, nil
	}
	fmt.Printf("%v Banner created (%v).\n", timeclock.Now(), b.status)

	promos[name] = b

	b.waitForNextTransition(t)

	return b, nil
}

// Name returns the name of the promo
func (b *Banner) Name() string {
	return b.name
}

// Status returns the current status of the promo
func (b *Banner) Status() string {
	return b.status.String()
}

var testIPs = [...]string{
	"10.0.0.1",
	"10.0.0.2",
}

func isTestIP(ip string) bool {
	for _, i := range testIPs {
		if i == ip {
			return true
		}
	}
	return false
}

func (b *Banner) allowDisplay(qa bool) bool {
	if qa {
		return (b.status != Expired) // QA behaviour
	}
	return (b.status == Active) // Real behaviour
}

// AllowDisplay indicates whether or not the promo may be displayed for the client IP
func (b *Banner) AllowDisplay(ip string) bool {
	return b.allowDisplay(isTestIP(ip))
}

// Choose returns the promo to be displayed
func Choose(ip string) *Banner {
	var b *Banner

	if promos == nil {
		return b
	}
	// TODO linear search inefficient for large numbers of promos
	for _, v := range promos {
		if b == nil && v.AllowDisplay(ip) {
			b = v
		} else if v.AllowDisplay(ip) && v.end.Before(b.end) {
			b = v
		}
	}
	return b
}
