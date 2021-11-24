package time

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnixMillis(t *testing.T) {
	ti := time.Date(2018, time.January, 2, 8, 28, 27, 559000000, time.UTC)
	eti := UnixMillis(int64(1514881707559)).In(time.UTC)
	assert.Equal(t, ti, eti, "Test UnixMillis")
}

func TestTodayMidnight(t *testing.T) {
	tests := []struct {
		Description                   string
		Timestamp                     int64
		UserTimeZoneOffset            int
		ExpetedTodayMidnightTimestamp int64
		ExpetedTodayMidnightY         int
		ExpetedTodayMidnightM         int
		ExpetedTodayMidnightD         int
	}{
		{
			"timestamp is in the same day with user time zone's today midnight timestamp",
			int64(1471315492),
			28800,
			int64(1471305600),
			2016,
			8,
			16,
		},
		{
			"timestamp is in the different day with user time zone's today midnight timestamp",
			int64(1471366800),
			28800,
			int64(1471392000),
			2016,
			8,
			17,
		},
	}
	for _, test := range tests {
		r := TodayMidnight(test.Timestamp, test.UserTimeZoneOffset)
		assert.Equal(t, test.ExpetedTodayMidnightTimestamp, r.Unix(), test.Description)
		assert.Equal(t, test.ExpetedTodayMidnightY, r.Year(), test.Description)
		assert.Equal(t, test.ExpetedTodayMidnightM, int(r.Month()), test.Description)
		assert.Equal(t, test.ExpetedTodayMidnightD, r.Day(), test.Description)
	}
}

func TestLegactime(t *testing.T) {
	testTime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	testTimeSingle := time.Date(2009, time.January, 8, 23, 0, 0, 0, time.UTC)

	// Test
	yyyy := Getyyyy(testTime)
	assert.Equal(t, "2009", yyyy, "test yyyy")

	yyyymm := Getyyyymm(testTime)
	assert.Equal(t, "200911", yyyymm, "test yyyymm")

	yyyymmSingle := Getyyyymm(testTimeSingle)
	assert.Equal(t, "200901", yyyymmSingle, "test yyyymm single")

	yyyymmdd := Getyyyymmdd(testTime)
	assert.Equal(t, "20091110", yyyymmdd, "test yyyymmdd")

	yyyymmddSingle := Getyyyymmdd(testTimeSingle)
	assert.Equal(t, "20090108", yyyymmddSingle, "test yyyymmdd single")

	_, ww := testTime.ISOWeek()
	yyyyww := Getyyyyww(testTime)
	assert.Equal(t, "2009_week_"+strconv.Itoa(ww), yyyyww, "test yyyyww")

	_, wwSingle := testTimeSingle.ISOWeek()
	yyyywwSingle := Getyyyyww(testTimeSingle)
	assert.Equal(t, "2009_week_0"+strconv.Itoa(wwSingle), yyyywwSingle, "test yyyyww single")
}

func TestNowTaipei(t *testing.T) {
	uloc, err := time.LoadLocation("UTC")
	assert.Nil(t, err, nil, "time.LoadLocation")
	utc := time.Now().In(uloc)
	nowTpe, err := NowTaipei()
	assert.Nil(t, err, nil, "NowTaipei")

	tests := []struct {
		Description       string
		ExpetedTaipeiHour int
		ActualTaipeiHour  int
	}{
		{
			"current hour is the same with Taipei hour",
			nowTpe.Hour(),
			utc.Add(8 * time.Hour).Hour(),
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.ExpetedTaipeiHour, test.ActualTaipeiHour, test.Description)
	}
}

func TestParseISO8601String(t *testing.T) {
	tests := []struct {
		duration  string
		expResult Duration
	}{
		{
			duration:  "",
			expResult: Duration{},
		},
		{
			duration:  "P0Y0M0D",
			expResult: Duration{},
		},
		{
			duration:  "P3M",
			expResult: Duration{Months: 3},
		},
		{
			duration:  "P30D",
			expResult: Duration{Days: 30},
		},
	}

	for _, test := range tests {
		d, err := ParseISO8601String(test.duration)
		assert.NoError(t, err)
		assert.Equal(t, test.expResult, *d)
	}
}

func getTWLocation() (*time.Location, error) {
	return time.LoadLocation("Asia/Taipei")
}

func TestParse(t *testing.T) {
	mockTimeString := "2017-10-01 00:00:00 (GMT+0800)"

	loc, err := getTWLocation()
	assert.NoError(t, err)
	res, err := Parse(mockTimeString)
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2017, time.October, 1, 0, 0, 0, 0, loc).Unix(), res.Unix())
}

func TestConvertFormat(t *testing.T) {
	mockTimeString := "2017/10/01"

	res, err := ConvertFormat(ParseFormatYYYYMMDDWithSlash, ParseFormatYYYYMMDDWithHyphen, mockTimeString)
	assert.NoError(t, err)
	assert.Equal(t, "2017-10-01", res)
}

func TestLitTimeFormat(t *testing.T) {
	t1 := Time(time.Unix(1234567890, 0).UTC())
	b, err := json.Marshal(t1)
	assert.NoError(t, err)
	assert.Equal(t, `"2009-02-13T23:31:30.000000"`, string(b))
	assert.Equal(t, "2009-02-13T23:31:30.000000", t1.String())

	t2 := Time{}
	err = json.Unmarshal(b, &t2)
	assert.NoError(t, err)
	assert.Equal(t, t1, t2)
	assert.Equal(t, t1.String(), t2.String())
}

func TestLitTimeGetPtr(t *testing.T) {
	t1 := Time(time.Unix(1234567890, 0).UTC())
	assert.IsType(t, &Time{}, t1.GetPtr())
}

func TestGetTaipeiLocMidnightBeforeToday(t *testing.T) {
	loc, err := getTWLocation()
	assert.NoError(t, err)
	timeNow = func() time.Time {
		return time.Date(2017, time.October, 1, 17, 17, 17, 17, loc)
	}
	assert.NoError(t, err)
	tm, err := GetTaipeiLocMidnightBeforeToday(0)
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2017, time.October, 1, 0, 0, 0, 0, loc), tm)
	tm, err = GetTaipeiLocMidnightBeforeToday(1)
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2017, time.September, 30, 0, 0, 0, 0, loc), tm)
}
