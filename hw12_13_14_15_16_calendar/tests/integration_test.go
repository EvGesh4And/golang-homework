package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Event API", func() {
	eventsGroup := []storage.Event{
		{
			ID:          uuid.New(),
			Title:       "test event",
			Description: "test desc",
			UserID:      uuid.New(),
			Start:       time.Now().Add(time.Hour),
			End:         time.Now().Add(2 * time.Hour),
			TimeBefore:  10 * time.Minute,
		},
		{
			ID:          uuid.New(),
			Title:       "test event 2",
			Description: "test desc 2",
			UserID:      uuid.New(),
			Start:       time.Now().Add(3 * time.Hour),
			End:         time.Now().Add(4 * time.Hour),
			TimeBefore:  15 * time.Minute,
		},
	}
	Context("create event", func() {
		It("creates an event successfully", func() {
			for _, event := range eventsGroup {
				eventDTO := storage.ToDTO(event)

				body, err := json.Marshal(eventDTO)
				Expect(err).To(BeNil())

				req, _ := http.NewRequest(http.MethodPost, "http://localhost:8888/event", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")

				resp, err := http.DefaultClient.Do(req)

				Expect(err).To(BeNil())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			}
		})

		It("creates an event invalid event: id", func() {
			invalidEvent := storage.Event{
				Title:       "test event",
				Description: "test desc",
				UserID:      uuid.New(),
				Start:       time.Now().Add(time.Hour),
				End:         time.Now().Add(2 * time.Hour),
				TimeBefore:  10 * time.Minute,
			}

			eventDTO := storage.ToDTO(invalidEvent)

			body, err := json.Marshal(eventDTO)
			Expect(err).To(BeNil())

			req, _ := http.NewRequest(http.MethodPost, "http://localhost:8888/event", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)

			Expect(err).To(BeNil())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("creates an event invalid event: start time", func() {
			invalidEvent := storage.Event{
				ID:          uuid.New(),
				Title:       "",
				Description: "test desc",
				UserID:      uuid.New(),
				Start:       time.Now().Add(-time.Hour),
				End:         time.Now().Add(2 * time.Hour),
				TimeBefore:  10 * time.Minute,
			}

			eventDTO := storage.ToDTO(invalidEvent)

			body, err := json.Marshal(eventDTO)
			Expect(err).To(BeNil())

			req, _ := http.NewRequest(http.MethodPost, "http://localhost:8888/event", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	Context("update event", func() {
		It("update an event successfully", func() {
			event := eventsGroup[0]
			event.Start = time.Now().Add(10 * time.Second)
			event.Title = "updated title"
			event.Description = "updated description"
			event.TimeBefore = 5 * time.Second

			eventDTO := storage.ToDTO(event)

			body, err := json.Marshal(eventDTO)
			Expect(err).To(BeNil())

			req, _ := http.NewRequest(http.MethodPut,
				"http://localhost:8888/event?id="+eventsGroup[0].ID.String(), bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})
	})

	Context("get events", func() {
		It("get an event day successfully", func() {
			req, _ := http.NewRequest(http.MethodGet,
				"http://localhost:8888/event/day?start="+time.Now().Add(-2*time.Hour).Format(time.RFC3339), nil)
			resp, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer resp.Body.Close()

			var eventsDTO []storage.EventDTO
			err = json.NewDecoder(resp.Body).Decode(&eventsDTO)
			Expect(err).To(BeNil())

			var events []storage.Event

			for _, eventDTO := range eventsDTO {
				events = append(events, storage.FromDTO(eventDTO))
			}

			Expect(events).To(HaveLen(2))
			Expect(events[0].Title).To(Equal("test event 2"))
			Expect(events[1].Title).To(Equal("updated title"))

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
		It("get an event month successfully", func() {
			req, _ := http.NewRequest(http.MethodGet,
				"http://localhost:8888/event/month?start="+time.Now().Add(-2*time.Hour).Format(time.RFC3339), nil)
			resp, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer resp.Body.Close()

			var eventsDTO []storage.EventDTO
			err = json.NewDecoder(resp.Body).Decode(&eventsDTO)
			Expect(err).To(BeNil())

			var events []storage.Event

			for _, eventDTO := range eventsDTO {
				events = append(events, storage.FromDTO(eventDTO))
			}

			Expect(events).To(HaveLen(2))
			Expect(events[0].Title).To(Equal("test event 2"))
			Expect(events[1].Title).To(Equal("updated title"))

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
		It("get an event week successfully", func() {
			req, _ := http.NewRequest(http.MethodGet,
				"http://localhost:8888/event/week?start="+time.Now().Add(-2*time.Hour).Format(time.RFC3339), nil)
			resp, err := http.DefaultClient.Do(req)
			Expect(err).To(BeNil())
			defer resp.Body.Close()

			var eventsDTO []storage.EventDTO
			err = json.NewDecoder(resp.Body).Decode(&eventsDTO)
			Expect(err).To(BeNil())

			var events []storage.Event

			for _, eventDTO := range eventsDTO {
				events = append(events, storage.FromDTO(eventDTO))
			}

			Expect(events).To(HaveLen(2))
			Expect(events[0].Title).To(Equal("test event 2"))
			Expect(events[1].Title).To(Equal("updated title"))

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})
	It("prints message on event", func(ctx SpecContext) {
		Eventually(func() string {
			out, _ := exec.CommandContext(ctx, "docker", "logs", "sender", "--since", "1s").CombinedOutput()
			return string(out)
		}).WithTimeout(10 * time.Second).Should(ContainSubstring("updated title"))
	})
})
