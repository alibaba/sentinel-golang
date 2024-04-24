package gray

import (
	"fmt"
	"testing"
)

func TestRewriteAddress(t *testing.T) {
	for i := 1; i <= 10; i++ {
		addr, port, err := getRewriteHost("gin-server-a.default.svc", "80", "gray")
		if err != nil {
			t.Error(err)
		}

		fmt.Printf("[TestGetEndpoints] load balance test round %d, rewrite address: %s:%s\n", i, addr, port)
	}
}
