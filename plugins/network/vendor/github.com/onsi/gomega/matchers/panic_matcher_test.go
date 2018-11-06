package matchers_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/matchers"
)

var _ = Describe("Panic", func() {
	Context("when passed something that's not a function that takes zero arguments and returns nothing", func() {
		It("should error", func() {
			success, err := (&PanicMatcher{}).Match("foo")
			Expect(success).Should(BeFalse())
			Expect(err).Should(HaveOccurred())

			success, err = (&PanicMatcher{}).Match(nil)
			Expect(success).Should(BeFalse())
			Expect(err).Should(HaveOccurred())

			success, err = (&PanicMatcher{}).Match(func(foo string) {})
			Expect(success).Should(BeFalse())
			Expect(err).Should(HaveOccurred())

			success, err = (&PanicMatcher{}).Match(func() string { return "bar" })
			Expect(success).Should(BeFalse())
			Expect(err).Should(HaveOccurred())
		})
	})

	Context("when passed a function of the correct type", func() {
		It("should call the function and pass if the function panics", func() {
			Expect(func() { panic("ack!") }).Should(Panic())
			Expect(func() {}).ShouldNot(Panic())
		})
	})

	Context("when assertion fails", func() {
		It("should print the object passed to Panic", func() {
			failuresMessages := InterceptGomegaFailures(func() {
				Expect(func() { panic("ack!") }).ShouldNot(Panic())
			})
			Expect(failuresMessages).Should(ConsistOf(MatchRegexp("not to panic, but panicked with\\s*<string>: ack!")))
		})
	})
})
