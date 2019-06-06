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

// Promo encapsulates the behaviour of when the promo is to be displayed
type Promo struct {
	// associated promotion?
	name   string
	start  time.Time
	end    time.Time
	status Status
}

var promos map[string]*Promo

func reset() {
	promos = nil
}

func (p *Promo) changeStatus() {
	// TODO need mutex around here for thread-safety?
	if p.status == Inactive {
		p.status = Active

		p.waitForNextTransition(p.end)
	} else if p.status == Active {
		p.status = Expired

		delete(promos, p.name)
	}
	fmt.Printf("%v Promo status changed(%v)\n", timeclock.Now(), p.status)
}

func (p *Promo) waitForNextTransition(t time.Time) {
	now := timeclock.Now()
	interval := t.Sub(now)

	// start timer ticking for the next promo state transition
	go func() {
		fmt.Printf("%v Promo waiting for %v.\n", timeclock.Now(), interval)
		timer := timeclock.Timer(interval)
		<-timer.C
		p.changeStatus()
	}()
}

// New returns a new Promo that will be active during the period between startT and endT.
func New(name string, startT time.Time, endT time.Time) (*Promo, error) {
	p := &Promo{
		name:  name,
		start: startT,
		end:   endT,
	}

	if promos == nil {
		promos = make(map[string]*Promo)
	}

	_, ok := promos[name]
	if ok {
		return nil, errors.New("can't make new promo with an already used name")
	}

	now := timeclock.Now()
	t := p.start
	if now.Equal(p.start) || (now.After(p.start) && now.Before(p.end)) {
		p.status = Active
		t = p.end
	} else if now.After(p.end) {
		p.status = Expired
		return p, nil
	}
	fmt.Printf("%v Promo created (%v).\n", timeclock.Now(), p.status)

	promos[name] = p

	p.waitForNextTransition(t)

	return p, nil
}

// Name returns the name of the promo
func (p *Promo) Name() string {
	return p.name
}

// Status returns the current status of the promo
func (p *Promo) Status() string {
	return p.status.String()
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

func (p *Promo) allowDisplay(qa bool) bool {
	if qa {
		return (p.status != Expired) // QA behaviour
	}
	return (p.status == Active) // Real behaviour
}

// AllowDisplay indicates whether or not the promo may be displayed for the client IP
func (p *Promo) AllowDisplay(ip string) bool {
	return p.allowDisplay(isTestIP(ip))
}

// Choose returns the promo to be displayed
func Choose(ip string) *Promo {
	var p *Promo

	if promos == nil {
		return p
	}
	// TODO linear search inefficient for large numbers of promos
	for _, v := range promos {
		if p == nil && v.AllowDisplay(ip) {
			p = v
		} else if v.AllowDisplay(ip) && v.end.Before(p.end) {
			p = v
		}
	}
	return p
}
