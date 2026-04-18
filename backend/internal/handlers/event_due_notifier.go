package handlers

import (
	"log"
	"time"
)

// StartEventDueNotifier launches a background goroutine that fires every minute,
// finds group events whose time has arrived, notifies all group members in
// real-time via WebSocket, and marks each event so it is not notified again.
func (h *HandlerStruct) StartEventDueNotifier() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			h.notifyDueEvents()
		}
	}()
}

func (h *HandlerStruct) notifyDueEvents() {
	if h.GroupRepo == nil {
		return
	}

	events, err := h.GroupRepo.FetchDueUnnotifiedEvents()
	if err != nil {
		log.Printf("event due notifier: fetch failed: %v", err)
		return
	}

	for _, event := range events {
		memberIDs, err := h.GroupRepo.MemberIDs(event.GroupID)
		if err != nil {
			log.Printf("event due notifier: member lookup failed group=%s: %v", event.GroupID, err)
			continue
		}

		for _, memberID := range memberIDs {
			h.createNotification(memberID, event.CreatorID, "group_event_due", event.ID)
		}

		if err := h.GroupRepo.MarkEventDueNotified(event.ID); err != nil {
			log.Printf("event due notifier: mark failed event=%s: %v", event.ID, err)
		}
	}
}
