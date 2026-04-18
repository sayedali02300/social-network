package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"socialnetwork/backend/internal/repository"
)

type groupPayload struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type invitePayload struct {
	ReceiverID string `json:"receiverId"`
	Status     string `json:"status"`
}

type joinRequestPayload struct {
	Status string `json:"status"`
}

type eventPayload struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	EventTime   string `json:"eventTime"`
}

type eventResponsePayload struct {
	Response string `json:"response"`
}

func (h *HandlerStruct) GroupsCollectionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListGroupsHandler(w, r)
	case http.MethodPost:
		h.CreateGroupHandler(w, r)
	default:
		sendJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *HandlerStruct) GroupHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetGroupHandler(w, r)
	case http.MethodPatch:
		h.UpdateGroupHandler(w, r)
	case http.MethodDelete:
		h.DeleteGroupHandler(w, r)
	default:
		sendJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *HandlerStruct) GroupInvitesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListGroupInvitesHandler(w, r)
	case http.MethodPost:
		h.CreateGroupInviteHandler(w, r)
	default:
		sendJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *HandlerStruct) GroupInviteResponseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		sendJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.RespondGroupInviteHandler(w, r)
}

func (h *HandlerStruct) GroupRequestsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListGroupRequestsHandler(w, r)
	case http.MethodPost:
		h.CreateGroupRequestHandler(w, r)
	default:
		sendJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *HandlerStruct) GroupRequestResponseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		sendJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.RespondGroupRequestHandler(w, r)
}

func (h *HandlerStruct) GroupMembersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.ListGroupMembersHandler(w, r)
}

func (h *HandlerStruct) GroupMemberHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.RemoveGroupMemberHandler(w, r)
}

func (h *HandlerStruct) GroupEventsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListGroupEventsHandler(w, r)
	case http.MethodPost:
		h.CreateGroupEventHandler(w, r)
	default:
		sendJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *HandlerStruct) GroupPostsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetGroupPostsHandler(w, r)
	case http.MethodPost:
		h.CreateGroupPostHandler(w, r)
	default:
		sendJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *HandlerStruct) EventHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetEventHandler(w, r)
	case http.MethodPatch:
		h.UpdateEventHandler(w, r)
	case http.MethodDelete:
		h.DeleteEventHandler(w, r)
	default:
		sendJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *HandlerStruct) EventResponseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.RespondEventHandler(w, r)
}

func (h *HandlerStruct) ListGroupsHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	groups, err := h.GroupRepo.ListGroups(actorID)
	if err != nil {
		sendJSONError(w, "failed to load groups", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, groups)
}

func (h *HandlerStruct) CreateGroupHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var payload groupPayload
	if err := decodeJSON(r, &payload); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !validTitle(payload.Title) {
		sendJSONError(w, "title must be between 3 and 80 characters", http.StatusBadRequest)
		return
	}
	if !validDescription(payload.Description) {
	sendJSONError(w, "description must be at most 300 characters", http.StatusBadRequest)
	return
	}
	group, err := h.GroupRepo.CreateGroup(actorID, strings.TrimSpace(payload.Title), strings.TrimSpace(payload.Description))
	if err != nil {
		sendJSONError(w, "failed to create group", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, group)
}

func (h *HandlerStruct) GetGroupHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	group, err := h.GroupRepo.GetGroup(r.PathValue("groupId"), actorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sendJSONError(w, "group not found", http.StatusNotFound)
			return
		}
		sendJSONError(w, "failed to load group", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, group)
}

func (h *HandlerStruct) UpdateGroupHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var payload groupPayload
	if err := decodeJSON(r, &payload); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !validTitle(payload.Title) {
		sendJSONError(w, "title must be between 3 and 80 characters", http.StatusBadRequest)
		return
	}
	if !validDescription(payload.Description) {
	sendJSONError(w, "description must be at most 300 characters", http.StatusBadRequest)
	return
	}
	group, err := h.GroupRepo.UpdateGroup(r.PathValue("groupId"), actorID, strings.TrimSpace(payload.Title), strings.TrimSpace(payload.Description))
	if err != nil {
		handleRepositoryError(w, err, "failed to update group")
		return
	}

	h.emitGroupSummaryRealtime(group.ID)

	writeJSON(w, http.StatusOK, group)
}

