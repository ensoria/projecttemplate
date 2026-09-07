package http

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ensoria/config/pkg/appconfig"
	"github.com/ensoria/ensoria-template/internal/plamo/apidoc"
)

// The chain and the names beside it are two statements of one fact, and only a
// spec can hold them together — nothing in the type system relates a
// []rest.Middleware to a []string. The list already went stale once, staying at
// four entries through the two releases that added CSRF and Auth, so the
// generated documentation described a pipeline the application had stopped
// running.
var _ = Describe("GlobalMiddlewareNames", func() {
	It("names exactly as many middlewares as the chain installs", func() {
		chain := globalMiddlewares(&appconfig.CORS{}, nil, nil, nil)

		Expect(GlobalMiddlewareNames).To(HaveLen(len(chain)))
	})

	// This one is read rather than merely listed: the cross-origin paragraph in
	// the generated browser-security section appears only when it is present.
	It("names the cross-origin guard the documentation asks about", func() {
		Expect(GlobalMiddlewareNames).To(ContainElement(apidoc.MiddlewareCSRF))
	})

	It("uses no name twice", func() {
		seen := map[string]bool{}
		for _, name := range GlobalMiddlewareNames {
			Expect(seen[name]).To(BeFalse(), "duplicate middleware name %q", name)
			seen[name] = true
		}
	})
})
