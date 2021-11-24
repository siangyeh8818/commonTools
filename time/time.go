package time

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

const (
	// This is template in golang to check string format
	// http://stackoverflow.com/a/14106561

	// ParseFormatYYYYMMDDHH for LongYear + ZeroMonth + ZeroDay + ZeroHour
	ParseFormatYYYYMMDDHH = "2006010215"
	// ParseFormatYYYYMMDD for LongYear + ZeroMonth + ZeroDay
	ParseFormatYYYYMMDD = "20060102"
	// ParseFormatYYYYMM for LongYear + ZeroMonth
	ParseFormatYYYYMM = "200601"
	// ParseFormatYYYY for LongYear
	ParseFormatYYYY = "2006"
	// ParseFormatYYYYMMDDHHWithHyphen for LongYear + - + ZeroMonth + - + ZeroDay + T + ZeroHour
	ParseFormatYYYYMMDDHHWithHyphen = "2006-01-02T15"
	// ParseFormatYYYYMMDDWithHyphen for LongYear + - + ZeroMonth + - + ZeroDay
	ParseFormatYYYYMMDDWithHyphen = "2006-01-02"
	// ParseFormatYYYYMMWithHyphen for LongYear + - + ZeroMonth
	ParseFormatYYYYMMWithHyphen = "2006-01"
	// ParseFormatYYYYMMDDWithSlash for LongYear + / + ZeroMonth + / + ZeroDay
	ParseFormatYYYYMMDDWithSlash = "2006/01/02"

	// template defined the time layout for announcement.yaml
	template = "2006-01-02 15:04:05 (GMT-0700)"

	litTimeFormat = "2006-01-02T15:04:05.000000"
)

var (
	// ErrBadFormat is returned when parsing fails
	ErrBadFormat = errors.New("bad format string")
	// ErrInvalidDayTimeSlot defines invalid time slot error
	ErrInvalidDayTimeSlot = errors.New("Invalid day time slot")

	full = regexp.MustCompile(`P((?P<year>\d+)Y)?((?P<month>\d+)M)?((?P<day>\d+)D)?(T((?P<hour>\d+)H)?((?P<minute>\d+)M)?((?P<second>\d+)S)?)?`)

	timeNow = time.Now

	// SysZeroTime 系統定義的零時 1970-01-01
	SysZeroTime = time.Unix(0, 0)
)

// Duration defines the amount of intervening time in a time interval
// Our duration is designed to solve calendar duration
// while go's time.Duration focus in absolute duration
//
// For example:
// One month calendar duration from 2017.DEC.16 is 2018.JAN.16
// If you use go's time.Duration, you add time.Hour*24*31 for 2017.DEC.16
// to get 2018.JAN.16
// However, you add time.Hour*24*30 for 2018.JAN.16 to get 2018.FEB.16
//
// You cannot just use go's time.Duration to calculate calendar duration
// (day, month, year) (not implement for week now)
// You can use our Duration by Duration{}.AddTo(time.Time) to get calendar duration time
type Duration struct {
	Years   int
	Months  int
	Days    int
	Hours   int
	Minutes int
	Seconds int
}

// ISO8601String returns the duration by the ISO8601 format P[n]Y[n]M[n]DT[n]H[n]M[n]S [ref: https://en.wikipedia.org/wiki/ISO_8601]
func (d *Duration) ISO8601String() string {
	return fmt.Sprintf("P%dY%dM%dDT%dH%dM%dS", d.Years, d.Months, d.Days, d.Hours, d.Minutes, d.Seconds)
}

// AddTo returns time.Time which is acquired by given time.Time plust Duration
func (d *Duration) AddTo(t time.Time) time.Time {
	return t.AddDate(d.Years, d.Months, d.Days).Add(time.Duration(d.Hours)*time.Hour + time.Duration(d.Minutes)*time.Minute + time.Duration(d.Seconds)*time.Second)
}

// ParseISO8601String parses ISO8601 string with format P[n]Y[n]M[n]DT[n]H[n]M[n]S to this package's Duration
func ParseISO8601String(dur string) (*Duration, error) {
	d := &Duration{}
	if dur == "" {
		return d, nil
	}
	var (
		match []string
		re    *regexp.Regexp
	)

	if full.MatchString(dur) {
		match = full.FindStringSubmatch(dur)
		re = full
	} else {
		return nil, ErrBadFormat
	}

	for i, name := range re.SubexpNames() {
		part := match[i]
		if i == 0 || name == "" || part == "" {
			continue
		}

		val, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		switch name {
		case "year":
			d.Years = val
		case "month":
			d.Months = val
		case "day":
			d.Days = val
		case "hour":
			d.Hours = val
		case "minute":
			d.Minutes = val
		case "second":
			d.Seconds = val
		default:
			return nil, fmt.Errorf("unknown field %s", name)
		}
	}
	return d, nil
}

// MilliSecond returns time.Time in millisecond
func MilliSecond(t time.Time) int64 {
	return t.UnixNano() / time.Millisecond.Nanoseconds()
}

