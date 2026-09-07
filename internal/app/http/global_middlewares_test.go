package http

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ensoria/config/pkg/appconfig"
	"github.com/ensoria/ensoria-template/internal/plamo/apidoc"
	"github.com/ensoria/ensoria-template/internal/plamo/authkit"
	"github.com/ensoria/rest/pkg/rest"
)

// The chain and its names used to be two lists, and they drifted: the names
// stayed at four entries through the two changes that added CSRF and Auth, so
// every generated document described a pipeline the application had stopped
// running. They are one list now, which makes that particular drift
// unrepresentable — what is left to check is that each entry is complete, since
// an entry with no name or no builder still compiles.
// chainDeps builds what the chain is assembled from, around the verifier under
// test. Everything else is the same for every spec here.
func chainDeps(verifier authkit.Verifier) *globalMiddlewareDeps {
	return &globalMiddlewareDeps{
		cors:          &appconfig.CORS{AllowOriginVal: "*"},
		crossOrigin:   http.NewCrossOriginProtection(),
		verifier:      verifier,
		panicResponse: &rest.Response{Code: http.StatusInternalServerError},
	}
}

var _ = Describe("globalMiddlewareChain", func() {
	It("gives every middleware it installs a name", func() {
		for i, m := range globalMiddlewareChain {
			Expect(m.Name).ToNot(BeEmpty(), "entry %d has no name", i)
		}
	})

	It("builds a middleware for every name it publishes", func() {
		chain := globalMiddlewares(chainDeps(anonymousVerifier{}))

		Expect(chain).To(HaveLen(len(GlobalMiddlewareNames())))
		for i, middleware := range chain {
			Expect(middleware).ToNot(BeNil(), "entry %d built nothing", i)
		}
	})

	It("uses no name twice", func() {
		seen := map[string]bool{}
		for _, name := range GlobalMiddlewareNames() {
			Expect(seen[name]).To(BeFalse(), "duplicate middleware name %q", name)
			seen[name] = true
		}
	})

	// This one is read rather than merely listed: the cross-origin paragraph in
	// the generated browser-security section appears only when it is present.
	It("names the cross-origin guard the documentation asks about", func() {
		Expect(GlobalMiddlewareNames()).To(ContainElement(apidoc.MiddlewareCSRF))
	})

	// The order is the request's path through the chain, and the doc comment
	// argues for this one specifically: a preflight is answered before
	// authentication, and a forged request is refused before the session store
	// is asked about its cookie.
	It("keeps authentication behind the cross-origin guard", func() {
		names := GlobalMiddlewareNames()
		Expect(indexOf(names, apidoc.MiddlewareCORS)).
			To(BeNumerically("<", indexOf(names, apidoc.MiddlewareCSRF)))
		Expect(indexOf(names, apidoc.MiddlewareCSRF)).
			To(BeNumerically("<", indexOf(names, apidoc.MiddlewareAuth)))
	})

	// The names outlive the call, so handing out the declaration itself would
	// let a renderer sort or truncate what the next caller reads.
	It("hands out a copy rather than the declaration", func() {
		first := GlobalMiddlewareNames()
		first[0] = "clobbered"

		Expect(GlobalMiddlewareNames()[0]).ToNot(Equal("clobbered"))
	})
})

// indexOf reports where name sits in the chain, or -1.
func indexOf(names []string, name string) int {
	for i, n := range names {
		if n == name {
			return i
		}
	}
	return -1
}

var _ = Describe("globalMiddlewares", func() {
	// Accepting the verifier and then forgetting to install the middleware is a
	// silent hole: the application compiles and serves every request unchecked.
	It("refuses a request whose credential cannot be trusted", func() {
		reached := false

		res := chain(globalMiddlewares(chainDeps(rejectingVerifier{})),
			func(*rest.Request) *rest.Response {
				reached = true
				return &rest.Response{Code: http.StatusOK}
			})(request())

		Expect(res.Code).To(Equal(http.StatusUnauthorized))
		Expect(reached).To(BeFalse(), "the request reached the handler without being authenticated")
	})

	It("keeps serving a request that presents no credential", func() {
		res := chain(globalMiddlewares(chainDeps(anonymousVerifier{})),
			func(*rest.Request) *rest.Response { return &rest.Response{Code: http.StatusOK} })(request())

		Expect(res.Code).To(Equal(http.StatusOK),
			"a public endpoint must still be reachable without a credential")
	})
})
