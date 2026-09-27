package dfdreview_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/internal/gittest"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/page"
)

// shop is the example module of the end-to-end tests at step 1 or 2. Step 2 doubles the cost
// of an item, in a method that cart.Total calls, and puts a type error in receipt/footer.go.
func shop(step int) map[string]string {
	cost, footer := "it.Price", `const Footer = "thank you"`
	if step == 2 {
		cost, footer = "2 * it.Price", `const Footer int = "thank you"`
	}
	return map[string]string{
		"go.mod": "module example.com/shop\n\ngo 1.22\n",
		"docs/flow.dfd": "{Customer}\n> order\n[1. Price the order\n (cart.Total)]\n> total\n" +
			"[2. Print the receipt\n (receipt.Print)]\n# type: order = cart.Order\n# type: total = int\n",
		"docs/vocabulary.md": "- order: the items a customer buys.\n",
		"cart/cart.go": "package cart\n\n// Order is the items a customer buys.\ntype Order struct{ Items []Item }\n\n" +
			"// Item is one line of an order.\ntype Item struct {\n\tName  string\n\tPrice int\n}\n\n" +
			"// Total sums the cost of the order's items.\nfunc Total(o Order) int {\n\tsum := 0\n" +
			"\tfor _, it := range o.Items {\n\t\tsum += it.cost()\n\t}\n\treturn sum\n}\n\n" +
			"func (it Item) cost() int { return " + cost + " }\n",
		"cart/cart_test.go": "package cart\n\nimport \"testing\"\n\nfunc TestTotal(t *testing.T) {\n" +
			"\tif Total(Order{Items: []Item{{Price: 2}}}) == 0 {\n\t\tt.Error(\"zero\")\n\t}\n}\n",
		"receipt/receipt.go": "package receipt\n\nimport (\n\t\"fmt\"\n\n\t\"example.com/shop/cart\"\n)\n\n" +
			"// Print writes the order's total.\n" +
			"func Print(o cart.Order) string { return fmt.Sprintf(\"total %d\", cart.Total(o)) }\n",
		"receipt/footer.go": "package receipt\n\n// Footer ends each receipt.\n" + footer + "\n",
	}
}

// shopPage writes the shop at step 1, commits it as the base, writes step 2, and returns the
// options that build its page with the dfd command dfd.
func shopPage(t *testing.T, dfd string) dfdreview.Options {
	t.Helper()
	g := gittest.New(t)
	g.Write(shop(1))
	base := g.Commit("base")
	g.Write(shop(2))
	docs := filepath.Join(g.Dir, "docs")
	return dfdreview.Options{
		Design:     filepath.Join(docs, "flow.dfd"),
		Base:       base,
		Vocabulary: filepath.Join(docs, "vocabulary.md"),
		DFD:        dfd,
		Out:        filepath.Join(t.TempDir(), "index.html"),
	}
}

func TestBuildShop(t *testing.T) {
	dfd := filepath.Join(t.TempDir(), "dfd")
	if err := os.WriteFile(dfd, []byte(echo), 0o755); err != nil {
		t.Fatal(err)
	}
	opts := shopPage(t, dfd)
	if err := dfdreview.Build(opts); err != nil {
		t.Fatal(err)
	}
	d := readPage(t, opts.Out)
	for key, want := range map[string]bool{
		"cart.Total":    true,  // it calls Item.cost, which changed
		"receipt.Print": true,  // it calls cart.Total, whose reach changed
		"cart.Order":    false, // a type keeps its own rule
	} {
		if got := d.Decls[key].Changed; got != want {
			t.Errorf("%s changed: %v, want %v", key, got, want)
		}
	}
	for _, path := range []string{"cart/cart.go", "cart/cart_test.go", "receipt/receipt.go", "receipt/footer.go"} {
		if f, ok := d.Files[path]; !ok || f.Before == nil || f.After == nil {
			t.Errorf("the page lacks a version of %s", path)
		}
	}
	has := func(links []page.Link, want page.Link) bool {
		for _, l := range links {
			if l == want {
				return true
			}
		}
		return false
	}
	for _, want := range []page.Link{
		{Line: 10, Col: 14, Len: 4, File: "cart/cart.go", To: 1},                   // the package cart
		{Line: 10, Col: 46, Len: 7, URL: "https://pkg.go.dev/fmt#Sprintf"},         // a function of the standard library
		{Line: 10, Col: 71, Len: 5, File: "cart/cart.go", To: 13, Mark: "reached"}, // cart.Total
	} {
		for _, links := range [][]page.Link{d.Files["receipt/receipt.go"].BeforeLinks, d.Files["receipt/receipt.go"].AfterLinks} {
			if !has(links, want) {
				t.Errorf("receipt/receipt.go lacks the link %+v in %+v", want, links)
			}
		}
	}
	// cost changed in its own text, and Total holds its call, so the link carries the mark.
	if !has(d.Files["cart/cart.go"].AfterLinks, page.Link{Line: 16, Col: 13, Len: 4, File: "cart/cart.go", To: 21, Mark: "changed"}) {
		t.Errorf("it.cost() does not link to the method with its mark: %+v", d.Files["cart/cart.go"].AfterLinks)
	}
	if !d.Decls["cart.Total"].Reached {
		t.Errorf("cart.Total counts as changed in its own text")
	}
	// No drawing links the constant Footer, and no linked function reaches it.
	if want := []page.Uncovered{{Key: "receipt.Footer", Mark: "changed", File: "receipt/footer.go", Line: 3, Version: "after"}}; !reflect.DeepEqual(d.Uncovered, want) {
		t.Errorf("changes outside the drawings: %+v, want %+v", d.Uncovered, want)
	}
	if len(d.Errors) != 1 || !strings.HasPrefix(d.Errors[0], "working tree: example.com/shop/receipt: receipt/footer.go:4:") {
		t.Errorf("package errors: %q", d.Errors)
	}
}

