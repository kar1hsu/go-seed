package management

import (
	"github.com/kar1hsu/go-seed/services/system/internal/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"testing"
)

func TestSettingValues(t *testing.T) {
	for _, tt := range []struct {
		kind, value, want string
		valid             bool
	}{
		{"text", "", "", true}, {"text", "line\n文本", "line\n文本", true},
		{"integer", "9223372036854775807", "9223372036854775807", true},
		{"integer", "-9223372036854775808", "-9223372036854775808", true},
		{"integer", "9223372036854775808", "", false}, {"integer", "1.5", "", false}, {"integer", "01", "", false}, {"integer", "-0", "", false},
		{"boolean", "false", "false", true}, {"boolean", "TRUE", "", false},
		{"json", " { \"limit\": 9007199254740993 } ", `{"limit":9007199254740993}`, true},
		{"json", "{broken}", "", false}, {"json", "{} {}", "", false},
		{"secret", "value", "", false}, {"text", strings.Repeat("中", 1366), "", false},
	} {
		got, err := settingValue(tt.kind, SettingInput{Name: "name", Value: tt.value})
		if tt.valid && (err != nil || got != tt.want) || !tt.valid && status.Code(err) != codes.InvalidArgument {
			t.Fatalf("%s value validation: got %q %v", tt.kind, got, err)
		}
	}
}
func TestDictionaryValidation(t *testing.T) {
	for _, value := range []string{"order.status", "x", "site_title", "a.b2"} {
		if !validSettingCode(value) {
			t.Fatal(value)
		}
	}
	for _, value := range []string{"", "A", "a..b", "1x", "a.", strings.Repeat("a", 65)} {
		if validSettingCode(value) {
			t.Fatal(value)
		}
	}
	valid := model.DictionaryItem{Value: "pending", Label: "待处理", Sort: 10, Enabled: true}
	if err := validateDictionary(DictionaryInput{Name: "订单状态", Items: []model.DictionaryItem{valid}}); err != nil {
		t.Fatal(err)
	}
	for _, items := range [][]model.DictionaryItem{
		{valid, valid}, {{Value: " ", Label: "blank"}}, {{Value: "x", Label: "\n"}}, {{Value: "x ", Label: "space"}}, {{Value: "x", Label: "x", Sort: -1}}, make([]model.DictionaryItem, 201),
	} {
		if status.Code(validateDictionary(DictionaryInput{Name: "状态", Items: items})) != codes.InvalidArgument {
			t.Fatal("accepted invalid dictionary")
		}
	}
}
