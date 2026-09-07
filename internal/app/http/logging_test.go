package http

import (
	"bytes"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ensoria/ensoria-template/internal/plamo/restkit"
)

// A panic reaches this function through middleware that knows nothing about
// declarations or authorization, so whatever the panic value can say about
// itself is all the record can say. Expanding it here is what turns a stack
// trace of generic controller frames into a record naming the endpoint.
var _ = Describe("logPanicDetails", func() {
	drift := &restkit.DeclarationDrift{
		Method: http.MethodPost, Path: "/things", Status: http.StatusAccepted,
	}

	It("expands the fields of a contract violation the panic carried", func() {
		lines := captureLogLines(func() { logPanicDetails(request(), drift, []byte("stack")) })

		Expect(lines).To(HaveLen(1))
		Expect(decodeRecord(lines[0])).To(HaveKeyWithValue(panicViolationGroup, map[string]any{
			"method": http.MethodPost,
			"path":   "/things",
			"status": float64(http.StatusAccepted),
		}))
	})

	// An alert on panics matches on the record's type, so the record must carry
	// exactly one. The count is taken on the raw line: the violation must
	// contribute no type at any level, and decoding first would collapse a
	// duplicate key and hide it.
	It("writes one type, and it is the one the caller chose", func() {
		lines := captureLogLines(func() { logPanicDetails(request(), drift, []byte("stack")) })

		Expect(bytes.Count(lines[0], []byte(`"type":`))).To(Equal(1))
		Expect(decodeRecord(lines[0])).To(HaveKeyWithValue("type", "panic_log"))
	})

	// The violation names an endpoint too, and it need not be the one the
	// request was made against — a group keeps the two apart instead of leaving
	// one "method" to overwrite the other.
	It("leaves the record's own fields describing the request", func() {
		lines := captureLogLines(func() { logPanicDetails(request(), drift, []byte("stack")) })

		record := decodeRecord(lines[0])
		Expect(record).To(HaveKeyWithValue("method", http.MethodGet))
		Expect(record[panicViolationGroup]).To(HaveKeyWithValue("method", http.MethodPost))
	})

	It("keeps the record it always wrote for a panic that is not a violation", func() {
		lines := captureLogLines(func() { logPanicDetails(request(), "boom", []byte("stack")) })

		record := decodeRecord(lines[0])
		Expect(record).To(HaveKeyWithValue("type", "panic_log"))
		Expect(record).To(HaveKeyWithValue("panic_value", "boom"))
		Expect(record).To(HaveKeyWithValue("stack_trace", "stack"))
		Expect(record).NotTo(HaveKey(panicViolationGroup))
	})
})
