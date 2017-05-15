package anomaly

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

type AnomalyWindow [2]string

func TestDetect(t *testing.T) {
	b, err := ioutil.ReadFile("./testdata/expected.json")
	if err != nil {
		t.Error(err)
	}

	results := make(map[string][]AnomalyWindow)
	err = json.Unmarshal(b, &results)
	if err != nil {
		t.Error(err)
	}

	filepath.Walk("./testdata/", func(p string, i os.FileInfo, err error) error {
		if err != nil {
			t.Error(err)
		}

		if i.IsDir() || filepath.Ext(p) != ".csv" {
			return nil
		}

		t.Run(p, func(t *testing.T) {
			fmt.Println("Running test with dataset ", p)
			tval, data, err := ReadDataFile(p)
			if err != nil {
				t.Error(err)
				return
			}
			anom := Detect(data)
			fmt.Println("Anom len ", len(anom))
			ws := GetAnomalyWindows(anom)
			for _, w := range ws {
				fmt.Println(w, tval[w[0]], tval[w[1]])
			}
		})
		return nil
	})
}

func GetAnomalyWindows(a []int) [][]int {
	ws := make([][]int, 0)
	start := a[0]
	for i := range a {
		if i > 0 && a[i]-a[i-1] > 1 {
			ws = append(ws, []int{start, a[i-1]})
			start = a[i]
		}
	}
	ws = append(ws, []int{start, a[len(a)-1]})
	return ws
}

func ReadDataFile(path string) ([]string, []float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	all, err := r.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	t := make([]string, len(all))
	d := make([]float64, len(all))
	for i, v := range all[1:] {
		t[i] = v[0]
		d[i], err = strconv.ParseFloat(v[1], 64)
		if err != nil {
			return nil, nil, err
		}
	}
	return t, d, nil
}
