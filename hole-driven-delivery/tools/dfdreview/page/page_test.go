package page_test

import (
	"strings"
	"testing"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/page"
)

func TestWriteKeepsTheDataInsideItsScript(t *testing.T) {
	var b strings.Builder
	d := page.Data{Title: "A <b>plan</b>", Report: "</script><script>alert(1)</script>"}
	if err := page.Write(&b, d); err != nil {
		t.Fatal(err)
	}
	html := b.String()
	if strings.Count(html, "</script>") != strings.Count(html, "<script") {
		t.Errorf("the data closed a script element:\n%s", html)
	}
	if !strings.Contains(html, "<title>A &lt;b&gt;plan&lt;/b&gt;</title>") {
		t.Errorf("the title is not escaped:\n%s", html)
	}
}
