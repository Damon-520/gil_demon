package base62x

import "testing"

func Test_IntToBase62(t *testing.T) {
	t.Log(IntToBase62(123456789012)) // 2Al26WS
	t.Log(Base62ToInt("2Al26WS"))    // 123456789012
}

func Test_Base62ToInt(t *testing.T) {
	_, err := Base62ToInt("122sfs8KJDsfsDKSLDLS")
	if err != nil {
		t.Log(err) // integer overflow
	}

	_, err = Base62ToInt("&sdwe20")
	if err != nil {
		t.Log(err) // invalid base62 string
	}

	n, err := Base62ToInt("1234567890")
	if err != nil {
		t.Log(err)
	}
	t.Log(n) // 13984563473216086
}
