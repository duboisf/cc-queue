package kitty_test

import (
	"testing"

	"github.com/duboisf/cc-queue/internal/kitty"
)

func TestLayoutManager_ImplementsFullTabber(t *testing.T) {
	var _ kitty.FullTabber = &kitty.LayoutManager{}
}
