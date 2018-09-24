package mrgraphite

import (
	"testing"
)

var (
	q50 = NewQuantile("timing.globalQuantile", 50)
)

func TestQuantileEmpty(t *testing.T) {
	q95 := NewQuantile("timing.quantile1", 95)
	_, err := q95.GetValue()
	if err != EmptyListError {
		t.Errorf("wrong err %v", err)
	}
}

func TestQuantileLazy(t *testing.T) {
	q50.AddValue(5)
	q50.AddValue(9)
	q50.AddValue(7)

	val1, err := q50.GetValue()
	if err != nil {
		t.Fatalf("q50 error %v", err)
	}
	if val1 != 7 {
		t.Errorf("q50 GetValue is wrong(%d)", val1)
	}
	if q50.GetName() != "timing.globalQuantile" {
		t.Errorf("q50 GetName is wrong")
	}
	if q50.GetQVal() != 50 {
		t.Errorf("q50 GetQVal is wrong")
	}
}

func TestQuantiles(t *testing.T) {
	q95 := NewQuantile("timing.quantile1", 95)
	for i:=0; i<100; i++ {
		q95.AddValue( int64(i*3) )
	}

	var err error
	val1, err := q95.GetValue()
	if err != nil {
		t.Fatalf("q95 error %v", err)
	}
	if val1 != 95*3 {
		t.Errorf("q95 GetValue is wrong(%d)", val1)
	}
	if q95.GetName() != "timing.quantile1" {
		t.Errorf("q95 GetName is wrong")
	}
	if q95.GetQVal() != 95 {
		t.Errorf("q95 GetQVal is wrong")
	}
}
