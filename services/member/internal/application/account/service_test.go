package account

import "testing"

func TestInputRules(t *testing.T) {
	for _, v := range []string{"ab", "UPPER", "用户", "a b"} {
		if ValidUsername(v) {
			t.Errorf("accepted username %q", v)
		}
	}
	for _, v := range []string{"", "  ", "bad\nname", string([]byte{0xff})} {
		if ValidNickname(v) {
			t.Errorf("accepted nickname %q", v)
		}
	}
	if !ValidNickname("小明") || !ValidUsername("member_1") || !ValidPassword("long-enough-password") {
		t.Fatal("valid inputs rejected")
	}
	if ValidPassword("short") || ValidPassword("password-with\x00nul") {
		t.Fatal("invalid password accepted")
	}
}
