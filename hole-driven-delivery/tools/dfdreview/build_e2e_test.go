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
		"cart/pricer.go":    "package cart\n\n// Pricer prices a line of an order.\ntype Pricer interface{ cost() int }\n",
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
		{Line: 10, Col: 14, Len: 4, File: "cart/cart.go", To: 1},           // the package cart
		{Line: 10, Col: 46, Len: 7, URL: "https://pkg.go.dev/fmt#Sprintf"}, // a function of the standard library
		{Line: 10, Col: 71, Len: 5, File: "cart/cart.go", To: 13, Key: "cart.Total", Mark: "reached"},
	} {
		for _, links := range [][]page.Link{d.Files["receipt/receipt.go"].BeforeLinks, d.Files["receipt/receipt.go"].AfterLinks} {
			if !has(links, want) {
				t.Errorf("receipt/receipt.go lacks the link %+v in %+v", want, links)
			}
		}
	}
	// cost changed in its own text, and Total holds its call, so the link carries the mark.
	if !has(d.Files["cart/cart.go"].AfterLinks, page.Link{Line: 16, Col: 13, Len: 4, File: "cart/cart.go", To: 21, Key: "cart.Item.cost", Mark: "changed"}) {
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
	// The search box finds methods too, and each declaration and method has its kind.
	for key, want := range map[string]page.Place{
		"cart.Total":     {File: "cart/cart.go", Start: 12, End: 19, Kind: "func"},
		"cart.Item.cost": {File: "cart/cart.go", Start: 21, End: 21, Kind: "method"},
		"cart.Order":     {File: "cart/cart.go", Start: 3, End: 4, Kind: "type"},
		"receipt.Footer": {File: "receipt/footer.go", Start: 3, End: 4, Kind: "const"},
	} {
		if got := d.Decls[key].After; got == nil || *got != want {
			t.Errorf("%s: %+v, want %+v", key, got, want)
		}
	}
	// Item implements Pricer through its method cost, in both versions.
	pricer := []page.Implementation{{Key: "cart.Item", File: "cart/cart.go", Line: 7}}
	for version, impls := range map[string]map[string][]page.Implementation{"before": d.Implementations.Before, "after": d.Implementations.After} {
		if !reflect.DeepEqual(impls["cart.Pricer"], pricer) {
			t.Errorf("%s: the implementations of cart.Pricer: %+v, want %+v", version, impls["cart.Pricer"], pricer)
		}
	}
	// A test calls cart.Total, and so does receipt.Print, in both versions.
	total := []page.Use{{In: "cart.TestTotal", File: "cart/cart_test.go", Line: 6}, {In: "receipt.Print", File: "receipt/receipt.go", Line: 10}}
	for version, uses := range map[string]map[string][]page.Use{"before": d.Uses.Before, "after": d.Uses.After} {
		if !reflect.DeepEqual(uses["cart.Total"], total) {
			t.Errorf("%s: the uses of cart.Total: %+v, want %+v", version, uses["cart.Total"], total)
		}
	}
}

