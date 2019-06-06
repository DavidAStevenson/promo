package promo

import (
	"runtime"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
)

const ip string = "192.168.0.1"

type statusStringTest struct {
	status Status
	str    string
}

var statusStringTests = []statusStringTest{
	{Inactive, "Inactive"},
	{Active, "Active"},
	{Expired, "Expired"},
}

func TestStatusString(t *testing.T) {
	for i, test := range statusStringTests {
		if res := test.status.String(); res != test.str {
			t.Errorf("#%d: bad string return, got: %v want: %#v", i, res, test.str)
		}
	}
}

func TestInitialState(t *testing.T) {
	reset()
	var b Banner
	if b.status != Inactive {
		t.Errorf("Bad initial Banner status, got: %v want Inactive", b.status.String())
	}
}

// Test that status transitions from Inactive to Active, once the start time is reached
func TestActivation(t *testing.T) {
	mockclock := clock.NewMock()
	setClock(mockclock) // replace clock with mock for speedy testing

	now := mockclock.Now()
	reset()
	b, _ := New("Banner1", now.Add(1*time.Hour), now.Add(24*time.Hour))

	runtime.Gosched()

	if res := b.AllowDisplay(ip); res != false {
		t.Errorf("Bad Banner status, got: %v want %v", res, false)
	}

	// wind clock forward until after start time; enter display period
	mockclock.Add(2 * time.Hour)

	if res := b.AllowDisplay(ip); res != true {
		t.Errorf("Bad Banner status, got: %v want %v", res, true)
	}
}

// Test that status transitions from Active to Expired, once the end time is reached
func TestExpiration(t *testing.T) {
	mockclock := clock.NewMock()
	setClock(mockclock) // replace clock with mock for speedy testing

	now := mockclock.Now()
	reset()
	b, _ := New("Banner1", now, now.Add(1*time.Hour))

	runtime.Gosched()

	if res := b.AllowDisplay(ip); res != true {
		t.Errorf("Bad Banner status, got: %v want %v", res, true)
	}

	// wind clock forward until after display period
	mockclock.Add(1*time.Hour + 1*time.Second)

	if res := b.AllowDisplay(ip); res != false {
		t.Errorf("Bad Banner status, got: %v want %v", res, false)
	}
}

// Test that an initially expired promo may not be displayed
func TestPreExpired(t *testing.T) {
	var tm time.Time // initial value is in the past
	reset()
	b, _ := New("Banner1", tm, tm)
	if res := b.AllowDisplay(ip); res != false {
		t.Errorf("Bad Banner status, got: %v want %v", res, false)
	}
}

func TestDifferentTimezones(t *testing.T) {
	mockclock := clock.NewMock()
	setClock(mockclock) // replace clock with mock for speedy testing

	d0 := time.Date(2019, 4, 4, 11, 59, 59, 0, time.UTC)
	mockclock.Set(d0)

	secondsEastOfUTC := int((9 * time.Hour).Seconds())
	tokyo := time.FixedZone("Tokyo Time", secondsEastOfUTC)

	timeInTokyo := time.Date(2019, 4, 4, 21, 0, 0, 0, tokyo)
	timeInUTC := time.Date(2019, 4, 4, 13, 0, 0, 0, time.UTC)

	reset()
	// with 9 hours difference between Tokyo and UTC, promotion period is 1 hour
	b, _ := New("Banner1", timeInTokyo, timeInUTC)

	runtime.Gosched()

	if res := b.AllowDisplay(ip); res != false {
		t.Errorf("Bad pre-period Banner status, got: %v want %v", res, false)
	}

	// wind clock forward two seconds into the display period
	mockclock.Add(2 * time.Second)

	if res := b.AllowDisplay(ip); res != true {
		t.Errorf("Bad display period Banner status, got: %v want %v", res, true)
	}

	// wind clock forward an hour until after display period
	mockclock.Add(1 * time.Hour)

	if res := b.AllowDisplay(ip); res != false {
		t.Errorf("Bad post-period Banner status, got: %v want %v", res, false)
	}
}

