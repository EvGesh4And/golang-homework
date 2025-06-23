package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("POST /event", func() {
	It("creates an event successfully", func() {
		event := storage.Event{
			ID:          uuid.New(),
			Title:       "test event",
			Description: "test desc",
			UserID:      uuid.New(),
			Start:       time.Now().Add(time.Hour),
			End:         time.Now().Add(2 * time.Hour),
			TimeBefore:  10 * time.Minute,
		}

		body, err := json.Marshal(event)
		Expect(err).To(BeNil())

		resp, err := http.Post("http://localhost:8888/event", "application/json", bytes.NewReader(body))
		Expect(err).To(BeNil())
		defer resp.Body.Close()

		Expect(resp.StatusCode).To(Equal(http.StatusCreated))
	})
})