// UnixMillis returns millisecond in time.Time
func UnixMillis(t int64) time.Time {
	return time.Unix(0, t*time.Millisecond.Nanoseconds())
}

//NowMS return now time in millisecond
func NowMS() int64 {
	return MilliSecond(time.Now())
}

//UTCNowMs return utc+0 time in millisecond
func UTCNowMs() int64 {
	return MilliSecond(time.Now().UTC())
}

// Getyyyy get the format of time yyyy for node uses
func Getyyyy(t time.Time) string {
	y, _, _ := t.Date()
	return strconv.Itoa(y)
}

// Getyyyymm get the format of time yyyymm for node uses
func Getyyyymm(t time.Time) string {
	_, m, _ := t.Date()
	mm := strconv.Itoa(int(m))
	if len([]rune(mm)) == 1 {
		mm = "0" + mm
	}
	return Getyyyy(t) + mm
}

// Getyyyymmdd get the format of time yyyymmdd for node uses
func Getyyyymmdd(t time.Time) string {
	_, _, d := t.Date()
	dd := strconv.Itoa(d)
	if len([]rune(dd)) == 1 {
		dd = "0" + dd
	}
	return Getyyyymm(t) + dd
}

// Getyyyyww get the format of time yyyyww for node uses
func Getyyyyww(t time.Time) string {
	_, w := t.ISOWeek()
	ww := strconv.Itoa(w)
	if len([]rune(ww)) == 1 {
		ww = "0" + ww
	}
	return Getyyyy(t) + "_week_" + ww
}

// Getyyyymmddhh gets the format of time yyyymmddhh
func Getyyyymmddhh(t time.Time) string {
	hh := fmt.Sprintf("%02d", t.Hour())
	return Getyyyymmdd(t) + hh
}

// TodayMidnight returns UTC 0:00 am's timestamp with user's time zone offset
func TodayMidnight(timestamp int64, userTimeZoneOffset int) time.Time {
	t := time.Unix(timestamp+int64(userTimeZoneOffset), 0).UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// LocationTaipei returns the *time.Location of "Asia/Taipei"
func LocationTaipei() (*time.Location, error) {
	return time.LoadLocation("Asia/Taipei")
}

// NowTaipei returns time structure with CST (Taipei time zone) offset
func NowTaipei() (time.Time, error) {
	loc, err := LocationTaipei()
	if err != nil {
		return time.Time{}, err
	}
	return timeNow().In(loc), nil
}

// Parse This function is for us to general use
// According to Roy's opinion, this format is more freindly for users.
func Parse(value string) (time.Time, error) {
	return time.Parse(template, value)
}

// ConvertFormat This function parse value using from and return time format using to
func ConvertFormat(from, to, value string) (string, error) {
	t, err := time.Parse(from, value)
	if err != nil {
		return "", err
	}

	return t.Format(to), nil
}

// Time is the time.Time alias for lit's models in order to
// use lit's time format for json's Marshal and Unmarshal.
// ref: https://www.cnblogs.com/xiaofengshuyu/p/5664654.html
type Time time.Time

// UnmarshalJSON customizes Time's json.Unmarshal
func (t *Time) UnmarshalJSON(data []byte) (err error) {
	now, err := time.ParseInLocation(`"`+litTimeFormat+`"`, string(data), time.UTC)
	*t = Time(now)
	return
}

// MarshalJSON customizes Time's json.Marshal
func (t Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(litTimeFormat)+2)
	b = append(b, '"')
	b = time.Time(t).UTC().AppendFormat(b, litTimeFormat)
	b = append(b, '"')
	return b, nil
}

// String print time with lit's time format
func (t *Time) String() string {
	return time.Time(*t).UTC().Format(litTimeFormat)
}

// GetPtr gets the pointer of Time struct.
// This method is used for field type `*Time` in structs to make `omitempty`
// in json marshaling work.
func (t Time) GetPtr() *Time {
	return &t
}

// DayTimeSlot defines a time slot duration in a day, which should be in
// 00h00m00s ~ 23h59h59s and EndTime should be after StartTime in a day
type DayTimeSlot struct {
	StartTime time.Duration
	EndTime   time.Duration
}

// Check checks if DayTimeSlot is reasonable or not
func (ts *DayTimeSlot) Check() error {
	start := ts.StartTime.Seconds()
	end := ts.EndTime.Seconds()
	if 0 <= start && start <= 86400 && 0 <= end && end <= 86400 && start <= end {
		return nil
	}
	return ErrInvalidDayTimeSlot
}

// GetTaipeiLocMidnightBeforeToday get the midnight time before today, if count is zero return today start's time
func GetTaipeiLocMidnightBeforeToday(count int64) (time.Time, error) {
	current, err := NowTaipei()
	if err != nil {
		return time.Time{}, err
	}
	result := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, current.Location())
	return result.AddDate(0, 0, int(-count)), nil
}