func TestAllowDisplayForQABeforePeriod(t *testing.T) {
	cases := []struct {
		ip   string
		want bool
	}{
		{"10.0.0.1", true},
		{"10.0.0.2", true},
		{"192.168.0.2", false},
	}

	timeclock = clock.New() // use real clock for this test

	now := time.Now()
	start := now.Add(1 * 24 * time.Hour)
	end := now.Add(8 * 24 * time.Hour)

	reset()
	b, _ := New("Banner1", start, end)

	for _, c := range cases {
		if got := b.AllowDisplay(c.ip); got != c.want {
			t.Errorf("AllowDisplay(%v)) => %t, want %t", c.ip, got, c.want)
		}
	}
}

func TestChoose(t *testing.T) {
	mockclock := clock.NewMock()
	setClock(mockclock) // replace clock with mock for speedy testing

	tests := []struct {
		name string
		st   time.Time
		en   time.Time
	}{
		{"Banner1", time.Date(2019, 4, 5, 13, 0, 0, 0, time.UTC), time.Date(2019, 4, 5, 17, 0, 0, 0, time.UTC)},
		{"Banner2", time.Date(2019, 4, 5, 14, 0, 0, 0, time.UTC), time.Date(2019, 4, 5, 16, 0, 0, 0, time.UTC)},
		{"Banner3", time.Date(2019, 4, 5, 13, 30, 0, 0, time.UTC), time.Date(2019, 4, 5, 17, 30, 0, 0, time.UTC)},
	}
	// Banner1 starts first, ends second
	// Banner2 starts last, ends first
	// Banner3 starts second, ends last

	reset()
	var b1, b2, b3 *Banner
	var err error

	i := 0
	if b1, err = New(tests[i].name, tests[i].st, tests[i].en); err != nil {
		t.Errorf("Couldn't create promo with name '%v', error: %v", tests[i].name, err)
	}
	i++
	if b2, err = New(tests[i].name, tests[i].st, tests[i].en); err != nil {
		t.Errorf("Couldn't create promo with name '%v', error: %v", tests[i].name, err)
	}
	i++
	if b3, err = New(tests[i].name, tests[i].st, tests[i].en); err != nil {
		t.Errorf("Couldn't create promo with name '%v', error: %v", tests[i].name, err)
	}

	runtime.Gosched()

	d0 := time.Date(2019, 4, 5, 12, 45, 0, 0, time.UTC)
	mockclock.Set(d0)

	// test none are to be displayed
	if got := Choose(ip); got != nil {
		t.Errorf("Choose() => %v, want nil", got.name)
	}

	// wind clock forward 30 minutes into Banner1 display period
	mockclock.Add(30 * time.Minute) // 13:15
	if got := Choose(ip); got != b1 {
		t.Errorf("Choose() => %v, want %v", got, b1)
	}

	// wind clock into Banner2 display period; its display period ends first, so has priority
	mockclock.Add(1 * time.Hour) // 14:15
	if got := Choose(ip); got != b2 {
		t.Errorf("Choose() => %v, want %v", got, b2)
	}

	// wind clock past Banner2 display period; Banner1 again has priority
	mockclock.Add(2 * time.Hour) // 16:15
	if got := Choose(ip); got != b1 {
		t.Errorf("Choose() => %v, want %v", got, b1)
	}

	// wind clock past Banner1 display period; Banner3 now has priority
	mockclock.Add(1 * time.Hour) // 17:15
	if got := Choose(ip); got != b3 {
		t.Errorf("Choose() => %v, want %v", got, b3)
	}

	// wind clock forward 30 minutes, and all promos are expired
	mockclock.Add(30 * time.Minute) // 17:45
	if got := Choose(ip); got != nil {
		t.Errorf("Choose() => %v, want nil", got.name)
	}
}
