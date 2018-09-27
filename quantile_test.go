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

type CalcCase struct {
	qval   float64
	answer int64
	values []int64
}

func TestQuantileCalc(t *testing.T) {
	cases := []CalcCase{
		{0, 1, []int64{1}},
		{0, 1, []int64{6, 1}},
		{0, 1, []int64{7, 5, 1}},

		{25, 1, []int64{1}},
		{25, 1, []int64{6, 1}},
		{25, 1, []int64{7, 5, 1}},

		{50, 1, []int64{1}},
		{50, 6, []int64{6, 1}},
		{50, 5, []int64{7, 5, 1}},
		{50, 7, []int64{7, 5, 9, 1}},

		{100, 1, []int64{1}},
		{100, 6, []int64{6, 1}},
		{100, 7, []int64{7, 5, 1}},

		//13
		{1, 3, []int64{3, 6, 9, 12, 15}},
		{15, 3, []int64{3, 6, 9, 12, 15}},
		{50, 9, []int64{3, 6, 9, 12, 15}},
		{79, 12, []int64{3, 6, 9, 12, 15}},
		{80, 15, []int64{3, 6, 9, 12, 15}},
		{81, 15, []int64{3, 6, 9, 12, 15}},
		{95, 15, []int64{3, 6, 9, 12, 15}},
	}

	for cn, c := range cases {
		q := NewQuantile("timing.qtestcase", c.qval)
		for _, v := range c.values {
			q.AddValue(v)
		}

		val1, err := q.GetValue()
		if err != nil {
			t.Fatalf("q[%d] error %v", cn, err)
		}
		if val1 != c.answer {
			t.Errorf("q[%d] GetValue is wrong. got=%d want=%d", cn, val1, c.answer)
		}
		if qv := q.GetQVal(); qv != c.qval {
			t.Errorf("q[%d] GetQVal is wrong. got=%.2f want=%.2f", cn, qv, c.qval)
		}
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
	for i := 0; i < 100; i++ {
		q95.AddValue(int64(i * 3))
	}

	var err error
	val1, err := q95.GetValue()
	if err != nil {
		t.Fatalf("q95 error %v", err)
	}
	if val1 != 95*3 {
		t.Errorf("q95 GetValue is wrong, got=%d want=%d", val1, 95*3)
	}
	if q95.GetName() != "timing.quantile1" {
		t.Errorf("q95 GetName is wrong")
	}
	if q95.GetQVal() != 95 {
		t.Errorf("q95 GetQVal is wrong")
	}
}
