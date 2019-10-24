package conversion

import (
	"testing"
)

func TestFileSizeBytesToFloat(t *testing.T) {

	tests := []struct {
		a, b   int
		output float64
	}{
		{2, 2, 0},
	}

	t.Run("int to float", func(t *testing.T) {
		var floatVals = make([]float64, len(tests))
		for _, test := range tests {
			floatVal := convertIntCalcToFloat(test.a, test.b)
			t.Log(floatVal)
			if floatVal != test.output {
				t.Errorf("Expected %v got %v", test.output, floatVal)
				return
			}
			floatVals = append(floatVals, floatVal)
		}
		t.Logf("floats: %v", floatVals)
	})
}

func convertIntCalcToFloat(a, b int) float64 {
	return float64(a >> b)
}
