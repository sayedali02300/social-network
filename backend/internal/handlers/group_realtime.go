package handlers

import (
	"log"
	"strings"

	"socialnetwork/backend/internal/models"
	wshandler "socialnetwork/backend/internal/ws"
)

const (
	groupSummaryEventType        = "group_summary_event"
	groupStateRefreshEventType   = "group_state_refresh"
	groupPostEventType           = "group_post_event"
	groupPostDeleteEventType     = "group_post_delete"
	groupCommentEventType        = "group_comment_event"
	groupCommentDeleteEventType  = "group_comment_delete"
	groupCalendarEventType       = "group_calendar_event"
	groupCalendarDeleteEventType = "group_calendar_delete"
)

func (h *HandlerStruct) emitGroupSummaryRealtime(groupID string) {
	if h == nil || h.GroupRepo == nil || h.Notifier == nil {
		return
	}

	memberIDs, err := h.GroupRepo.MemberIDs(strings.TrimSpace(groupID))
	if err != nil {
		log.Printf("group realtime summary member lookup failed group=%s err=%v", groupID, err)
		return
	}

	for _, memberID := range memberIDs {
		group, groupErr := h.GroupRepo.GetGroup(groupID, memberID)
		if groupErr != nil {
			log.Printf("group realtime summary load failed group=%s user=%s err=%v", groupID, memberID, groupErr)
			continue
		}

		h.Notifier.EmitToUser(memberID, wshandler.OutgoingEnvelope{
			Type:    groupSummaryEventType,
			Payload: group,
		})
	}
}

func (h *HandlerStruct) emitGroupStateRefreshRealtime(groupID string) {
	if h == nil || h.GroupRepo == nil || h.NotifierMany == nil {
		return
	}

	memberIDs, err := h.GroupRepo.MemberIDs(strings.TrimSpace(groupID))
	if err != nil {
		log.Printf("group realtime refresh member lookup failed group=%s err=%v", groupID, err)
		return
	}

	h.NotifierMany.EmitToUsers(memberIDs, wshandler.OutgoingEnvelope{
		Type: groupStateRefreshEventType,
		Payload: map[string]string{
			"groupId": groupID,
		},
	})
}

func (h *HandlerStruct) emitGroupPostRealtime(groupID, postID, viewerID string) {
	if h == nil || h.GroupRepo == nil || h.PostRepo == nil || h.NotifierMany == nil {
		return
	}

	memberIDs, err := h.GroupRepo.MemberIDs(strings.TrimSpace(groupID))
	if err != nil {
		log.Printf("group realtime post member lookup failed group=%s err=%v", groupID, err)
		return
	}
	if len(memberIDs) == 0 {
		return
	}

	post, err := h.PostRepo.GetSinglePost(postID, viewerID)
	if err != nil || post == nil {
		fallbackViewerID := viewerID
		if strings.TrimSpace(fallbackViewerID) == "" {
			fallbackViewerID = memberIDs[0]
		}
		post, err = h.PostRepo.GetSinglePost(postID, fallbackViewerID)
	}
	if err != nil || post == nil {
		log.Printf("group realtime post load failed group=%s post=%s err=%v", groupID, postID, err)
		return
	}

	h.NotifierMany.EmitToUsers(memberIDs, wshandler.OutgoingEnvelope{
		Type:    groupPostEventType,
		Payload: post,
	})
}

func (h *HandlerStruct) emitGroupPostDeleteRealtime(groupID, postID string) {
	if h == nil || h.GroupRepo == nil || h.NotifierMany == nil {
		return
	}

	memberIDs, err := h.GroupRepo.MemberIDs(strings.TrimSpace(groupID))
	if err != nil {
		log.Printf("group realtime post delete member lookup failed group=%s err=%v", groupID, err)
		return
	}

	h.NotifierMany.EmitToUsers(memberIDs, wshandler.OutgoingEnvelope{
		Type: groupPostDeleteEventType,
		Payload: map[string]string{
			"groupId": groupID,
			"postId":  postID,
		},
	})
}

func (h *HandlerStruct) emitGroupCommentRealtime(groupID, commentID string) {
	if h == nil || h.GroupRepo == nil || h.CommentRepo == nil || h.NotifierMany == nil {
		return
	}

	memberIDs, err := h.GroupRepo.MemberIDs(strings.TrimSpace(groupID))
	if err != nil {
		log.Printf("group realtime comment member lookup failed group=%s err=%v", groupID, err)
		return
	}

	comment, err := h.CommentRepo.GetCommentByID(commentID)
	if err != nil || comment == nil {
		log.Printf("group realtime comment load failed group=%s comment=%s err=%v", groupID, commentID, err)
		return
	}

	h.NotifierMany.EmitToUsers(memberIDs, wshandler.OutgoingEnvelope{
		Type:    groupCommentEventType,
		Payload: comment,
	})
}

func (h *HandlerStruct) emitGroupCommentDeleteRealtime(groupID, postID, commentID string) {
	if h == nil || h.GroupRepo == nil || h.NotifierMany == nil {
		return
	}

	memberIDs, err := h.GroupRepo.MemberIDs(strings.TrimSpace(groupID))
	if err != nil {
		log.Printf("group realtime comment delete member lookup failed group=%s err=%v", groupID, err)
		return
	}

	h.NotifierMany.EmitToUsers(memberIDs, wshandler.OutgoingEnvelope{
		Type: groupCommentDeleteEventType,
		Payload: map[string]string{
			"groupId":   groupID,
			"postId":    postID,
			"commentId": commentID,
		},
	})
}

func (h *HandlerStruct) emitGroupCalendarRealtime(groupID, eventID string) {
	if h == nil || h.GroupRepo == nil || h.Notifier == nil {
		return
	}

	memberIDs, err := h.GroupRepo.MemberIDs(strings.TrimSpace(groupID))
	if err != nil {
		log.Printf("group realtime event member lookup failed group=%s err=%v", groupID, err)
		return
	}

	for _, memberID := range memberIDs {
		event, eventErr := h.GroupRepo.GetEvent(eventID, memberID)
		if eventErr != nil {
			log.Printf("group realtime event load failed group=%s event=%s user=%s err=%v", groupID, eventID, memberID, eventErr)
			continue
		}

		h.Notifier.EmitToUser(memberID, wshandler.OutgoingEnvelope{
			Type:    groupCalendarEventType,
			Payload: event,
		})
	}
}

func (h *HandlerStruct) emitGroupCalendarDeleteRealtime(groupID, eventID string) {
	if h == nil || h.GroupRepo == nil || h.NotifierMany == nil {
		return
	}

	memberIDs, err := h.GroupRepo.MemberIDs(strings.TrimSpace(groupID))
	if err != nil {
		log.Printf("group realtime event delete member lookup failed group=%s err=%v", groupID, err)
		return
	}

	h.NotifierMany.EmitToUsers(memberIDs, wshandler.OutgoingEnvelope{
		Type: groupCalendarDeleteEventType,
		Payload: map[string]string{
			"groupId": groupID,
			"eventId": eventID,
		},
	})
}

func isGroupPost(post *models.Post) bool {
	return post != nil && strings.TrimSpace(post.GroupID) != ""
}