func (h *HandlerStruct) DeleteGroupHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	if err := h.GroupRepo.DeleteGroup(r.PathValue("groupId"), actorID); err != nil {
		handleRepositoryError(w, err, "failed to delete group")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "group deleted"})
}

func (h *HandlerStruct) CreateGroupInviteHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var payload invitePayload
	if err := decodeJSON(r, &payload); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	invite, err := h.GroupRepo.CreateInvite(r.PathValue("groupId"), actorID, strings.TrimSpace(payload.ReceiverID))
	if err != nil {
		handleRepositoryError(w, err, "failed to create invite")
		return
	}
	h.createNotification(invite.ReceiverID, actorID, "group_invitation_received", invite.ID)
	h.emitGroupStateRefreshRealtime(invite.GroupID)

	writeJSON(w, http.StatusCreated, invite)
}

func (h *HandlerStruct) ListGroupInvitesHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	invites, err := h.GroupRepo.ListInvites(r.PathValue("groupId"), actorID)
	if err != nil {
		handleRepositoryError(w, err, "failed to load invites")
		return
	}

	writeJSON(w, http.StatusOK, invites)
}

func (h *HandlerStruct) RespondGroupInviteHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var payload invitePayload
	if err := decodeJSON(r, &payload); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	invite, err := h.GroupRepo.RespondInvite(r.PathValue("inviteId"), actorID, strings.TrimSpace(payload.Status))
	if err != nil {
		handleRepositoryError(w, err, "failed to update invite")
		return
	}

	// Notify the sender of the invite about the response.
	notifType := "group_invitation_accepted"
	if invite.Status == "declined" {
		notifType = "group_invitation_declined"
	}
	h.createNotification(invite.SenderID, actorID, notifType, invite.GroupID)

	h.emitGroupSummaryRealtime(invite.GroupID)
	h.emitGroupStateRefreshRealtime(invite.GroupID)

	writeJSON(w, http.StatusOK, invite)
}

func (h *HandlerStruct) CreateGroupRequestHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	request, err := h.GroupRepo.CreateJoinRequest(r.PathValue("groupId"), actorID)
	if err != nil {
		handleRepositoryError(w, err, "failed to create join request")
		return
	}
	group, groupErr := h.GroupRepo.GetGroup(r.PathValue("groupId"), actorID)
	if groupErr == nil {
		h.createNotification(group.CreatorID, actorID, "group_join_request_received", request.ID)
	}
	h.emitGroupStateRefreshRealtime(request.GroupID)

	writeJSON(w, http.StatusCreated, request)
}

func (h *HandlerStruct) ListGroupRequestsHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	requests, err := h.GroupRepo.ListJoinRequests(r.PathValue("groupId"), actorID)
	if err != nil {
		handleRepositoryError(w, err, "failed to load join requests")
		return
	}

	writeJSON(w, http.StatusOK, requests)
}

func (h *HandlerStruct) RespondGroupRequestHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var payload joinRequestPayload
	if err := decodeJSON(r, &payload); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	request, err := h.GroupRepo.RespondJoinRequest(r.PathValue("requestId"), actorID, strings.TrimSpace(payload.Status))
	if err != nil {
		handleRepositoryError(w, err, "failed to update join request")
		return
	}

	// Notify the original requester about the decision.
	notifType := "group_join_request_accepted"
	if request.Status == "declined" {
		notifType = "group_join_request_declined"
	}
	h.createNotification(request.UserID, actorID, notifType, request.GroupID)

	h.emitGroupSummaryRealtime(request.GroupID)
	h.emitGroupStateRefreshRealtime(request.GroupID)

	writeJSON(w, http.StatusOK, request)
}

