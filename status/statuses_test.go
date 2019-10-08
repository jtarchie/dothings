package status_test

import (
	. "github.com/jtarchie/dothings/status"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Statuses", func() {
	When("no status is defined", func() {
		It("allows unstarted as the initial Stater", func() {
			statuses := NewStatuses()
			Expect(statuses.Get(task("A"))).To(Equal([]Type{}))

			err := statuses.Add(task("A"), Unstarted)
			Expect(err).ToNot(HaveOccurred())
			Expect(statuses.Get(task("A"))).To(Equal([]Type{Unstarted}))
		})

		It("does not allow any other Stater as the initial Stater", func() {
			for _, status := range []Type{Running, Success, Failed} {
				statuses := NewStatuses()
				err := statuses.Add(task("A"), status)
				Expect(err).To(HaveOccurred())
				Expect(statuses.Get(task("A"))).To(Equal([]Type{}))
			}
		})
	})

	When("transitioning from unstarted", func() {
		It("successfully transitions to running", func() {
			statuses := NewStatuses()
			err := statuses.Add(task("A"), Unstarted)
			Expect(err).ToNot(HaveOccurred())
			Expect(statuses.Get(task("A"))).To(Equal([]Type{Unstarted}))

			err = statuses.Add(task("A"), Running)
			Expect(err).ToNot(HaveOccurred())
			Expect(statuses.Get(task("A"))).To(Equal([]Type{Running}))
		})

		It("fails transitions to success, failed, errored", func() {
			statuses := NewStatuses()
			err := statuses.Add(task("A"), Unstarted)
			Expect(err).ToNot(HaveOccurred())
			Expect(statuses.Get(task("A"))).To(Equal([]Type{Unstarted}))

			for _, s := range []Type{Success, Failed, Errored} {
				err = statuses.Add(task("A"), s)
				Expect(err).To(HaveOccurred())
				Expect(statuses.Get(task("A"))).To(Equal([]Type{Unstarted}))
			}
		})
	})

	When("transitioning from running", func() {
		It("successfully transitions to success and failed", func() {
			statuses := NewStatuses()
			err := statuses.Add(task("A"), Unstarted)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Running)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Success)
			Expect(err).ToNot(HaveOccurred())
			Expect(statuses.Get(task("A"))).To(Equal([]Type{Success}))

			statuses = NewStatuses()
			err = statuses.Add(task("A"), Unstarted)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Running)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Failed)
			Expect(err).ToNot(HaveOccurred())
			Expect(statuses.Get(task("A"))).To(Equal([]Type{Failed}))
		})

		It("fails transitions to running and unstarted", func() {
			statuses := NewStatuses()
			err := statuses.Add(task("A"), Unstarted)
			Expect(err).ToNot(HaveOccurred())
			Expect(statuses.Get(task("A"))).To(Equal([]Type{Unstarted}))

			err = statuses.Add(task("A"), Running)
			Expect(err).ToNot(HaveOccurred())
			Expect(statuses.Get(task("A"))).To(Equal([]Type{Running}))
		})
	})

	When("transitioning from failure", func() {
		It("creates a new Stater when transitioning to unstarted", func() {
			statuses := NewStatuses()
			err := statuses.Add(task("A"), Unstarted)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Running)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Failed)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Unstarted)
			Expect(err).ToNot(HaveOccurred())
			Expect(statuses.Get(task("A"))).To(Equal([]Type{Failed, Unstarted}))
		})

		It("fails transitioning to anything else", func() {
			statuses := NewStatuses()
			err := statuses.Add(task("A"), Unstarted)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Running)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Failed)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Running)
			Expect(err).To(HaveOccurred())
			err = statuses.Add(task("A"), Success)
			Expect(err).To(HaveOccurred())
			err = statuses.Add(task("A"), Failed)
			Expect(err).To(HaveOccurred())

			Expect(statuses.Get(task("A"))).To(Equal([]Type{Failed}))
		})
	})

	When("transitioning from success", func() {
		It("creates a new Stater when transitioning to unstarted", func() {
			statuses := NewStatuses()
			err := statuses.Add(task("A"), Unstarted)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Running)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Success)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Unstarted)
			Expect(err).ToNot(HaveOccurred())
			Expect(statuses.Get(task("A"))).To(Equal([]Type{Success, Unstarted}))
		})

		It("fails transitioning to anything else", func() {
			statuses := NewStatuses()
			err := statuses.Add(task("A"), Unstarted)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Running)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Success)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Running)
			Expect(err).To(HaveOccurred())
			err = statuses.Add(task("A"), Success)
			Expect(err).To(HaveOccurred())
			err = statuses.Add(task("A"), Failed)
			Expect(err).To(HaveOccurred())

			Expect(statuses.Get(task("A"))).To(Equal([]Type{Success}))
		})
	})

	When("transitioning from errored", func() {
		It("creates a new Stater when transitioning to unstarted", func() {
			statuses := NewStatuses()
			err := statuses.Add(task("A"), Unstarted)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Running)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Errored)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Unstarted)
			Expect(err).ToNot(HaveOccurred())
			Expect(statuses.Get(task("A"))).To(Equal([]Type{Errored, Unstarted}))
		})

		It("fails transitioning to anything else", func() {
			statuses := NewStatuses()
			err := statuses.Add(task("A"), Unstarted)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Running)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Errored)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Running)
			Expect(err).To(HaveOccurred())
			err = statuses.Add(task("A"), Success)
			Expect(err).To(HaveOccurred())
			err = statuses.Add(task("A"), Failed)
			Expect(err).To(HaveOccurred())

			Expect(statuses.Get(task("A"))).To(Equal([]Type{Errored}))
		})
	})

	It("handles multiple transitions", func() {
		statuses := NewStatuses()
		for i := 0; i < 3; i++ {
			err := statuses.Add(task("A"), Unstarted)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Running)
			Expect(err).ToNot(HaveOccurred())
			err = statuses.Add(task("A"), Success)
			Expect(err).ToNot(HaveOccurred())
		}
		Expect(statuses.Get(task("A"))).To(Equal([]Type{Success, Success, Success}))
	})
})