// TestShopInChrome renders the shop's page in headless Chrome, opened at cart.Total, and
// checks the drawing's marks, the identifier links and the package errors in its DOM.
func TestShopInChrome(t *testing.T) {
	chrome := chromePath(t)
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

// chromePath returns the Chrome that CHROME names, or the one of macOS, or google-chrome on the
// PATH. Without one, it skips the test.
func chromePath(t *testing.T) string {
	t.Helper()
	chrome := os.Getenv("CHROME")
	if chrome == "" {
		chrome = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	}
	if _, err := os.Stat(chrome); err != nil {
		if chrome, err = exec.LookPath("google-chrome"); err != nil {
			t.Skip("no Chrome; set CHROME to one")
		}
	}
	return chrome
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
	chrome := chromePath(t)
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

// actions drive a page in the tests. type puts text into the search box, key presses a key
// in it, and click clicks the first element that a selector matches. Each event can be
// cancelled, as a user's can.
const actions = `<script>
const type = (text) => {
  const q = document.getElementById("search");
  q.value = text;
  q.dispatchEvent(new Event("input", { bubbles: true }));
};
const key = (name) => document.getElementById("search").dispatchEvent(new KeyboardEvent("keydown", { key: name, bubbles: true, cancelable: true }));
const click = (selector) => document.querySelector(selector).dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
`

// driven writes a copy of the page at path that runs script after the page's own script, and
// returns the copy's path. The copy's body records the URL hash at the end, in data-hash.
func driven(t *testing.T, path, script string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// The page's data escapes every "<", so the first "</body>" ends the body.
	run := actions + script + "\ndocument.body.dataset.hash = location.hash;\n</script>\n</body>"
	out := filepath.Join(t.TempDir(), "index.html")
	if err := os.WriteFile(out, []byte(strings.Replace(string(b), "</body>", run, 1)), 0o644); err != nil {
		t.Fatal(err)
	}
	return out
}

// TestShopSearchInChrome drives the search box and the "Show references" link of the shop's
// page in headless Chrome, and checks the DOM after each series of actions.
func TestShopSearchInChrome(t *testing.T) {
	chrome := chromePath(t)
	dfd := filepath.Join(t.TempDir(), "dfd")
	if err := os.WriteFile(dfd, []byte(echo), 0o755); err != nil {
		t.Fatal(err)
	}
	opts := shopPage(t, dfd)
	if err := dfdreview.Build(opts); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name, hash, script string
		want               map[string]string
	}{
		{
			name:   "typing part of a name",
			script: `type("t");`,
			want: map[string]string{
				"the open list": `<ul id="search-results" role="listbox" aria-label="[^"]*">`,
				// Every key holds a "t", but only one name starts with it.
				"the name that starts with the text, first": `id="result-0" role="option" data-key="cart\.Total"[^>]*>` +
					`<span class="kind">func</span><span>cart\.Total</span><span class="mark reached"[^>]*>~</span><span class="where">cart/cart\.go:12</span>`,
				"the other keys, in order": `id="result-1" role="option" data-key="cart\.Item".*id="result-2" role="option" data-key="cart\.Item\.cost"`,
			},
		},
		{
			name:   "choosing a method and showing its references",
			script: `type("cost"); key("Enter"); click("#refs-link");`,
			want: map[string]string{
				"the status of the method": `cart\.Item\.cost in cart/cart\.go:21 \(After\)`,
				"the closed search list":   `<ul id="search-results" role="listbox" aria-label="[^"]*" hidden="">`,
				"the list's heading":       `<p>References to cart\.Item\.cost in the After version:</p>`,
				"the call in cart.Total": `<li><a href="#" data-key="cart\.Total" data-file="cart/cart\.go" data-line="16" data-version="after">cart\.Total</a>` +
					`<span class="where">cart/cart\.go:<a [^>]*data-line="16"[^>]*>16</a></span></li>`,
				"the link to close the list": `aria-expanded="true" aria-controls="refs">Hide references</button>`,
			},
		},
		{
			name:   "following a reference",
			hash:   "#code/cart.Total",
			script: `click("#refs-link"); click("#refs a[data-key='receipt.Print']");`,
			want: map[string]string{
				"the status of the calling function": `receipt\.Print in receipt/receipt\.go:10 \(After\)`,
				"its doc comment, selected":          `<div class="line sel" id="L9">`,
				"the line of the call, in focus":     `<div class="line sel focus" id="L10">`,
				"the closed list":                    `<div id="refs" hidden="">`,
				"the link, ready again":              `aria-expanded="false" aria-controls="refs">Show references</button>`,
				"the place in the history":           `data-hash="#f=receipt/receipt\.go&amp;fv=after&amp;k=receipt\.Print&amp;l=10"`,
			},
		},
		{
			name:   "following an identifier to a declaration",
			hash:   "#f=receipt/receipt.go&fv=after",
			script: `click("#code-body a[data-key='cart.Total']");`,
			want: map[string]string{
				"the status at the definition's line": `cart\.Total in cart/cart\.go:13 \(After\)`,
				"its doc comment, selected":           `<div class="line sel" id="L12">`,
				"the definition's line, in focus":     `<div class="line sel focus" id="L13">`,
				"its last line, selected":             `<div class="line sel" id="L19">`,
				"the place in the history":            `data-hash="#f=cart/cart\.go&amp;fv=after&amp;k=cart\.Total&amp;l=13"`,
				"no implementations for a function":   `aria-controls="impls" hidden="">Show implementations</button>`,
			},
		},
		{
			name:   "showing the implementations of an interface",
			hash:   "#code/cart.Pricer",
			script: `click("#impls-link");`,
			want: map[string]string{
				"the status of the interface": `cart\.Pricer in cart/pricer\.go:3 \(After\)`,
				"the link to close the list":  `aria-expanded="true" aria-controls="impls">Hide implementations</button>`,
				"the list's heading":          `<p>Implementations of cart\.Pricer in the After version:</p>`,
				"the implementing type": `<li><a href="#" data-key="cart\.Item" data-file="cart/cart\.go" data-line="7" data-version="after">cart\.Item</a>` +
					`<span class="where">cart/cart\.go:7</span></li>`,
			},
		},
		{
			name:   "selecting a declaration by its name",
			hash:   "#f=cart/cart.go&fv=after",
			script: `click("#code-body a[data-key='cart.Total'][data-end]"); click("#refs-link");`,
			want: map[string]string{
				"the status of the declaration": `cart\.Total in cart/cart\.go:12 \(After\)`,
				"its doc comment, selected":     `<div class="line sel" id="L12">`,
				"its last line, selected":       `<div class="line sel" id="L19">`,
				"the line after it, unselected": `<div class="line" id="L20">`,
				"a caller in the list":          `<a href="#" data-key="receipt\.Print" data-file="receipt/receipt\.go" data-line="10" data-version="after">receipt\.Print</a>`,
				"the place in the history":      `data-hash="#f=cart/cart\.go&amp;fv=after&amp;k=cart\.Total&amp;l=12&amp;e=19"`,
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			dom := dumpDOM(t, chrome, "file://"+driven(t, opts.Out, c.script)+c.hash)
			for name, pattern := range c.want {
				if !regexp.MustCompile(pattern).MatchString(dom) {
					t.Errorf("the page lacks %s: %s", name, pattern)
				}
			}
		})
	}
}
