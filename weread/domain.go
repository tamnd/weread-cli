package weread

import (
	"context"

	"github.com/tamnd/any-cli/kit"
)

// domain.go exposes weread as a kit Domain: a driver that a multi-domain
// host (ant) enables with a single blank import,
//
//	import _ "github.com/tamnd/weread-cli/weread"
//
// exactly as a database/sql program enables a driver with `import _
// "github.com/lib/pq"`. The init below registers it; the host then dereferences
// weread:// URIs by routing to the operations Register installs. The same
// Domain also builds the standalone weread binary, so the binary and a host
// share one source of truth.
func init() { kit.Register(Domain{}) }

// Domain is the weread driver. It carries no state; the per-run client is
// built by the factory Register hands kit.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	cfg := DefaultConfig()
	return kit.DomainInfo{
		Scheme: "weread",
		Hosts:  []string{"weread.qq.com"},
		Identity: kit.Identity{
			Binary: "weread",
			Short:  "A command line for WeRead.",
			Long: `A command line for WeRead (微信读书).

Search books on WeRead and get structured results. No API key required.`,
			Site: cfg.BaseURL,
			Repo: "https://github.com/tamnd/weread-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	kit.Handle(app, kit.OpMeta{Name: "search", Group: "books", List: true,
		URIType: "book", Summary: "Search books by keyword",
		Args: []kit.Arg{{Name: "keyword", Help: "search keyword"}}}, searchCmd)
}

// newClient builds the client from the host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	dc := DefaultConfig()
	if cfg.UserAgent != "" {
		dc.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		dc.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		dc.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		dc.Timeout = cfg.Timeout
	}
	return NewClient(dc), nil
}

type searchIn struct {
	Keyword string  `kit:"arg" help:"search keyword"`
	Limit   int     `kit:"flag,inherit" help:"max results"`
	Client  *Client `kit:"inject"`
}

func searchCmd(ctx context.Context, in searchIn, emit func(*Book) error) error {
	books, err := in.Client.Search(ctx, in.Keyword, in.Limit)
	if err != nil {
		return err
	}
	for i := range books {
		if err := emit(&books[i]); err != nil {
			return err
		}
	}
	return nil
}