// TestShopInChrome renders the shop's page in headless Chrome, opened at cart.Total, and
// checks the drawing's marks, the identifier links and the package errors in its DOM.
func TestShopInChrome(t *testing.T) {
	chrome := os.Getenv("CHROME")
	if chrome == "" {
		chrome = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	}
	if _, err := os.Stat(chrome); err != nil {
		if chrome, err = exec.LookPath("google-chrome"); err != nil {
			t.Skip("no Chrome; set CHROME to one")
		}
	}
	dfd := os.Getenv("DFD")
	if dfd == "" {
		dfd = "dfd"
	}
	if help, _ := exec.Command(dfd, "--help").CombinedOutput(); !strings.Contains(string(help), "-patch") {
		t.Skip("no dfd with --patch; set DFD to one")
	}
	opts := shopPage(t, dfd)
	if err := dfdreview.Build(opts); err != nil {
		t.Fatal(err)
	}
	dom := dumpDOM(t, chrome, "file://"+opts.Out+"#code/cart.Total")
	for name, pattern := range map[string]string{
		"the status of the deep link":                `cart\.Total in cart/cart\.go:12 \(After\)`,
		"the reach mark on cart.Total":               `href="#code/cart\.Total"><tspan[^>]*>cart\.Total</tspan><tspan class="reached">~</tspan>`,
		"the highlight behind cart.Total":            `<rect [^>]*class="hl reached"`,
		"the reach mark on receipt.Print":            `href="#code/receipt\.Print"><tspan[^>]*>receipt\.Print</tspan><tspan class="reached">~</tspan>`,
		"the changed mark on the tab, from the list": `Overview<span class="mark changed"[^>]*>\*</span>`,
		"the list of changes outside the drawings":   `<h2>Changes outside the drawings</h2>`,
		"the changed constant in the list":           `class="ident mark-changed"[^>]*>receipt\.Footer</a>`,
		"the marked link from it.cost() to cost":     `<a class="ident mark-changed"[^>]*data-file="cart/cart\.go" data-line="21"[^>]*>cost</a>`,
		"the link from o.Items to the field":         `<a class="ident" href="#" data-file="cart/cart\.go" data-line="4" data-version="after">Items</a>`,
		"the package errors":                         `<h2>Package errors</h2><ul class="errors"><li>working tree: example\.com/shop/receipt: receipt/footer\.go:4:`,
	} {
		if !regexp.MustCompile(pattern).MatchString(dom) {
			t.Errorf("the page lacks %s: %s", name, pattern)
		}
	}
}

// dumpDOM returns the DOM of the page at url after headless Chrome loads it. Chrome can stay
// up after it prints the DOM, so dumpDOM ends it once the DOM is complete.
func dumpDOM(t *testing.T, chrome, url string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, chrome, "--headless=new", "--disable-gpu", "--no-first-run",
		"--no-default-browser-check", "--user-data-dir="+t.TempDir(), "--dump-dom", url)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	var dom strings.Builder
	buf := make([]byte, 64<<10)
	for !strings.Contains(dom.String(), "</html>") {
		n, err := stdout.Read(buf)
		dom.Write(buf[:n])
		if err != nil {
			break
		}
	}
	if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		t.Errorf("end chrome: %v", err)
	}
	_ = cmd.Wait() // it reports the kill
	if !strings.Contains(dom.String(), "</html>") {
		t.Fatalf("chrome printed no complete DOM:\n%.500s", dom.String())
	}
	return dom.String()
}

// TestShopPlaceInChrome opens the shop's page at a place that its URL hash names, as the
// page writes it into the browser's history, and checks the place in headless Chrome's DOM.
func TestShopPlaceInChrome(t *testing.T) {
	chrome := os.Getenv("CHROME")
	if chrome == "" {
		chrome = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	}
	if _, err := os.Stat(chrome); err != nil {
		if chrome, err = exec.LookPath("google-chrome"); err != nil {
			t.Skip("no Chrome; set CHROME to one")
		}
	}
	dfd := filepath.Join(t.TempDir(), "dfd")
	if err := os.WriteFile(dfd, []byte(echo), 0o755); err != nil {
		t.Fatal(err)
	}
	opts := shopPage(t, dfd)
	if err := dfdreview.Build(opts); err != nil {
		t.Fatal(err)
	}
	dom := dumpDOM(t, chrome, "file://"+opts.Out+"#f=cart/cart.go&fv=before&l=21")
	for name, pattern := range map[string]string{
		"the status of the place": `cart/cart\.go:21 \(Before\)`,
		"the selected line":       `<div class="line sel" id="L21">`,
		"the Before view's links": `data-line="7" data-version="before">Item</a>`,
	} {
		if !regexp.MustCompile(pattern).MatchString(dom) {
			t.Errorf("the page lacks %s: %s", name, pattern)
		}
	}
}
