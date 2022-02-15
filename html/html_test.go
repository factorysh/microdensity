package html

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPage(t *testing.T) {
	p := &Page{
		Detail: "Bob",
		Inner: `<html><body>Beuha
		</body></html>`,
	}
	w := httptest.NewRecorder()
	err := p.Render(w)
	assert.NoError(t, err)
	w.Flush()
	assert.Equal(t, 200, w.Code)

	h := w.Body.String()
	assert.True(t, len(h) > 0)
	fmt.Println("l", len(h))
	fmt.Println(h)
	assert.True(t, strings.Contains(h, "µdensity - Bob"))
}
