package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
)

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
		})
	}
}

func TestConfig_Load(t *testing.T) {
	t.Skip()
	tests := []struct {
		name string
		c    *Config
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.Load()
		})
	}
}

func TestTLDef_SetCaptureTimes(t *testing.T) {
	// layout := "Jan 2 2006 15:04:05 -0700 MST"
	loc := time.Local

	day1 := time.Date(2020, 5, 27, 14, 0, 0, 0, loc)
	day2 := time.Date(2020, 5, 28, 14, 0, 0, 0, loc)

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
			want:    []time.Time{day1},
		},
		{name: "May28",
			tld: &TLDef{
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
			},
			day:     day2,
			wantErr: false,
			want:    []time.Time{day2},
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

	for _, tc := range tests {
		// log.Printf("%s: %+v\n", tc.name, tc.mtld)

		if err := tc.mtld.Append(&tc.newTLD); err != nil {
			t.Fatal(err)
		}
		// log.Printf("%s, after Add: %+v\n", tc.name, tc.mtld)

		if !reflect.DeepEqual(tc.expected, tc.mtld) {
			t.Errorf("%s expected %+v, got %+v", tc.name, tc.expected, tc.mtld)
		}

	}
}

func Test_SSDayInfo_Get(t *testing.T) {
	var time1 time.Time
	var err error

	if time1, err = time.Parse("Jan 2 2006", "May 27 2020"); err != nil {
		t.Fatalf("time.Parse: %v", err)
	}

	tests := []struct {
		name    string
		lat     float64
		long    float64
		t       time.Time
		wantErr bool
		want    Solar
	}{
		{name: "valid",
			lat:  40.437787,
			long: -121.5360307,
			t:    time1,
			want: Solar{
				SunriseUTC:   "12:39:41 PM",
				SunsetUTC:    "3:27:15 AM",
				SolarNoonUTC: "8:03:28 PM",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *Solar
			var err error
			if got, err = got.GetSolar(tt.t); (err != nil) != tt.wantErr {
				t.Errorf("SSDayInfo.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("SSDayInfo.Get(), got %+v, want %+v", *got, tt.want)
			}
		})
	}
}
