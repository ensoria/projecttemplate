package describe

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The Environments section used to be the one part of a generated document that
// did not follow --env: it was written as "local" pointing at localhost whatever
// environment was being described, so a document generated for production
// announced itself as local and sent its readers to their own machine.
//
// These specs cover the translation on its own. What the section resolves to
// end to end depends on the configuration, which this package deliberately does
// not initialize (see http_test.go).
var _ = Describe("websocketAddress", func() {
	It("takes the host out of the address the deployment answers on", func() {
		protocol, host := websocketAddress("http://localhost:8080")

		Expect(protocol).To(Equal("ws"))
		Expect(host).To(Equal("localhost:8080"))
	})

	// Publishing ws for an application served over https tells a browser to
	// open a connection it then refuses as mixed content.
	It("upgrades the scheme when the application is served over TLS", func() {
		protocol, host := websocketAddress("https://api.example.com")

		Expect(protocol).To(Equal("wss"))
		Expect(host).To(Equal("api.example.com"))
	})

	// AsyncAPI puts the path on the channel, and this application's channels
	// declare their own — carrying it here would double it.
	It("drops a path prefix", func() {
		_, host := websocketAddress("https://example.com/api")

		Expect(host).To(Equal("example.com"))
	})

	It("keeps a port", func() {
		_, host := websocketAddress("https://api.example.com:8443")

		Expect(host).To(Equal("api.example.com:8443"))
	})

	// The configuration refuses a URL with no host, so this is unreachable in
	// practice; answering with something is still better than answering with
	// an empty host that publishes a server nobody can connect to.
	It("falls back to the value it was given when it is not a URL", func() {
		protocol, host := websocketAddress("localhost:8080")

		Expect(protocol).To(Equal("ws"))
		Expect(host).To(Equal("localhost:8080"))
	})
})
