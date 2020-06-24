package main

import (
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"github.com/julienschmidt/httprouter"
	"github.com/peterpla/timelapse/capture"
)

func TestMain(m *testing.M) {
	funcframework.RegisterHTTPFunction("/", capture.WebcamImage)

	// Use PORT environment variable, or default to 8080.
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	go startFramework(port) // call ListenAndServe from a separate go routine so main can listen for signals

	exitcode := m.Run()
	os.Exit(exitcode)
}

// startFramework starts funcframework which calls ListenAndServe
func startFramework(port string) {
	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}

func Test_newServer(t *testing.T) {
	t.Skip()
	tests := []struct {
		name string
		s    *server
		want httprouter.Handle
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.handleHome(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newServer() got %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_server_handleHome(t *testing.T) {
	t.Skip()
	tests := []struct {
		name string
		s    *server
		want httprouter.Handle
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.handleHome(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("server.handleHome() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_server_handleNew(t *testing.T) {
	t.Skip()
	tests := []struct {
		name string
		s    *server
		want httprouter.Handle
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.handleNew(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("server.handleNew() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_server_initTemplates(t *testing.T) {
	t.Skip()
	type args struct {
		dir string
		ext string
	}
	tests := []struct {
		name string
		s    *server
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.initTemplates(tt.args.dir, tt.args.ext)
			// if got := tt.s.initTemplates(); !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("server.initTemplates() got %v, want %v", got, tt.want)
			// }
		})
	}
}

func TestConfig_Load(t *testing.T) {
	t.Skip()
	tests := []struct {
		name string
		c    Config
		want Config
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ttc := &tt.c
			ttc.Load()
			if !reflect.DeepEqual(tt.c, tt.want) {
				t.Errorf("Config.Load() got %v, want %v", tt.c, tt.want)
			}
		})
	}
}

func TestNewTLDef(t *testing.T) {
	t.Skip()
	tests := []struct {
		name string
		want TLDef
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newTLDef()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config.Load() got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTLDef_SetCaptureTimes(t *testing.T) {
	// layout := "Jan 2 2006 15:04:05 -0700 MST"
	loc := time.Local

	day1 := time.Date(2020, 5, 27, 0, 0, 0, 0, loc)
	day1Capture := []time.Time{
		time.Date(2020, 5, 27, 5, 39, 41, 0, loc),  // Sunrise
		time.Date(2020, 5, 27, 13, 3, 28, 0, loc),  // SolarNoon
		time.Date(2020, 5, 27, 20, 27, 15, 0, loc), // Sunset
	}

	day2 := time.Date(2020, 5, 28, 0, 0, 0, 0, loc)
	day2Capture := []time.Time{
		time.Date(2020, 5, 28, 5, 39, 9, 0, loc),  // Sunrise
		time.Date(2020, 5, 28, 13, 3, 36, 0, loc), // SolarNoon
		time.Date(2020, 5, 28, 20, 28, 2, 0, loc), // Sunset
	}

	tests := []struct {
		name    string
		tld     *TLDef
		day     time.Time
		wantErr bool
		want    []time.Time
	}{
		{name: "May27",
			tld: &TLDef{
				Name:         "Kohm Yah-man-yeh",
				URL:          "https://www.nps.gov/webcams-lavo/kyvc_webcam1.jpg?1589316288166",
				Latitude:     40.437787,
				Longitude:    -121.5360307,
				FirstTime:    false,
				FirstSunrise: true,
				LastTime:     false,
				LastSunset:   true,
				Additional:   1,
				FolderPath:   "/Volumes/ExtFiles/OneDrive/Pictures/Timelapse/Kohm-Yah-mah-nee",
			},
			day:     day1,
			wantErr: false,
			want:    day1Capture,
		},
		{name: "May28",
			tld: &TLDef{ // includes previous day's capture times
				Name:         "Kohm Yah-man-yeh",
				URL:          "https://www.nps.gov/webcams-lavo/kyvc_webcam1.jpg?1589316288166",
				Latitude:     40.4375635,
				Longitude:    -121.5357176,
				FirstTime:    false,
				FirstSunrise: true,
				LastTime:     false,
				LastSunset:   true,
				Additional:   1,
				FolderPath:   "/Volumes/ExtFiles/OneDrive/Pictures/Timelapse/Kohm-Yah-mah-nee",
				CaptureTimes: day1Capture,
			},
			day:     day2,
			wantErr: false,
			want:    day2Capture,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tld.SetCaptureTimes(tt.day); (err != nil) != tt.wantErr {
				t.Errorf("TLDef.SetCaptureTimes() error = %v, wantErr %v", err, tt.wantErr)
			}
			got := tt.tld.CaptureTimes
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TLDef.SetCaptureTimes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTLDef_SetFirstCapture(t *testing.T) {
	loc := time.Local
	sunrise := time.Date(2020, 5, 27, 5, 39, 41, 0, loc)
	// firstTime := time.Date(2020, 5, 27, 6, 0, 1, 0, loc)

	tests := []struct {
		name    string
		tld     TLDef
		wantErr bool
		want    []time.Time
	}{
		{name: "sunrise",
			tld: TLDef{
				FirstTime:    false,
				FirstSunrise: true,
				SunriseUTC:   sunrise,
			},
			wantErr: false,
			want:    []time.Time{sunrise},
		},
		// {name: "first time",
		// 	tld: TLDef{
		// 		FirstTime:    true,
		// 		FirstSunrise: false,
		// 	},
		// 	wantErr: false,
		// 	want:    []time.Time{firstTime},
		// },
		{name: "both",
			tld: TLDef{
				FirstTime:    true,
				FirstSunrise: true,
			},
			wantErr: true,
		},
		{name: "neither",
			tld: TLDef{
				FirstTime:    false,
				FirstSunrise: false,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tld.SetFirstCapture(); (err != nil) != tt.wantErr {
				t.Errorf("TLDef.SetFirstCapture() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				got := tt.tld.CaptureTimes
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("TLDef.SetFirstCapture() got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestTLDef_SetAdditional(t *testing.T) {
	loc := time.Local
	sunrise := time.Date(2020, 5, 27, 5, 39, 41, 0, loc)   // Sunrise
	solarNoon := time.Date(2020, 5, 27, 13, 3, 28, 0, loc) // SolarNoon
	sunset := time.Date(2020, 5, 27, 20, 27, 15, 0, loc)   // Sunset

	addTwoTheFirst := time.Date(2020, 5, 27, 10, 35, 32, 0, loc)
	addTwoTheSecond := time.Date(2020, 5, 27, 15, 31, 23, 0, loc)

	addThreeTheFirst := time.Date(2020, 5, 27, 9, 21, 34, 0, loc)
	addThreeTheSecond := solarNoon
	addThreeTheThird := time.Date(2020, 5, 27, 16, 45, 21, 0, loc)

	addFourTheFirst := time.Date(2020, 5, 27, 8, 37, 11, 0, loc)
	addFourTheSecond := time.Date(2020, 5, 27, 11, 34, 41, 0, loc)
	addFourTheThird := time.Date(2020, 5, 27, 14, 32, 11, 0, loc)
	addFourTheFourth := time.Date(2020, 5, 27, 17, 29, 41, 0, loc)

	addFiveTheFirst := time.Date(2020, 5, 27, 8, 7, 36, 0, loc)
	addFiveTheSecond := time.Date(2020, 5, 27, 10, 35, 31, 0, loc)
	addFiveTheThird := solarNoon
	addFiveTheFourth := time.Date(2020, 5, 27, 15, 31, 23, 0, loc)
	addFiveTheFifth := time.Date(2020, 5, 27, 17, 59, 18, 0, loc)

	tests := []struct {
		name string
		tld  TLDef
		want []time.Time
	}{
		{name: "add 0",
			tld: TLDef{
				Name:         "test",
				FirstSunrise: true,
				FirstTime:    false,
				LastTime:     false,
				LastSunset:   true,
				Additional:   0,
				SunriseUTC:   sunrise.In(time.UTC),
				SunsetUTC:    sunset.In(time.UTC),
				CaptureTimes: []time.Time{sunrise},
			},
			want: []time.Time{sunrise, sunset},
		},
		{name: "add 1", // always add solar noon when adding 1 capture
			tld: TLDef{
				Name:         "test",
				FirstSunrise: true,
				FirstTime:    false,
				LastTime:     false,
				LastSunset:   true,
				Additional:   1,
				SunriseUTC:   sunrise.In(time.UTC),
				SolarNoonUTC: solarNoon.In(time.UTC),
				SunsetUTC:    sunset.In(time.UTC),
				CaptureTimes: []time.Time{sunrise},
			},
			want: []time.Time{sunrise, solarNoon, sunset},
		},
		{name: "add 2",
			tld: TLDef{
				Name:         "test",
				FirstSunrise: true,
				FirstTime:    false,
				LastTime:     false,
				LastSunset:   true,
				Additional:   2,
				SunriseUTC:   sunrise.In(time.UTC),
				SunsetUTC:    sunset.In(time.UTC),
				CaptureTimes: []time.Time{sunrise},
			},
			want: []time.Time{sunrise, addTwoTheFirst, addTwoTheSecond, sunset},
		},
		{name: "add 3",
			tld: TLDef{
				Name:         "test",
				FirstSunrise: true,
				FirstTime:    false,
				LastTime:     false,
				LastSunset:   true,
				Additional:   3,
				SunriseUTC:   sunrise.In(time.UTC),
				SolarNoonUTC: solarNoon.In(time.UTC),
				SunsetUTC:    sunset.In(time.UTC),
				CaptureTimes: []time.Time{sunrise},
			},
			want: []time.Time{sunrise, addThreeTheFirst, addThreeTheSecond, addThreeTheThird, sunset},
		},
		{name: "add 4",
			tld: TLDef{
				Name:         "test",
				FirstSunrise: true,
				FirstTime:    false,
				LastTime:     false,
				LastSunset:   true,
				Additional:   4,
				SunriseUTC:   sunrise.In(time.UTC),
				SunsetUTC:    sunset.In(time.UTC),
				CaptureTimes: []time.Time{sunrise},
			},
			want: []time.Time{sunrise, addFourTheFirst, addFourTheSecond, addFourTheThird, addFourTheFourth, sunset}, // the last capture time is set seprately
		},
		{name: "add 5",
			tld: TLDef{
				Name:         "test",
				FirstSunrise: true,
				FirstTime:    false,
				LastTime:     false,
				LastSunset:   true,
				Additional:   5,
				SunriseUTC:   sunrise.In(time.UTC),
				SolarNoonUTC: solarNoon.In(time.UTC),
				SunsetUTC:    sunset.In(time.UTC),
				CaptureTimes: []time.Time{sunrise},
			},
			want: []time.Time{sunrise, addFiveTheFirst, addFiveTheSecond, addFiveTheThird, addFiveTheFourth, addFiveTheFifth, sunset}, // the last capture time is set seprately
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tld.SetAdditional()
			got := tt.tld.CaptureTimes
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TLDef.SetAdditional() got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTLDef_SetLastCapture(t *testing.T) {
	loc := time.Local
	sunset := time.Date(2020, 5, 27, 20, 27, 15, 0, loc) // Sunset
	// lastTime := time.Date(2020, 5, 27, 21, 0, 1, 0, loc)

	tests := []struct {
		name    string
		tld     TLDef
		wantErr bool
		want    []time.Time
	}{
		{name: "sunset",
			tld: TLDef{
				LastTime:   false,
				LastSunset: true,
				SunsetUTC:  sunset,
			},
			wantErr: false,
			want:    []time.Time{sunset},
		},
		// {name: "last time",
		// 	tld: TLDef{
		// 		LastTime:   true,
		// 		LastSunset: false,
		// 	},
		// 	wantErr: false,
		// 	want:    []time.Time{lastTime},
		// },
		{name: "both",
			tld: TLDef{
				LastTime:   true,
				LastSunset: true,
			},
			wantErr: true,
		},
		{name: "neither",
			tld: TLDef{
				LastTime:   false,
				LastSunset: false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.tld.SetLastCapture(); (err != nil) != tt.wantErr {
				t.Errorf("TLDef.SetLastCapture() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				got := tt.tld.CaptureTimes
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("TLDef.SetLastCapture() got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestTLDef_UpdateNextCapture(t *testing.T) {
	t.Skip()
	tests := []struct {
		name string
		tld  TLDef
		want TLDef
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tld.UpdateNextCapture()
			if !reflect.DeepEqual(tt.tld, tt.want) {
				t.Errorf("TLDef.UpdateNextCapture() got %v, want %v", tt.tld, tt.want)
			}
		})
	}
}

func TestTLDef_NextCaptureTime(t *testing.T) {
	t.Skip()
	tests := []struct {
		name string
		tld  TLDef
		want time.Time
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tld.NextCaptureTime(); got != tt.want {
				t.Errorf("TLDef.NextCaptureTime() got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTLDef_IsTimeForCapture(t *testing.T) {
	t.Skip()
	tests := []struct {
		name string
		tld  TLDef
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tld.IsTimeForCapture(); got != tt.want {
				t.Errorf("TLDef.IsTimeForCapture() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTLDef_TimeToSecond(t *testing.T) {
	t.Skip()
	tests := []struct {
		name     string
		testTime time.Time
		want     time.Time
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TimeToSecond(tt.testTime); got != tt.want {
				t.Errorf("TLDef.TimeToSecond() got %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newMasterTLDefs(t *testing.T) {
	t.Skip()
	tests := []struct {
		name string
		want *masterTLDefs
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newMasterTLDefs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newMasterTLDefs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_masterTLDefs_Read(t *testing.T) {
	t.Skip()
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		mtld    masterTLDefs
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mtld.Read(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("masterTLDefs.Read() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_masterTLDefs_Write(t *testing.T) {
	t.Skip()
	tests := []struct {
		name    string
		mtld    masterTLDefs
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.mtld.Write(); (err != nil) != tt.wantErr {
				t.Errorf("masterTLDefs.Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_masterTLDefs_Append(t *testing.T) {
	value1 := TLDef{
		Name:         "testName",
		URL:          "testURL",
		FirstTime:    true,
		FirstSunrise: false,
		LastTime:     true,
		LastSunset:   false,
		Additional:   1,
		FolderPath:   "testPath",
	}
	mtldWithValue1 := new(masterTLDefs)
	*mtldWithValue1 = append(*mtldWithValue1, value1)

	value2 := TLDef{
		Name:         "testName2",
		URL:          "testURL2",
		FirstTime:    false,
		FirstSunrise: true,
		LastTime:     false,
		LastSunset:   true,
		Additional:   2,
		FolderPath:   "testPath2",
	}
	mtldWithValue1AndValue2 := new(masterTLDefs)
	*mtldWithValue1AndValue2 = append(*mtldWithValue1AndValue2, value1, value2)

	type test struct {
		name     string
		mtld     *masterTLDefs
		newTLD   TLDef
		expected *masterTLDefs
	}

	var workingMTLD = new(masterTLDefs)

	tests := []test{
		{name: "add to empty",
			mtld:     workingMTLD,
			newTLD:   value1,
			expected: mtldWithValue1,
		},
		{name: "add second element",
			mtld:     workingMTLD,
			newTLD:   value2,
			expected: mtldWithValue1AndValue2,
		},
	}

	for _, tt := range tests {
		// log.Printf("%s: %+v\n", tt.name, tt.mtld)

		if err := tt.mtld.Append(&tt.newTLD); err != nil {
			t.Fatal(err)
		}
		// log.Printf("%s, after Add: %+v\n", tt.name, tt.mtld)

		if !reflect.DeepEqual(tt.expected, tt.mtld) {
			t.Errorf("%s expected %+v, got %+v", tt.name, tt.expected, tt.mtld)
		}

	}
}

func TestTLDef_GetSolarTimes(t *testing.T) {
	var testDate, testSunrise, testSunset, testSolarNoon time.Time
	var err error
	const (
		layoutDate        = "Jan 2 2006"
		layoutDateAndTime = "Jan 2 2006 3:04:05 PM"
	)

	if testDate, err = time.Parse(layoutDate, "May 27 2020"); err != nil {
		t.Fatalf("time.Parse: %v", err)
	}
	if testSunrise, err = time.Parse(layoutDateAndTime, "May 27 2020 12:39:41 PM"); err != nil {
		t.Fatalf("time.Parse: %v", err)
	}
	if testSunset, err = time.Parse(layoutDateAndTime, "May 28 2020 3:27:15 AM"); err != nil {
		t.Fatalf("time.Parse: %v", err)
	}
	if testSolarNoon, err = time.Parse(layoutDateAndTime, "May 27 2020 8:03:28 PM"); err != nil {
		t.Fatalf("time.Parse: %v", err)
	}

	tests := []struct {
		name    string
		tld     TLDef
		date    time.Time
		wantErr bool
		want    TLDef
	}{
		{name: "valid",
			tld: TLDef{
				URL:       "https://www.nps.gov/webcams-lavo/kyvc_webcam1.jpg?1589316288166",
				Latitude:  40.437787,
				Longitude: -121.5360307,
			},
			date:    testDate,
			wantErr: false,
			want: TLDef{
				SunriseUTC:   testSunrise,
				SunsetUTC:    testSunset,
				SolarNoonUTC: testSolarNoon,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if err = tt.tld.GetSolarTimes(tt.date); (err != nil) != tt.wantErr {
				t.Errorf("GetSolarTimes() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.tld.SunriseUTC != tt.want.SunriseUTC {
				t.Errorf("GetSolarTimes(), SunriseUTC got %+v, want %+v", tt.tld.SunriseUTC, tt.want.SunriseUTC)
			}
			if tt.tld.SunsetUTC != tt.want.SunsetUTC {
				t.Errorf("GetSolarTimes(), SunsetUTC got %+v, want %+v", tt.tld.SunsetUTC, tt.want.SunsetUTC)
			}
			if tt.tld.SolarNoonUTC != tt.want.SolarNoonUTC {
				t.Errorf("GetSolarTimes(), SolarNoonUTC got %+v, want %+v", tt.tld.SolarNoonUTC, tt.want.SolarNoonUTC)
			}
		})
	}
}

func TestTLDef_NewSSDayInfo(t *testing.T) {
	t.Skip()
	tests := []struct {
		name string
		tld  TLDef
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tld.IsTimeForCapture(); got != tt.want {
				t.Errorf("TLDef.NewSSDayInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSSDayInfo_buildQuery(t *testing.T) {
	t.Skip()
	tests := []struct {
		name string
		tld  TLDef
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tld.IsTimeForCapture(); got != tt.want {
				t.Errorf("SSDayInfo.buildQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTLDef_NewTimeZoneDB(t *testing.T) {
	t.Skip()
	tests := []struct {
		name string
		tld  TLDef
		want TimeZoneDB
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTimeZoneDB(&tt.tld)
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("%s got %+v, want %+v", tt.name, got, tt.want)
			}
		})
	}
}

func TestTLDef_SetWebcamTZ(t *testing.T) {
	t.Skip()
	tests := []struct {
		name string
		tld  TLDef
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tld.IsTimeForCapture(); got != tt.want {
				t.Errorf("TLDef.SetWebcamTZ() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeZoneDB_buildQuery(t *testing.T) {
	t.Skip()
	tests := []struct {
		name string
		tld  TLDef
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tld.IsTimeForCapture(); got != tt.want {
				t.Errorf("TimeZoneDB.buildQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}
