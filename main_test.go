package main

import (
	"reflect"
	"testing"
)

func TestMTLDRead(t *testing.T) {
	t.Skip()
}

func TestMTLDWrite(t *testing.T) {
	t.Skip()
}

func TestMTLDAdd(t *testing.T) {

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