func (h *HandlerStruct) ListGroupMembersHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	members, err := h.GroupRepo.ListMembers(r.PathValue("groupId"), actorID)
	if err != nil {
		handleRepositoryError(w, err, "failed to load members")
		return
	}

	writeJSON(w, http.StatusOK, members)
}

func (h *HandlerStruct) RemoveGroupMemberHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	if err := h.GroupRepo.RemoveMember(r.PathValue("groupId"), r.PathValue("userId"), actorID); err != nil {
		handleRepositoryError(w, err, "failed to remove member")
		return
	}

	h.emitGroupSummaryRealtime(r.PathValue("groupId"))
	h.emitGroupStateRefreshRealtime(r.PathValue("groupId"))

	writeJSON(w, http.StatusOK, map[string]string{"message": "member removed"})
}

func (h *HandlerStruct) CreateGroupEventHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var payload eventPayload
	if err := decodeJSON(r, &payload); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	eventTime, err := parseEventTime(payload.EventTime)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := validateEventTimeNotPast(eventTime); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !validTitle(payload.Title) {
		sendJSONError(w, "title must be between 3 and 80 characters", http.StatusBadRequest)
		return
	}
	if !validDescription(payload.Description) {
	sendJSONError(w, "description must be at most 300 characters", http.StatusBadRequest)
	return
	}
	event, err := h.GroupRepo.CreateEvent(r.PathValue("groupId"), actorID, strings.TrimSpace(payload.Title), strings.TrimSpace(payload.Description), eventTime)
	if err != nil {
		handleRepositoryError(w, err, "failed to create event")
		return
	}
	memberIDs, memberErr := h.GroupRepo.MemberIDs(event.GroupID)
	if memberErr == nil {
		for _, memberID := range memberIDs {
			if memberID == actorID {
				continue
			}
			h.createNotification(memberID, actorID, "group_event_created", event.ID)
		}
	}
	h.emitGroupCalendarRealtime(event.GroupID, event.ID)

	writeJSON(w, http.StatusCreated, event)
}

func (h *HandlerStruct) ListGroupEventsHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	events, err := h.GroupRepo.ListEvents(r.PathValue("groupId"), actorID)
	if err != nil {
		handleRepositoryError(w, err, "failed to load events")
		return
	}

	writeJSON(w, http.StatusOK, events)
}

func (h *HandlerStruct) GetEventHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	event, err := h.GroupRepo.GetEvent(r.PathValue("eventId"), actorID)
	if err != nil {
		handleRepositoryError(w, err, "failed to load event")
		return
	}

	writeJSON(w, http.StatusOK, event)
}

func (h *HandlerStruct) UpdateEventHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var payload eventPayload
	if err := decodeJSON(r, &payload); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	eventTime, err := parseEventTime(payload.EventTime)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	existing, err := h.GroupRepo.GetEvent(r.PathValue("eventId"), actorID)
	if err != nil {
		handleRepositoryError(w, err, "failed to load event")
		return
	}
	// Only reject a past time if the user is actually changing it — editing the
	// title or description of an event whose time has already passed is still valid.
	if !eventTime.UTC().Truncate(time.Minute).Equal(existing.EventTime.UTC().Truncate(time.Minute)) {
		if err := validateEventTimeNotPast(eventTime); err != nil {
			sendJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if !validTitle(payload.Title) {
		sendJSONError(w, "title must be between 3 and 80 characters", http.StatusBadRequest)
		return
	}
	if !validDescription(payload.Description) {
	sendJSONError(w, "description must be at most 300 characters", http.StatusBadRequest)
	return
	}
	event, err := h.GroupRepo.UpdateEvent(r.PathValue("eventId"), actorID, strings.TrimSpace(payload.Title), strings.TrimSpace(payload.Description), eventTime)
	if err != nil {
		handleRepositoryError(w, err, "failed to update event")
		return
	}

	memberIDs, memberErr := h.GroupRepo.MemberIDs(event.GroupID)
	if memberErr == nil {
		for _, memberID := range memberIDs {
			if memberID == actorID {
				continue
			}
			h.createNotification(memberID, actorID, "group_event_updated", event.ID)
		}
	}

	h.emitGroupCalendarRealtime(event.GroupID, event.ID)

	writeJSON(w, http.StatusOK, event)
}

func (h *HandlerStruct) DeleteEventHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	event, eventErr := h.GroupRepo.GetEvent(r.PathValue("eventId"), actorID)
	if eventErr != nil {
		handleRepositoryError(w, eventErr, "failed to load event")
		return
	}

	type deletedEventPayload struct {
		GroupID    string `json:"groupId"`
		EventTitle string `json:"eventTitle"`
	}
	deletedPayloadBytes, _ := json.Marshal(deletedEventPayload{
		GroupID:    event.GroupID,
		EventTitle: event.Title,
	})
	deletedPayload := string(deletedPayloadBytes)

	memberIDs, memberErr := h.GroupRepo.MemberIDs(event.GroupID)

	if err := h.GroupRepo.DeleteEvent(r.PathValue("eventId"), actorID); err != nil {
		handleRepositoryError(w, err, "failed to delete event")
		return
	}

	if memberErr == nil {
		for _, memberID := range memberIDs {
			if memberID == actorID {
				continue
			}
			h.createNotification(memberID, actorID, "group_event_deleted", deletedPayload)
		}
	}

	h.emitGroupCalendarDeleteRealtime(event.GroupID, event.ID)

	writeJSON(w, http.StatusOK, map[string]string{"message": "event deleted"})
}

