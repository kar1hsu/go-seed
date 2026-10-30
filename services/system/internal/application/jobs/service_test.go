package jobs

import "testing"

func TestCSVFormulaNeutralization(t *testing.T) {
	for _, value := range []string{"=1+1", "+SUM(A1)", "-1+2", "@SUM(A1)", "  =cmd", "\tvalue", "\rvalue"} {
		if got := CSVCell(value); got != "'"+value {
			t.Errorf("unsafe cell %q", got)
		}
	}
	if CSVCell("正常文本") != "正常文本" {
		t.Fatal("ordinary text changed")
	}
}