func (h *HandlerStruct) RespondEventHandler(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userIDFromRequest(r)
	if !ok {
		sendJSONError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var payload eventResponsePayload
	if err := decodeJSON(r, &payload); err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := h.GroupRepo.UpsertEventResponse(r.PathValue("eventId"), actorID, strings.TrimSpace(payload.Response))
	if err != nil {
		handleRepositoryError(w, err, "failed to save event response")
		return
	}

	event, eventErr := h.GroupRepo.GetEvent(response.EventID, actorID)
	if eventErr == nil {
		h.emitGroupCalendarRealtime(event.GroupID, event.ID)
	}

	writeJSON(w, http.StatusOK, response)
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return errors.New("invalid JSON body")
	}

	return nil
}

func validTitle(title string) bool {
	trimmed := strings.TrimSpace(title)
	return len(trimmed) >= 3 && len(trimmed) <= 80
}
func validDescription(desc string) bool {
	trimmed := strings.TrimSpace(desc)
	return len(trimmed) <= 300
}
func parseEventTime(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, errors.New("eventTime is required")
	}

	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err == nil {
		return parsed, nil
	}

	parsed, err = time.Parse("2006-01-02T15:04", trimmed)
	if err != nil {
		return time.Time{}, errors.New("eventTime must be ISO formatted")
	}

	return parsed, nil
}

func validateEventTimeNotPast(eventTime time.Time) error {
	now := time.Now()
	currentMinute := now.Truncate(time.Minute)
	if eventTime.Before(currentMinute) {
		return errors.New("eventTime cannot be in the past")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func handleRepositoryError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		sendJSONError(w, "resource not found", http.StatusNotFound)
	case repository.IsConflictError(err):
		sendJSONError(w, err.Error(), http.StatusConflict)
	case strings.Contains(strings.ToLower(err.Error()), "only "):
		sendJSONError(w, err.Error(), http.StatusForbidden)
	case strings.Contains(strings.ToLower(err.Error()), "already"):
		sendJSONError(w, err.Error(), http.StatusConflict)
	case strings.Contains(strings.ToLower(err.Error()), "must be"):
		sendJSONError(w, err.Error(), http.StatusBadRequest)
	case strings.Contains(strings.ToLower(err.Error()), "required"):
		sendJSONError(w, err.Error(), http.StatusBadRequest)
	case strings.Contains(strings.ToLower(err.Error()), "cannot be"):
		sendJSONError(w, err.Error(), http.StatusBadRequest)
	default:
		sendJSONError(w, fallback, http.StatusInternalServerError)
	}
}
