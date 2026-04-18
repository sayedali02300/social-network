<template>
  <section class="page">
    <header class="header-card">
      <div class="header-content">
        <router-link class="back-link" to="/groups">Back to groups</router-link>
        <div class="title-row">
          <h1>{{ group?.title || 'Loading group...' }}</h1>
          <div v-if="isCreator" class="settings-wrap">
            <button
              class="settings-trigger secondary"
              aria-label="Open group settings"
              title="Group settings"
              @click="toggleSettingsMenu"
            >
              <Cog6ToothIcon />
            </button>
          </div>
        </div>
        <p class="description">{{ group?.description }}</p>
      </div>
    </header>

    <p v-if="loading" class="message">Loading group workspace...</p>
    <p v-else-if="error" class="message error">{{ error }}</p>

    <template v-else-if="group">
      <section class="summary-grid">
        <article class="summary-card">
          <span class="label">Creator</span>
          <strong>{{ group.creatorName }}</strong>
        </article>
        <article class="summary-card">
          <span class="label">Members</span>
          <strong>{{ group.membersCount }}</strong>
        </article>
        <article class="summary-card">
          <span class="label">Status</span>
          <strong>{{ currentUserStatusLabel }}</strong>
        </article>
      </section>

      <section v-if="isCreator && pendingJoinRequests.length > 0" class="moderation-grid">
        <article class="panel moderation-panel">
          <div class="panel-header">
            <div>
              <h2>Join Request</h2>
            </div>
            <span class="request-counter">{{ pendingJoinRequests.length }} pending</span>
          </div>

          <ul class="request-list">
            <li
              v-for="request in pendingJoinRequests"
              :key="request.id"
              class="request-item"
              :data-request-id="request.id"
            >
              <div class="request-meta">
                <strong>{{ requestDisplayNames[request.userId] || request.userId }}</strong>
              </div>
              <div class="queue-actions">
                <button
                  class="secondary compact"
                  :aria-label="`Accept join request from ${request.userId}`"
                  @click="handleJoinRequestResponse(request.id, 'accepted')"
                >
                  Accept request
                </button>
                <button
                  class="danger compact"
                  :aria-label="`Decline join request from ${request.userId}`"
                  @click="handleJoinRequestResponse(request.id, 'declined')"
                >
                  Decline request
                </button>
              </div>
            </li>
          </ul>
        </article>
      </section>

      <section class="action-row">
        <button v-if="canRequestJoin" @click="handleJoinRequest">Request to join</button>
        <button
          v-if="group.hasPendingInvite && myPendingInvite"
          class="secondary"
          @click="handleInviteResponse(myPendingInvite.id, 'accepted')"
        >
          Accept invite
        </button>
        <button
          v-if="group.hasPendingInvite && myPendingInvite"
          class="danger"
          @click="handleInviteResponse(myPendingInvite.id, 'declined')"
        >
          Decline invite
        </button>
        <router-link v-if="group.isMember" class="secondary-link" :to="`/groups/${group.id}/events`">
          Create Event
        </router-link>
        <router-link
          v-if="group.isMember"
          class="secondary-link"
          :to="`/chats/groups/${group.id}?name=${encodeURIComponent(group.title)}`"
        >
          Open group chat
        </router-link>
        <button v-if="group.isMember" type="button" class="secondary" @click="openMemberInviteModal">
          Invite followers
        </button>
        <button v-if="group.isMember" type="button" class="danger" @click="handleLeaveGroup">
          Leave group
        </button>
      </section>

      <div v-if="showMemberInviteModal" class="member-invite-overlay" @click="closeMemberInviteModal">
        <div class="member-invite-popover" @click.stop>
          <div class="member-invite-header">
            <div>
              <h3>Invite followers</h3>
              <p>Choose followers to invite into this group.</p>
            </div>
            <button type="button" class="settings-close secondary" @click="closeMemberInviteModal">
              <XMarkIcon />
            </button>
          </div>

          <div class="member-invite-toolbar">
            <input
              v-model.trim="memberInviteSearch"
              type="text"
              placeholder="Search followers"
              :disabled="memberInviteLoading || memberInviteSubmitting"
            />
            <button
              type="button"
              class="secondary"
              :disabled="memberInviteSubmitting || selectedMemberInviteIds.length === 0"
              @click="submitMemberInvites"
            >
              {{ memberInviteSubmitting ? 'Sending...' : `Invite (${selectedMemberInviteIds.length})` }}
            </button>
          </div>

          <p v-if="memberInviteError" class="message error compact-message">{{ memberInviteError }}</p>
          <p v-else-if="memberInviteLoading" class="small-copy">Loading followers...</p>
          <p v-else-if="memberInviteCandidates.length === 0" class="small-copy">
            No followers available to invite.
          </p>

          <ul v-else class="member-invite-list">
            <li v-for="user in memberInviteCandidates" :key="user.id" class="member-invite-item">
              <label class="member-invite-option">
                <input
                  type="checkbox"
                  :checked="selectedMemberInviteIds.includes(user.id)"
                  :disabled="memberInviteSubmitting"
                  @change="toggleMemberInviteSelection(user.id)"
                />
                <span class="member-invite-copy">
                  <strong>{{ formatUserDisplayName(user) }}</strong>
                  <small>{{ user.firstName }} {{ user.lastName }}</small>
                </span>
              </label>
            </li>
          </ul>
        </div>
      </div>

      <p v-if="feedback" class="message success">{{ feedback }}</p>

      <section v-if="group.isMember" class="panel group-posts-panel">
        <div class="panel-header">
          <h2>Group posts</h2>
          <span class="label">{{ groupPosts.length }} posts</span>
        </div>

        <form class="group-post-form" @submit.prevent="handleCreateGroupPost">
          <input
            v-model.trim="postTitle"
            type="text"
            placeholder="Post title"
            maxlength="60"
            :disabled="submittingPost"
          />
          <span class="char-counter" :class="{ 'char-warn': postTitle.length >= 50, 'char-limit': postTitle.length >= 60 }">
            {{ postTitle.length }} / 60
          </span>
          <textarea
            v-model.trim="postBody"
            rows="4"
            placeholder="Share something with the group"
            maxlength="5000"
            :disabled="submittingPost"
          />
          <span class="char-counter" :class="{ 'char-warn': postBody.length >= 4800, 'char-limit': postBody.length >= 5000 }">
            {{ postBody.length }} / 5000
          </span>
          <div class="group-post-actions">
            <label class="secondary-link upload-trigger" for="group-post-image">
              {{ postImage ? postImage.name : 'Add image' }}
            </label>
            <input
              id="group-post-image"
              class="hidden-file-input"
              type="file"
              accept="image/jpeg, image/png, image/gif"
              @change="handlePostImageChange"
            />
            <button type="submit" class="secondary" :disabled="submittingPost">
              {{ submittingPost ? 'Posting...' : 'Publish post' }}
            </button>
          </div>
        </form>

        <p v-if="groupPosts.length === 0" class="empty-copy">No group posts yet.</p>
        <ul v-else class="gp-post-list">
          <li v-for="post in groupPosts" :key="post.id" class="gp-post">

            <!-- Author row -->
            <div class="gp-author-row">
              <div class="gp-avatar-col">
                <img v-if="post.author?.avatar" :src="getImgURL(post.author.avatar)" class="gp-avatar" alt="avatar" />
                <div v-else class="gp-avatar gp-avatar-fallback">
                  {{ ((post.author?.firstName?.[0] || '') + (post.author?.lastName?.[0] || '')).toUpperCase() || '?' }}
                </div>
              </div>
              <div class="gp-author-meta">
                <span class="gp-display-name">
                  {{ post.author?.nickname || ((post.author?.firstName || '') + ' ' + (post.author?.lastName || '')).trim() }}
                </span>
                <span v-if="isGroupCreatorPost(post)" class="gp-creator-badge">Creator</span>
                <span class="gp-handle" v-if="post.author?.nickname">@{{ post.author.nickname }}</span>
              </div>
              <div class="gp-post-actions-wrap">
                <div v-if="canManagePost(post)" class="post-menu-wrap">
                  <button class="secondary post-menu-trigger" type="button" aria-label="Open post actions" @click.stop="togglePostMenu(post.id)">
                    <EllipsisHorizontalIcon />
                  </button>
                  <div v-if="activePostMenuId === post.id" class="post-menu-popover">
                    <button type="button" class="post-menu-item" @click="openEditPostDialog(post)">Edit post</button>
                    <button type="button" class="post-menu-item danger-text" @click="openDeletePostConfirm(post.id)">Delete post</button>
                  </div>
                </div>
              </div>
            </div>

            <!-- Post title -->
            <p v-if="post.title" class="gp-post-title">{{ post.title }}</p>

            <!-- Post body -->
            <div class="gp-post-body">{{ post.content }}</div>

            <!-- Post image -->
            <div v-if="post.imagePath" class="gp-post-image-wrap">
              <img :src="getImgURL(post.imagePath)" alt="Post image" class="gp-post-image" />
            </div>

            <!-- Timestamp -->
            <div class="gp-timestamp">{{ formatPostDate(post.createdAt) }}</div>

            <div class="gp-divider"></div>

            <!-- Stats row -->
            <div class="gp-stats-row">
              <span class="gp-stat">
                <strong>{{ commentsByPost[post.id]?.length || 0 }}</strong>
                Comment{{ (commentsByPost[post.id]?.length || 0) !== 1 ? 's' : '' }}
              </span>
              <span class="gp-stat">
                <strong>{{ groupPostLikes[post.id]?.likeCount ?? post.likeCount ?? 0 }}</strong>
                Like{{ (groupPostLikes[post.id]?.likeCount ?? post.likeCount ?? 0) !== 1 ? 's' : '' }}
              </span>
              <span class="gp-stat">
                <strong>{{ groupPostLikes[post.id]?.dislikeCount ?? post.dislikeCount ?? 0 }}</strong>
                Dislike{{ (groupPostLikes[post.id]?.dislikeCount ?? post.dislikeCount ?? 0) !== 1 ? 's' : '' }}
              </span>
            </div>

            <div class="gp-divider"></div>

            <!-- Action bar -->
            <div class="gp-action-bar">
              <button class="gp-action-btn" title="Comment">
                <svg class="gp-action-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>
                <span class="gp-action-count">{{ commentsByPost[post.id]?.length || 0 }}</span>
              </button>
              <button
                class="gp-action-btn gp-like-btn"
                :class="{ 'gp-active-like': (groupPostLikes[post.id]?.myReaction ?? post.myReaction ?? 0) === 1 }"
                title="Like"
                @click="handleGroupPostReaction(post.id, 1)"
              >
                <svg class="gp-action-icon" viewBox="0 0 24 24" :fill="(groupPostLikes[post.id]?.myReaction ?? post.myReaction ?? 0) === 1 ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="1.8"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/></svg>
                <span class="gp-action-count">{{ groupPostLikes[post.id]?.likeCount ?? post.likeCount ?? 0 }}</span>
              </button>
              <button
                class="gp-action-btn gp-dislike-btn"
                :class="{ 'gp-active-dislike': (groupPostLikes[post.id]?.myReaction ?? post.myReaction ?? 0) === -1 }"
                title="Dislike"
                @click="handleGroupPostReaction(post.id, -1)"
              >
                <svg class="gp-action-icon" viewBox="0 0 24 24" :fill="(groupPostLikes[post.id]?.myReaction ?? post.myReaction ?? 0) === -1 ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="1.8"><path d="M10 15v4a3 3 0 0 0 3 3l4-9V2H5.72a2 2 0 0 0-2 1.7l-1.38 9a2 2 0 0 0 2 2.3H10z"/><path d="M17 2h2.67A2.31 2.31 0 0 1 22 4v7a2.31 2.31 0 0 1-2.33 2H17"/></svg>
                <span class="gp-action-count">{{ groupPostLikes[post.id]?.dislikeCount ?? post.dislikeCount ?? 0 }}</span>
              </button>
              <span v-if="likeErrorByPost[post.id]" class="gp-like-error">{{ likeErrorByPost[post.id] }}</span>
            </div>

            <div class="gp-divider"></div>

            <!-- Reply composer -->
            <form class="gp-composer" @submit.prevent="submitGroupPostComment(post.id)">
              <div class="gp-composer-body">
                <textarea
                  v-model.trim="commentDraftByPost[post.id]"
                  class="gp-composer-input"
                  placeholder="Post your reply…"
                  rows="2"
                  :disabled="submittingCommentPostId === post.id"
                ></textarea>
                <div class="gp-composer-footer">
                  <label class="gp-attach-btn" :for="`group-comment-image-${post.id}`" title="Attach image">
                    <input :id="`group-comment-image-${post.id}`" class="hidden-file-input" type="file" accept="image/jpeg, image/png, image/gif" @change="handleCommentImageChange(post.id, $event)" />
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>
                    <span v-if="commentImageByPost[post.id]" class="gp-file-name">{{ commentImageByPost[post.id]?.name }}</span>
                  </label>
                  <button class="gp-reply-btn" type="submit" :disabled="(!commentDraftByPost[post.id]?.trim() && !commentImageByPost[post.id]) || submittingCommentPostId === post.id">
                    <span v-if="submittingCommentPostId !== post.id">Reply</span>
                    <span v-else>…</span>
                  </button>
                </div>
                <p v-if="commentErrorByPost[post.id]" class="gp-form-error">{{ commentErrorByPost[post.id] }}</p>
              </div>
            </form>

            <div class="gp-divider"></div>

            <!-- Comments list -->
            <p v-if="commentsLoadingByPost[post.id]" class="gp-loading">Loading replies…</p>
            <p v-else-if="!commentsByPost[post.id] || commentsByPost[post.id]?.length === 0" class="gp-empty-comments">No replies yet.</p>
            <template v-else>
              <div v-for="comment in commentsByPost[post.id]" :key="comment.id" class="gp-comment-row">
                <div class="gp-comment-left">
                  <img v-if="comment.author?.avatar" :src="getImgURL(comment.author.avatar)" class="gp-avatar gp-avatar-sm" alt="avatar" />
                  <div v-else class="gp-avatar gp-avatar-sm gp-avatar-fallback">
                    {{ ((comment.author?.firstName?.[0] || '?') + (comment.author?.lastName?.[0] || '?')).toUpperCase() }}
                  </div>
                </div>
                <div class="gp-comment-content">
                  <div class="gp-comment-header">
                    <span class="gp-comment-name">
                      {{ comment.author?.nickname || ((comment.author?.firstName || '') + ' ' + (comment.author?.lastName || '')).trim() || 'User' }}
                    </span>
                    <span class="gp-comment-dot">·</span>
                    <span class="gp-comment-time">{{ formatPostDate(comment.createdAt) }}</span>
                    <div v-if="canManageComment(comment)" class="comment-menu-wrap">
                      <button class="secondary comment-menu-trigger" type="button" aria-label="Open comment actions" @click.stop="toggleCommentMenu(post.id, comment.id)">
                        <EllipsisHorizontalIcon />
                      </button>
                      <div v-if="activeCommentMenuKey === getCommentMenuKey(post.id, comment.id)" class="comment-menu-popover">
                        <button type="button" class="comment-menu-item" @click="openEditCommentDialog(post.id, comment)">Edit</button>
                        <button type="button" class="comment-menu-item danger-text" @click="openDeleteCommentConfirm(post.id, comment.id)">Delete</button>
                      </div>
                    </div>
                  </div>
                  <p class="gp-comment-text">{{ comment.content }}</p>
                  <img v-if="comment.imagePath" class="gp-comment-image" :src="getImgURL(comment.imagePath)" alt="Comment image" />
                </div>
              </div>
            </template>

          </li>
        </ul>
      </section>

    </template>

    <div v-if="showSettingsMenu" class="settings-overlay" @click="closeSettingsMenu">
      <div class="settings-popover" @click.stop>
        <div class="settings-header">
          <h3>Group settings</h3>
          <button class="settings-close secondary" @click="closeSettingsMenu">
            <XMarkIcon />
          </button>
        </div>

        <div class="settings-nav">
          <div class="accordion-item">
            <button class="secondary nav-action" @click="selectCreatorPanel('edit')">
              <span>Edit</span>
              <ChevronUpIcon v-if="activeCreatorPanel === 'edit'" />
              <ChevronDownIcon v-else />
            </button>
            <div v-if="activeCreatorPanel === 'edit'" class="accordion-content">
              <div class="inline-form">
                <input v-model.trim="editTitle" placeholder="Group title" />
              </div>
              <textarea v-model.trim="editDescription" rows="3" placeholder="Group description" />
              <div class="inline-form">
                <button class="secondary" :disabled="submittingGroup" @click="handleUpdateGroup">
                  {{ submittingGroup ? 'Saving...' : 'Save group changes' }}
                </button>
              </div>
            </div>
          </div>

          <div class="accordion-item">
            <button class="secondary nav-action" @click="selectCreatorPanel('invite')">
              <span>Invite</span>
              <ChevronUpIcon v-if="activeCreatorPanel === 'invite'" />
              <ChevronDownIcon v-else />
            </button>
            <div v-if="activeCreatorPanel === 'invite'" class="accordion-content">
              <div class="invite-search">
                <div class="inline-form">
                  <input
                    v-model.trim="inviteSearchTerm"
                    placeholder="Search by username or name"
                    @focus="handleInviteInputFocus"
                  />
                  <button class="secondary" :disabled="submittingInvite || !selectedInviteUserId" @click="handleCreateInvite">
                    {{ submittingInvite ? 'Sending...' : 'Send invite' }}
                  </button>
                </div>
                <ul v-if="showInviteSuggestions && inviteSuggestions.length > 0" class="invite-suggestions">
                  <li v-for="user in inviteSuggestions" :key="user.id">
                    <button type="button" class="invite-suggestion" @click="selectInviteUser(user)">
                      <strong>{{ formatUserDisplayName(user) }}</strong>
                      <small>{{ user.firstName }} {{ user.lastName }}</small>
                    </button>
                  </li>
                </ul>
                <p v-else-if="showInviteSuggestions && inviteSearchTerm.length > 0 && !inviteSearchLoading" class="small-copy">
                  No matching users.
                </p>
                <p v-if="inviteSearchLoading" class="small-copy">Searching...</p>
              </div>
              <ul class="item-list">
                <li v-for="invite in invites" :key="invite.id">
                  <span>{{ inviteDisplayNames[invite.receiverId] || invite.receiverId }}</span>
                  <small>{{ invite.status }}</small>
                </li>
              </ul>
            </div>
          </div>

          <div class="accordion-item">
            <button class="secondary nav-action" @click="selectCreatorPanel('members')">
              <span>Members preview</span>
              <ChevronUpIcon v-if="activeCreatorPanel === 'members'" />
              <ChevronDownIcon v-else />
            </button>
            <div v-if="activeCreatorPanel === 'members'" class="accordion-content">
              <div class="panel-header">
                <router-link :to="`/groups/${group?.id}/members`">Open full list</router-link>
              </div>
              <ul class="item-list">
                <li v-for="member in members.slice(0, 4)" :key="member.userId">
                  <span>{{ member.nickname }}</span>
                  <small>{{ formatMemberRole(member.role) }}</small>
                </li>
              </ul>
            </div>
          </div>

          <div class="accordion-item">
            <button class="secondary nav-action" @click="selectCreatorPanel('requests')">
              <span>Join request</span>
              <ChevronUpIcon v-if="activeCreatorPanel === 'requests'" />
              <ChevronDownIcon v-else />
            </button>
            <div v-if="activeCreatorPanel === 'requests'" class="accordion-content">
              <ul class="item-list">
                <li v-for="request in joinRequests" :key="request.id">
                  <span>{{ requestDisplayNames[request.userId] || request.userId }}</span>
                  <div class="queue-actions" v-if="request.status === 'pending'">
                    <button class="secondary compact" @click="handleJoinRequestResponse(request.id, 'accepted')">
                      Accept
                    </button>
                    <button class="danger compact" @click="handleJoinRequestResponse(request.id, 'declined')">
                      Decline
                    </button>
                  </div>
                  <small v-else>{{ request.status }}</small>
                </li>
              </ul>
            </div>
          </div>

          <div class="accordion-item">
            <button class="secondary nav-action" @click="selectCreatorPanel('events')">
              <span>Upcoming event</span>
              <ChevronUpIcon v-if="activeCreatorPanel === 'events'" />
              <ChevronDownIcon v-else />
            </button>
            <div v-if="activeCreatorPanel === 'events'" class="accordion-content">
              <div class="panel-header">
                <router-link :to="`/groups/${group?.id}/events`">Open events board</router-link>
              </div>
              <ul class="item-list">
                <li v-for="event in events.slice(0, 4)" :key="event.id">
                  <span>{{ event.title }}</span>
                  <small>{{ formatDate(event.eventTime) }}</small>
                </li>
              </ul>
            </div>
          </div>

          <button class="danger nav-action delete-nav" @click="openDeleteConfirm">
            <span>Delete</span>
            <ChevronRightIcon />
          </button>
        </div>
      </div>
    </div>

    <div v-if="leaveConfirmOpen" class="confirm-overlay" @click="leaveConfirmOpen = false">
      <div class="confirm-dialog" @click.stop>
        <h3>Leave group</h3>
        <p>{{ leaveConfirmMessage }}</p>
        <div class="confirm-actions">
          <button class="secondary" @click="leaveConfirmOpen = false">Cancel</button>
          <button class="danger" @click="confirmLeaveGroup">Leave</button>
        </div>
      </div>
    </div>

    <div v-if="deleteConfirmOpen" class="confirm-overlay" @click="deleteConfirmOpen = false">
      <div class="confirm-dialog" @click.stop>
        <h3>Delete group</h3>
        <p>Are you sure you want to delete group?</p>
        <div class="confirm-actions">
          <button class="secondary" @click="deleteConfirmOpen = false">Cancel</button>
          <button class="danger" :disabled="deletingGroup" @click="confirmDeleteGroup">
            {{ deletingGroup ? 'Deleting...' : 'Delete group' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="editingPostId" class="settings-overlay" @click="closeEditPostDialog">
      <div class="settings-popover" @click.stop>
        <div class="settings-header">
          <h3>Edit post</h3>
          <button class="settings-close secondary" @click="closeEditPostDialog">
            <XMarkIcon />
          </button>
        </div>

        <div class="accordion-content">
          <div class="inline-form">
            <input v-model.trim="editPostTitle" placeholder="Post title" />
          </div>
          <textarea
            v-model.trim="editPostBody"
            rows="5"
            placeholder="Update your post"
            class="edit-post-textarea"
          />
          <div class="confirm-actions">
            <button class="secondary" type="button" @click="closeEditPostDialog">Cancel</button>
            <button class="secondary" :disabled="savingPostEdit" @click="handleSavePostEdit">
              {{ savingPostEdit ? 'Saving...' : 'Save changes' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="postDeleteConfirmId" class="confirm-overlay" @click="closeDeletePostConfirm">
      <div class="confirm-dialog" @click.stop>
        <h3>Delete post</h3>
        <p>Are you sure you want to delete this group post?</p>
        <div class="confirm-actions">
          <button class="secondary" @click="closeDeletePostConfirm">Cancel</button>
          <button class="danger" :disabled="deletingPostId === postDeleteConfirmId" @click="confirmDeletePost">
            {{ deletingPostId === postDeleteConfirmId ? 'Deleting...' : 'Delete post' }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="editingCommentId" class="settings-overlay" @click="closeEditCommentDialog">
      <div class="settings-popover" @click.stop>
        <div class="settings-header">
          <h3>Edit comment</h3>
          <button class="settings-close secondary" @click="closeEditCommentDialog">
            <XMarkIcon />
          </button>
        </div>

        <div class="accordion-content">
          <textarea
            v-model.trim="editCommentContent"
            rows="4"
            placeholder="Update your comment"
            class="edit-post-textarea"
          />
          <div class="confirm-actions">
            <button class="secondary" type="button" @click="closeEditCommentDialog">Cancel</button>
            <button class="secondary" :disabled="savingCommentEdit" @click="handleSaveCommentEdit">
              {{ savingCommentEdit ? 'Saving...' : 'Save changes' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="commentDeleteTarget" class="confirm-overlay" @click="closeDeleteCommentConfirm">
      <div class="confirm-dialog" @click.stop>
        <h3>Delete comment</h3>
        <p>Are you sure you want to delete this comment?</p>
        <div class="confirm-actions">
          <button class="secondary" @click="closeDeleteCommentConfirm">Cancel</button>
          <button class="danger" :disabled="deletingCommentId === commentDeleteTarget.commentId" @click="confirmDeleteComment">
            {{ deletingCommentId === commentDeleteTarget.commentId ? 'Deleting...' : 'Delete comment' }}
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import {
  ChevronDownIcon,
  ChevronRightIcon,
  ChevronUpIcon,
  Cog6ToothIcon,
  EllipsisHorizontalIcon,
  XMarkIcon,
} from '@heroicons/vue/24/outline'
import { useRoute, useRouter } from 'vue-router'

import { apiURL } from '@/api/api'
import { API_BASE_URL, API_ROUTES } from '@/api/api'
import { getUserFollowers, type ConnectionUser } from '@/api/followers'
import { buildWebSocketURL } from '@/api/websocket'
import {
  deleteComment,
  deleteGroupPost,
  createGroupPost,
  createGroupInvite,
  createJoinRequest,
  deleteGroup,
  getGroup,
  listEvents,
  listGroupInvites,
  listGroupMembers,
  listGroupPosts,
  listJoinRequests,
  removeGroupMember,
  respondToInvite,
  respondToJoinRequest,
  updateComment,
  updateGroupPost,
  updateGroup,
} from '@/api/groups'
import { fetchSessionData } from '@/router'
import { formatPostDate, getImgURL } from '@/utils/helpers'
import type { Post } from '@/types/post'
import type {
  EventItem,
  GroupInvite,
  GroupJoinRequest,
  GroupMember,
  GroupSummary,
} from '@/types/groups'
import type { SessionData, User } from '@/types/User'

type GroupPostComment = {
  id: string
  postId: string
  content: string
  imagePath?: string
  createdAt: string
  author: {
    user_id?: string
    avatar?: string
    nickname?: string
    firstName?: string
    lastName?: string
  }
}

type GroupStateRefreshPayload = {
  groupId: string
}

type GroupPostDeletePayload = {
  groupId: string
  postId: string
}

type GroupCommentDeletePayload = {
  groupId: string
  postId: string
  commentId: string
}

type GroupCalendarDeletePayload = {
  groupId: string
  eventId: string
}

const route = useRoute()
const router = useRouter()
const group = ref<GroupSummary | null>(null)
const members = ref<GroupMember[]>([])
const invites = ref<GroupInvite[]>([])
const joinRequests = ref<GroupJoinRequest[]>([])
const events = ref<EventItem[]>([])
const groupPosts = ref<Post[]>([])
const commentsByPost = ref<Record<string, GroupPostComment[]>>({})
const commentsLoadingByPost = ref<Record<string, boolean>>({})
const loading = ref(false)
const error = ref('')
const feedback = ref('')
const inviteSearchTerm = ref('')
const selectedInviteUserId = ref('')
const inviteSuggestions = ref<ConnectionUser[]>([])
const inviteFollowers = ref<ConnectionUser[]>([])
const inviteSearchLoading = ref(false)
const showInviteSuggestions = ref(false)
const showMemberInviteModal = ref(false)
const memberInviteSearch = ref('')
const memberInviteFollowers = ref<ConnectionUser[]>([])
const memberInviteLoading = ref(false)
const memberInviteSubmitting = ref(false)
const memberInviteError = ref('')
const selectedMemberInviteIds = ref<string[]>([])
const postTitle = ref('')
const postBody = ref('')
const postImage = ref<File | null>(null)
const commentDraftByPost = ref<Record<string, string>>({})
const commentImageByPost = ref<Record<string, File | null>>({})
const commentErrorByPost = ref<Record<string, string>>({})
const editTitle = ref('')
const editDescription = ref('')
const submittingGroup = ref(false)
const submittingInvite = ref(false)
const submittingPost = ref(false)
const submittingCommentPostId = ref('')
const deletingGroup = ref(false)
const deletingPostId = ref('')
const savingPostEdit = ref(false)
const savingCommentEdit = ref(false)
const showSettingsMenu = ref(false)
const deleteConfirmOpen = ref(false)
const leaveConfirmOpen = ref(false)
const leaveConfirmMessage = ref('')
const activePostMenuId = ref('')
const activeCommentMenuKey = ref('')
const editingPostId = ref('')
const editPostTitle = ref('')
const editPostBody = ref('')
const editingCommentId = ref('')
const editingCommentPostId = ref('')
const editCommentContent = ref('')
const postDeleteConfirmId = ref('')
const commentDeleteTarget = ref<{ postId: string; commentId: string } | null>(null)
const deletingCommentId = ref('')
const groupPostLikes = ref<Record<string, { likeCount: number; dislikeCount: number; myReaction: number }>>({})
const likeErrorByPost = ref<Record<string, string>>({})
const activeCreatorPanel = ref<'edit' | 'invite' | 'members' | 'requests' | 'events' | null>('members')
const currentUserId = ref('')
const requestDisplayNames = ref<Record<string, string>>({})
const inviteDisplayNames = ref<Record<string, string>>({})
let suppressInviteSearchWatcher = false
let socket: WebSocket | null = null
let reconnectTimerId: number | null = null
let socketClosedManually = false
let groupReloadInFlight = false

const groupId = computed(() => route.params.groupId as string)
const isCreator = computed(() => group.value?.creatorId === currentUserId.value)
const currentUserMember = computed(() =>
  members.value.find((member) => member.userId === currentUserId.value),
)
const pendingJoinRequests = computed(() =>
  joinRequests.value.filter((request) => request.status === 'pending'),
)
const currentUserStatusLabel = computed(() => {
  if (!group.value?.isMember) return 'Guest'
  if (isCreator.value) return 'Creator'
  return formatMemberRole(currentUserMember.value?.role)
})
const canRequestJoin = computed(
  () =>
    !!group.value &&
    !group.value.isMember &&
    !group.value.hasPendingInvite &&
    !group.value.hasPendingRequest,
)
const myPendingInvite = computed(() =>
  invites.value.find(
    (invite) => invite.receiverId === currentUserId.value && invite.status === 'pending',
  ),
)
const blockedInviteUserIds = computed(() => {
  const ids = new Set<string>()
  members.value.forEach((member) => ids.add(member.userId))
  invites.value
    .filter((invite) => invite.status === 'pending')
    .forEach((invite) => ids.add(invite.receiverId))
  if (currentUserId.value) {
    ids.add(currentUserId.value)
  }
  return ids
})
const memberInviteCandidates = computed(() => {
  const trimmed = memberInviteSearch.value.trim().toLowerCase()

  return memberInviteFollowers.value.filter((user) => {
    if (blockedInviteUserIds.value.has(user.id)) {
      return false
    }

    if (!trimmed) {
      return true
    }

    const nickname = (user.nickname || '').toLowerCase()
    const firstName = (user.firstName || '').toLowerCase()
    const lastName = (user.lastName || '').toLowerCase()
    const fullName = `${firstName} ${lastName}`.trim()

    return (
      user.id.toLowerCase().includes(trimmed) ||
      nickname.includes(trimmed) ||
      firstName.includes(trimmed) ||
      lastName.includes(trimmed) ||
      fullName.includes(trimmed)
    )
  })
})

function sortGroupPosts(items: Post[]) {
  return [...items].sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
}

function initLikeState(posts: Post[]) {
  const state: Record<string, { likeCount: number; dislikeCount: number; myReaction: number }> = {}
  posts.forEach((post) => {
    state[post.id] = {
      likeCount: post.likeCount ?? 0,
      dislikeCount: post.dislikeCount ?? 0,
      myReaction: post.myReaction ?? 0,
    }
  })
  groupPostLikes.value = state
}

async function handleGroupPostReaction(postId: string, value: 1 | -1) {
  const current = groupPostLikes.value[postId] ?? { likeCount: 0, dislikeCount: 0, myReaction: 0 }
  const isSame = current.myReaction === value
  const prev = { ...current }

  // Optimistic update
  groupPostLikes.value = {
    ...groupPostLikes.value,
    [postId]: isSame
      ? {
          likeCount: value === 1 ? current.likeCount - 1 : current.likeCount,
          dislikeCount: value === -1 ? current.dislikeCount - 1 : current.dislikeCount,
          myReaction: 0,
        }
      : {
          likeCount: value === 1 ? current.likeCount + 1 : current.myReaction === 1 ? current.likeCount - 1 : current.likeCount,
          dislikeCount: value === -1 ? current.dislikeCount + 1 : current.myReaction === -1 ? current.dislikeCount - 1 : current.dislikeCount,
          myReaction: value,
        },
  }
  likeErrorByPost.value = { ...likeErrorByPost.value, [postId]: '' }

  try {
    if (isSame) {
      const res = await fetch(apiURL(`/api/posts/${encodeURIComponent(postId)}/like`), { method: 'DELETE', credentials: 'include' })
      if (!res.ok) throw new Error()
    } else {
      const res = await fetch(apiURL(`/api/posts/${encodeURIComponent(postId)}/like`), {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ value }),
      })
      if (!res.ok) throw new Error()
    }
  } catch {
    groupPostLikes.value = { ...groupPostLikes.value, [postId]: prev }
    likeErrorByPost.value = { ...likeErrorByPost.value, [postId]: 'Could not update reaction.' }
  }
}

function sortComments(items: GroupPostComment[]) {
  return [...items].sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
}

function sortEvents(items: EventItem[]) {
  return [...items].sort((a, b) => new Date(a.eventTime).getTime() - new Date(b.eventTime).getTime())
}

function upsertGroupPost(post: Post) {
  if (post.groupId !== groupId.value) {
    return
  }

  groupPosts.value = sortGroupPosts([
    post,
    ...groupPosts.value.filter((item) => item.id !== post.id),
  ])

  if (!groupPostLikes.value[post.id]) {
    groupPostLikes.value = {
      ...groupPostLikes.value,
      [post.id]: { likeCount: post.likeCount ?? 0, dislikeCount: post.dislikeCount ?? 0, myReaction: post.myReaction ?? 0 },
    }
  }

  if (!commentsByPost.value[post.id]) {
    commentsByPost.value = {
      ...commentsByPost.value,
      [post.id]: [],
    }
  }
}

function removeGroupPost(postId: string) {
  groupPosts.value = groupPosts.value.filter((post) => post.id !== postId)

  const nextComments = { ...commentsByPost.value }
  delete nextComments[postId]
  commentsByPost.value = nextComments

  const nextLoading = { ...commentsLoadingByPost.value }
  delete nextLoading[postId]
  commentsLoadingByPost.value = nextLoading
}

function upsertGroupComment(comment: GroupPostComment) {
  const targetPost = groupPosts.value.find((post) => post.id === comment.postId)
  if (!targetPost) {
    return
  }

  commentsByPost.value = {
    ...commentsByPost.value,
    [comment.postId]: sortComments([
      comment,
      ...(commentsByPost.value[comment.postId] || []).filter((item) => item.id !== comment.id),
    ]),
  }
}

function removeGroupComment(postId: string, commentId: string) {
  commentsByPost.value = {
    ...commentsByPost.value,
    [postId]: (commentsByPost.value[postId] || []).filter((comment) => comment.id !== commentId),
  }
}

function upsertGroupEvent(event: EventItem) {
  if (event.groupId !== groupId.value) {
    return
  }

  events.value = sortEvents([
    event,
    ...events.value.filter((item) => item.id !== event.id),
  ])
}

function removeGroupEvent(eventId: string) {
  events.value = events.value.filter((event) => event.id !== eventId)
}

async function reloadGroupWorkspace() {
  if (groupReloadInFlight) {
    return
  }

  groupReloadInFlight = true
  try {
    await loadPage()
  } finally {
    groupReloadInFlight = false
  }
}

function isGroupSummaryPayload(payload: unknown): payload is GroupSummary {
  if (!payload || typeof payload !== 'object') return false
  const item = payload as Partial<GroupSummary>
  return typeof item.id === 'string' && typeof item.title === 'string' && typeof item.creatorId === 'string'
}

function isGroupPostPayload(payload: unknown): payload is Post {
  if (!payload || typeof payload !== 'object') return false
  const item = payload as Partial<Post>
  return typeof item.id === 'string' && typeof item.groupId === 'string' && typeof item.title === 'string'
}

function isGroupCommentPayload(payload: unknown): payload is GroupPostComment {
  if (!payload || typeof payload !== 'object') return false
  const item = payload as Partial<GroupPostComment> & { postId?: string }
  return typeof item.id === 'string' && typeof item.postId === 'string'
}

function isGroupEventPayload(payload: unknown): payload is EventItem {
  if (!payload || typeof payload !== 'object') return false
  const item = payload as Partial<EventItem>
  return typeof item.id === 'string' && typeof item.groupId === 'string' && typeof item.eventTime === 'string'
}

function isGroupStateRefreshPayload(payload: unknown): payload is GroupStateRefreshPayload {
  if (!payload || typeof payload !== 'object') return false
  return typeof (payload as Partial<GroupStateRefreshPayload>).groupId === 'string'
}

function isGroupPostDeletePayload(payload: unknown): payload is GroupPostDeletePayload {
  if (!payload || typeof payload !== 'object') return false
  const item = payload as Partial<GroupPostDeletePayload>
  return typeof item.groupId === 'string' && typeof item.postId === 'string'
}

function isGroupCommentDeletePayload(payload: unknown): payload is GroupCommentDeletePayload {
  if (!payload || typeof payload !== 'object') return false
  const item = payload as Partial<GroupCommentDeletePayload>
  return typeof item.groupId === 'string' && typeof item.postId === 'string' && typeof item.commentId === 'string'
}

function isGroupCalendarDeletePayload(payload: unknown): payload is GroupCalendarDeletePayload {
  if (!payload || typeof payload !== 'object') return false
  const item = payload as Partial<GroupCalendarDeletePayload>
  return typeof item.groupId === 'string' && typeof item.eventId === 'string'
}

function scheduleSocketReconnect() {
  if (socketClosedManually || reconnectTimerId !== null) {
    return
  }

  reconnectTimerId = window.setTimeout(() => {
    reconnectTimerId = null
    connectRealtimeSocket()
  }, 1500)
}

function handleRealtimeMessage(raw: { type?: string; payload?: unknown }) {
  switch (raw.type) {
    case 'group_summary_event':
      if (isGroupSummaryPayload(raw.payload) && raw.payload.id === groupId.value) {
        group.value = raw.payload
        editTitle.value = raw.payload.title
        editDescription.value = raw.payload.description
      }
      return
    case 'group_state_refresh':
      if (isGroupStateRefreshPayload(raw.payload) && raw.payload.groupId === groupId.value) {
        void reloadGroupWorkspace()
      }
      return
    case 'group_post_event':
      if (isGroupPostPayload(raw.payload)) {
        upsertGroupPost(raw.payload)
      }
      return
    case 'group_post_delete':
      if (isGroupPostDeletePayload(raw.payload) && raw.payload.groupId === groupId.value) {
        removeGroupPost(raw.payload.postId)
      }
      return
    case 'group_comment_event':
      if (isGroupCommentPayload(raw.payload)) {
        upsertGroupComment(raw.payload)
      }
      return
    case 'group_comment_delete':
      if (isGroupCommentDeletePayload(raw.payload) && raw.payload.groupId === groupId.value) {
        removeGroupComment(raw.payload.postId, raw.payload.commentId)
      }
      return
    case 'group_calendar_event':
      if (isGroupEventPayload(raw.payload)) {
        upsertGroupEvent(raw.payload)
      }
      return
    case 'group_calendar_delete':
      if (isGroupCalendarDeletePayload(raw.payload) && raw.payload.groupId === groupId.value) {
        removeGroupEvent(raw.payload.eventId)
      }
      return
  }
}

function connectRealtimeSocket() {
  if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
    return
  }

  socketClosedManually = false
  socket = new WebSocket(buildWebSocketURL('/ws'))

  socket.onmessage = (event) => {
    try {
      const raw = JSON.parse(event.data) as { type?: string; payload?: unknown }
      handleRealtimeMessage(raw)
    } catch {
      // Ignore invalid websocket payloads.
    }
  }

  socket.onclose = () => {
    socket = null
    scheduleSocketReconnect()
  }

  socket.onerror = () => {
    socket?.close()
  }
}

function disconnectRealtimeSocket() {
  socketClosedManually = true
  if (reconnectTimerId !== null) {
    window.clearTimeout(reconnectTimerId)
    reconnectTimerId = null
  }
  if (socket) {
    socket.close()
    socket = null
  }
}

function formatDate(value: string) {
  return new Date(value).toLocaleString()
}

function formatMemberRole(role?: string) {
  switch ((role || '').toLowerCase()) {
    case 'creator':
      return 'Creator'
    case 'admin':
      return 'Admin'
    case 'member':
      return 'Member'
    default:
      return 'Member'
  }
}

function isGroupCreatorPost(post: Post) {
  if (!group.value?.creatorId) {
    return false
  }

  return post.userId === group.value.creatorId || post.author.id === group.value.creatorId
}

function canManagePost(post: Post) {
  return post.userId === currentUserId.value
}

function canManageComment(comment: GroupPostComment) {
  return comment.author.user_id === currentUserId.value
}

function getCommentMenuKey(postId: string, commentId: string) {
  return `${postId}:${commentId}`
}

function formatUserDisplayName(user: {
  id: string
  nickname?: string
  firstName: string
  lastName: string
}) {
  if (user.nickname?.trim()) return `@${user.nickname.trim()}`
  const fullName = `${user.firstName || ''} ${user.lastName || ''}`.trim()
  return fullName || user.id
}

function friendlyInviteError(err: unknown, fallback: string) {
  if (!(err instanceof Error)) {
    return fallback
  }

  const message = err.message.toLowerCase()
  if (
    message.includes('pending group invite') ||
    message.includes('invite already pending') ||
    message.includes('unique constraint failed: group_invites.group_id, group_invites.receiver_id')
  ) {
    return 'This user already has an active invite for the group.'
  }

  if (message.includes('pending join request')) {
    return 'This user already sent a join request. Accept it from the join requests list instead.'
  }

  if (message.includes('only your followers can be invited')) {
    return 'You can only invite users who follow you.'
  }

  if (message.includes('user is already a group member')) {
    return 'This user is already a member of the group.'
  }

  return err.message
}

async function loadRequestDisplayNames(requests: GroupJoinRequest[]) {
  const uniqueUserIDs = [...new Set(requests.map((request) => request.userId).filter(Boolean))]

  if (uniqueUserIDs.length === 0) {
    requestDisplayNames.value = {}
    return
  }

  requestDisplayNames.value = await loadUserDisplayNames(uniqueUserIDs)
}

async function loadInviteDisplayNames(items: GroupInvite[]) {
  const uniqueUserIDs = [...new Set(items.map((invite) => invite.receiverId).filter(Boolean))]

  if (uniqueUserIDs.length === 0) {
    inviteDisplayNames.value = {}
    return
  }

  inviteDisplayNames.value = await loadUserDisplayNames(uniqueUserIDs)
}

async function loadUserDisplayNames(userIDs: string[]) {
  const entries = await Promise.allSettled(
    userIDs.map(async (userID) => {
      const response = await fetch(apiURL(`/api/users/${encodeURIComponent(userID)}`), {
        credentials: 'include',
      })
      if (!response.ok) {
        throw new Error(`Failed to load user ${userID}`)
      }

      const user = (await response.json()) as User
      return [userID, formatUserDisplayName(user)] as const
    }),
  )

  const names: Record<string, string> = {}
  entries.forEach((entry) => {
    if (entry.status === 'fulfilled') {
      const [userID, label] = entry.value
      names[userID] = label
    }
  })

  return names
}

async function searchInviteUsers(query: string) {
  inviteSearchLoading.value = true
  try {
    if (inviteFollowers.value.length === 0 && currentUserId.value) {
      inviteFollowers.value = await getUserFollowers(currentUserId.value)
    }

    const trimmed = query.trim().toLowerCase()
    inviteSuggestions.value = inviteFollowers.value.filter((user) => {
      if (members.value.some((member) => member.userId === user.id)) {
        return false
      }
      if (invites.value.some((invite) => invite.receiverId === user.id && invite.status === 'pending')) {
        return false
      }
      if (!trimmed) {
        return true
      }

      const nickname = (user.nickname || '').toLowerCase()
      const firstName = user.firstName.toLowerCase()
      const lastName = user.lastName.toLowerCase()
      const fullName = `${firstName} ${lastName}`.trim()

      return (
        user.id.toLowerCase().includes(trimmed) ||
        nickname.includes(trimmed) ||
        firstName.includes(trimmed) ||
        lastName.includes(trimmed) ||
        fullName.includes(trimmed)
      )
    })
  } catch {
    inviteSuggestions.value = []
  } finally {
    inviteSearchLoading.value = false
  }
}

function selectInviteUser(user: ConnectionUser) {
  suppressInviteSearchWatcher = true
  selectedInviteUserId.value = user.id
  inviteSearchTerm.value = formatUserDisplayName(user)
  showInviteSuggestions.value = false
}

function handleInviteInputFocus() {
  showInviteSuggestions.value = true
  void searchInviteUsers(inviteSearchTerm.value)
}

async function loadMemberInviteFollowers() {
  if (!currentUserId.value) {
    memberInviteFollowers.value = []
    return
  }

  memberInviteLoading.value = true
  memberInviteError.value = ''
  try {
    memberInviteFollowers.value = await getUserFollowers(currentUserId.value)
  } catch (err) {
    memberInviteFollowers.value = []
    memberInviteError.value = err instanceof Error ? err.message : 'Failed to load followers.'
  } finally {
    memberInviteLoading.value = false
  }
}

function openMemberInviteModal() {
  showMemberInviteModal.value = true
  memberInviteSearch.value = ''
  memberInviteError.value = ''
  selectedMemberInviteIds.value = []
  void loadMemberInviteFollowers()
}

function closeMemberInviteModal() {
  if (memberInviteSubmitting.value) {
    return
  }

  showMemberInviteModal.value = false
  memberInviteSearch.value = ''
  memberInviteFollowers.value = []
  memberInviteError.value = ''
  selectedMemberInviteIds.value = []
}

function toggleMemberInviteSelection(userId: string) {
  if (selectedMemberInviteIds.value.includes(userId)) {
    selectedMemberInviteIds.value = selectedMemberInviteIds.value.filter((id) => id !== userId)
    return
  }

  selectedMemberInviteIds.value = [...selectedMemberInviteIds.value, userId]
}

async function submitMemberInvites() {
  if (selectedMemberInviteIds.value.length === 0 || memberInviteSubmitting.value) {
    return
  }

  memberInviteSubmitting.value = true
  memberInviteError.value = ''

  try {
    await Promise.all(selectedMemberInviteIds.value.map((userId) => createGroupInvite(groupId.value, userId)))
    feedback.value =
      selectedMemberInviteIds.value.length === 1 ? 'Invite sent.' : 'Invites sent.'
    closeMemberInviteModal()
    await loadPage()
  } catch (err) {
    memberInviteError.value = friendlyInviteError(err, 'Failed to send invites.')
  } finally {
    memberInviteSubmitting.value = false
  }
}

async function fetchCommentsForPost(postId: string) {
  commentsLoadingByPost.value = {
    ...commentsLoadingByPost.value,
    [postId]: true,
  }

  try {
    const response = await fetch(`${API_BASE_URL}${API_ROUTES.POSTS}/${postId}/comments`, {
      method: 'GET',
      credentials: 'include',
    })
    if (!response.ok) {
      throw new Error('Failed to load comments')
    }

    const payload = (await response.json()) as GroupPostComment[]
    commentsByPost.value = {
      ...commentsByPost.value,
      [postId]: Array.isArray(payload) ? payload : [],
    }
  } catch {
    commentsByPost.value = {
      ...commentsByPost.value,
      [postId]: [],
    }
  } finally {
    commentsLoadingByPost.value = {
      ...commentsLoadingByPost.value,
      [postId]: false,
    }
  }
}

async function loadCommentsForPosts(posts: Post[]) {
  if (posts.length === 0) {
    commentsByPost.value = {}
    commentsLoadingByPost.value = {}
    return
  }

  await Promise.all(posts.map((post) => fetchCommentsForPost(post.id)))
}

async function loadPage() {
  loading.value = true
  error.value = ''
  feedback.value = ''

  try {
    group.value = await getGroup(groupId.value)
    editTitle.value = group.value.title
    editDescription.value = group.value.description
    const pending = await Promise.allSettled([
      listGroupMembers(groupId.value),
      listGroupInvites(groupId.value),
      group.value.isMember ? listEvents(groupId.value) : Promise.resolve([]),
      group.value.isMember ? listGroupPosts(groupId.value) : Promise.resolve([]),
      group.value.creatorId === currentUserId.value
        ? listJoinRequests(groupId.value)
        : Promise.resolve([]),
    ])

    if (pending[0].status === 'fulfilled') {
      members.value = Array.isArray(pending[0].value) ? pending[0].value : []
    } else {
      members.value = []
    }

    if (pending[1].status === 'fulfilled') {
      invites.value = Array.isArray(pending[1].value) ? pending[1].value : []
      await loadInviteDisplayNames(invites.value)
    } else {
      invites.value = []
      inviteDisplayNames.value = {}
    }

    if (pending[2].status === 'fulfilled') {
      events.value = Array.isArray(pending[2].value) ? pending[2].value : []
    } else {
      events.value = []
    }

    if (pending[3].status === 'fulfilled') {
      groupPosts.value = Array.isArray(pending[3].value) ? pending[3].value : []
      initLikeState(groupPosts.value)
      await loadCommentsForPosts(groupPosts.value)
    } else {
      groupPosts.value = []
      commentsByPost.value = {}
      commentsLoadingByPost.value = {}
    }

    if (pending[4].status === 'fulfilled') {
      joinRequests.value = Array.isArray(pending[4].value) ? pending[4].value : []
      await loadRequestDisplayNames(joinRequests.value)
    } else {
      joinRequests.value = []
      requestDisplayNames.value = {}
    }

    const firstFailed = pending.find((result) => result.status === 'rejected')
    if (firstFailed && firstFailed.status === 'rejected') {
      error.value = firstFailed.reason instanceof Error ? firstFailed.reason.message : 'Some group data failed to load.'
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load group.'
  } finally {
    loading.value = false
  }
}

async function handleUpdateGroup() {
  if (!group.value) {
    return
  }

  submittingGroup.value = true
  error.value = ''
  try {
    await updateGroup(groupId.value, {
      title: editTitle.value,
      description: editDescription.value,
    })
    feedback.value = 'Group updated.'
    await loadPage()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to update group.'
  } finally {
    submittingGroup.value = false
  }
}

async function handleDeleteGroup() {
  if (!group.value) {
    return
  }

  deletingGroup.value = true
  error.value = ''
  try {
    await deleteGroup(groupId.value)
    await router.push('/groups')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to delete group.'
  } finally {
    deletingGroup.value = false
  }
}

function toggleSettingsMenu() {
  showSettingsMenu.value = !showSettingsMenu.value
}

function selectCreatorPanel(panel: 'edit' | 'invite' | 'members' | 'requests' | 'events') {
  activeCreatorPanel.value = activeCreatorPanel.value === panel ? null : panel
}

function openDeleteConfirm() {
  deleteConfirmOpen.value = true
  showSettingsMenu.value = false
}

function closeSettingsMenu() {
  showSettingsMenu.value = false
}

function togglePostMenu(postId: string) {
  activePostMenuId.value = activePostMenuId.value === postId ? '' : postId
}

function closePostMenu() {
  activePostMenuId.value = ''
}

function toggleCommentMenu(postId: string, commentId: string) {
  const key = getCommentMenuKey(postId, commentId)
  activeCommentMenuKey.value = activeCommentMenuKey.value === key ? '' : key
}

function closeCommentMenu() {
  activeCommentMenuKey.value = ''
}

async function confirmDeleteGroup() {
  deleteConfirmOpen.value = false
  await handleDeleteGroup()
}

function openEditPostDialog(post: Post) {
  editPostTitle.value = post.title
  editPostBody.value = post.content
  editingPostId.value = post.id
  closePostMenu()
}

function closeEditPostDialog() {
  if (savingPostEdit.value) {
    return
  }

  editingPostId.value = ''
  editPostTitle.value = ''
  editPostBody.value = ''
}

function openEditCommentDialog(postId: string, comment: GroupPostComment) {
  editingCommentPostId.value = postId
  editingCommentId.value = comment.id
  editCommentContent.value = comment.content
  closeCommentMenu()
}

function closeEditCommentDialog() {
  if (savingCommentEdit.value) {
    return
  }

  editingCommentPostId.value = ''
  editingCommentId.value = ''
  editCommentContent.value = ''
}

function openDeletePostConfirm(postId: string) {
  postDeleteConfirmId.value = postId
  closePostMenu()
}

function closeDeletePostConfirm() {
  if (deletingPostId.value) {
    return
  }

  postDeleteConfirmId.value = ''
}

function openDeleteCommentConfirm(postId: string, commentId: string) {
  commentDeleteTarget.value = { postId, commentId }
  closeCommentMenu()
}

function closeDeleteCommentConfirm() {
  if (deletingCommentId.value) {
    return
  }

  commentDeleteTarget.value = null
}

async function handleSavePostEdit() {
  if (!editingPostId.value || savingPostEdit.value) {
    return
  }

  savingPostEdit.value = true
  error.value = ''

  try {
    const updatedPost = await updateGroupPost(editingPostId.value, {
      title: editPostTitle.value,
      content: editPostBody.value,
    })

    groupPosts.value = groupPosts.value.map((post) => (post.id === updatedPost.id ? updatedPost : post))
    feedback.value = 'Post updated.'
    closeEditPostDialog()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to update post.'
  } finally {
    savingPostEdit.value = false
  }
}

async function handleSaveCommentEdit() {
  if (!editingCommentId.value || !editingCommentPostId.value || savingCommentEdit.value) {
    return
  }

  savingCommentEdit.value = true
  commentErrorByPost.value = {
    ...commentErrorByPost.value,
    [editingCommentPostId.value]: '',
  }

  try {
    const updatedComment = (await updateComment(editingCommentId.value, {
      content: editCommentContent.value,
    })) as GroupPostComment

    commentsByPost.value = {
      ...commentsByPost.value,
      [editingCommentPostId.value]: (commentsByPost.value[editingCommentPostId.value] || []).map((comment) =>
        comment.id === updatedComment.id ? updatedComment : comment,
      ),
    }

    closeEditCommentDialog()
  } catch (err) {
    commentErrorByPost.value = {
      ...commentErrorByPost.value,
      [editingCommentPostId.value]: err instanceof Error ? err.message : 'Failed to update comment.',
    }
  } finally {
    savingCommentEdit.value = false
  }
}

async function confirmDeletePost() {
  if (!postDeleteConfirmId.value || deletingPostId.value) {
    return
  }

  const postId = postDeleteConfirmId.value
  deletingPostId.value = postId
  error.value = ''

  try {
    await deleteGroupPost(postId)
    groupPosts.value = groupPosts.value.filter((post) => post.id !== postId)
    const nextComments = { ...commentsByPost.value }
    delete nextComments[postId]
    commentsByPost.value = nextComments
    const nextCommentLoading = { ...commentsLoadingByPost.value }
    delete nextCommentLoading[postId]
    commentsLoadingByPost.value = nextCommentLoading
    feedback.value = 'Post deleted.'
    postDeleteConfirmId.value = ''
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to delete post.'
  } finally {
    deletingPostId.value = ''
  }
}

function handleDocumentClick(event: MouseEvent) {
  const target = event.target
  if (!(target instanceof Element)) {
    return
  }

  if (!target.closest('.post-menu-wrap')) {
    closePostMenu()
  }

  if (!target.closest('.comment-menu-wrap')) {
    closeCommentMenu()
  }
}

async function handleCreateInvite() {
  if (!selectedInviteUserId.value || submittingInvite.value) {
    return
  }

  submittingInvite.value = true
  try {
    await createGroupInvite(groupId.value, selectedInviteUserId.value)
    inviteSearchTerm.value = ''
    selectedInviteUserId.value = ''
    inviteSuggestions.value = []
    showInviteSuggestions.value = false
    feedback.value = 'Invite created.'
    await loadPage()
  } catch (err) {
    error.value = friendlyInviteError(err, 'Failed to create invite.')
  } finally {
    submittingInvite.value = false
  }
}

async function handleJoinRequest() {
  try {
    await createJoinRequest(groupId.value)
    feedback.value = 'Join request sent.'
    await loadPage()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to create request.'
  }
}

function handlePostImageChange(event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.item(0) ?? null

  if (!file) {
    postImage.value = null
    return
  }

  if (!['image/jpeg', 'image/png', 'image/gif'].includes(file.type)) {
    error.value = 'Invalid File Type [only .jpg, .png, .gif are allowed]'
    postImage.value = null
    target.value = ''
    return
  }

  if (file.size > 10 * 1024 * 1024) {
    error.value = 'Image must be 10MB or smaller.'
    postImage.value = null
    target.value = ''
    return
  }

  postImage.value = file
}

async function handleCreateGroupPost() {
  if (!group.value?.isMember || submittingPost.value) {
    return
  }

  submittingPost.value = true
  error.value = ''

  try {
    await createGroupPost(groupId.value, {
      title: postTitle.value,
      body: postBody.value,
      image: postImage.value,
    })
    postTitle.value = ''
    postBody.value = ''
    postImage.value = null
    feedback.value = 'Group post published.'
    await loadPage()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to publish group post.'
  } finally {
    submittingPost.value = false
  }
}

function handleCommentImageChange(postId: string, event: Event) {
  const target = event.target as HTMLInputElement
  const file = target.files?.item(0) ?? null

  if (!file) {
    commentImageByPost.value = { ...commentImageByPost.value, [postId]: null }
    return
  }

  if (!['image/jpeg', 'image/png', 'image/gif'].includes(file.type)) {
    commentErrorByPost.value = {
      ...commentErrorByPost.value,
      [postId]: 'Invalid file type. Only JPEG, PNG, and GIF allowed.',
    }
    commentImageByPost.value = { ...commentImageByPost.value, [postId]: null }
    target.value = ''
    return
  }

  if (file.size > 10 * 1024 * 1024) {
    commentErrorByPost.value = {
      ...commentErrorByPost.value,
      [postId]: 'File size must be under 10MB.',
    }
    commentImageByPost.value = { ...commentImageByPost.value, [postId]: null }
    target.value = ''
    return
  }

  commentErrorByPost.value = {
    ...commentErrorByPost.value,
    [postId]: '',
  }
  commentImageByPost.value = { ...commentImageByPost.value, [postId]: file }
}

async function submitGroupPostComment(postId: string) {
  const content = (commentDraftByPost.value[postId] || '').trim()
  const image = commentImageByPost.value[postId]
  if (!content && !image) {
    return
  }

  submittingCommentPostId.value = postId
  commentErrorByPost.value = {
    ...commentErrorByPost.value,
    [postId]: '',
  }

  try {
    const body = new FormData()
    body.append('content', content)
    if (image) {
      body.append('image', image)
    }

    const response = await fetch(`${API_BASE_URL}${API_ROUTES.POSTS}/${postId}/comments`, {
      method: 'POST',
      credentials: 'include',
      body,
    })

    if (!response.ok) {
      const text = await response.text()
      throw new Error(text || 'Failed to post comment.')
    }

    commentDraftByPost.value = { ...commentDraftByPost.value, [postId]: '' }
    commentImageByPost.value = { ...commentImageByPost.value, [postId]: null }
    await fetchCommentsForPost(postId)
  } catch (err) {
    commentErrorByPost.value = {
      ...commentErrorByPost.value,
      [postId]: err instanceof Error ? err.message : 'Failed to post comment.',
    }
  } finally {
    submittingCommentPostId.value = ''
  }
}

async function handleDeleteGroupPostComment(postId: string, commentId: string) {
  try {
    await deleteComment(commentId)

    await fetchCommentsForPost(postId)
  } catch (err) {
    commentErrorByPost.value = {
      ...commentErrorByPost.value,
      [postId]: err instanceof Error ? err.message : 'Failed to delete comment.',
    }
  }
}

async function confirmDeleteComment() {
  if (!commentDeleteTarget.value || deletingCommentId.value) {
    return
  }

  const { postId, commentId } = commentDeleteTarget.value
  deletingCommentId.value = commentId

  try {
    await handleDeleteGroupPostComment(postId, commentId)
    commentDeleteTarget.value = null
  } finally {
    deletingCommentId.value = ''
  }
}

async function handleInviteResponse(inviteId: string, status: 'accepted' | 'declined') {
  try {
    await respondToInvite(inviteId, status)
    feedback.value = `Invite ${status}.`
    await loadPage()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to update invite.'
  }
}

function handleLeaveGroup() {
  if (!group.value || !currentUserId.value) return
  leaveConfirmMessage.value = isCreator.value
    ? 'Leave this group? Ownership will transfer to another member if one is available.'
    : 'Leave this group?'
  leaveConfirmOpen.value = true
}

async function confirmLeaveGroup() {
  leaveConfirmOpen.value = false
  if (!group.value || !currentUserId.value) return
  error.value = ''
  feedback.value = ''
  try {
    await removeGroupMember(group.value.id, currentUserId.value)
    await router.push('/groups')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to leave group.'
  }
}

async function handleJoinRequestResponse(
  requestId: string,
  status: 'accepted' | 'declined',
) {
  try {
    await respondToJoinRequest(requestId, status)
    feedback.value = `Join request ${status}.`
    await loadPage()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to update request.'
  }
}

watch(
  () => route.params.groupId,
  async () => {
    await loadPage()
  },
)

watch(inviteSearchTerm, async (value) => {
  if (suppressInviteSearchWatcher) {
    suppressInviteSearchWatcher = false
    return
  }
  selectedInviteUserId.value = ''
  showInviteSuggestions.value = true
  await searchInviteUsers(value)
})

onMounted(async () => {
  document.addEventListener('click', handleDocumentClick)
  connectRealtimeSocket()
  const sessionData: SessionData | null = await fetchSessionData()
  currentUserId.value = sessionData?.user.id ?? ''
  await loadPage()
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleDocumentClick)
  disconnectRealtimeSocket()
})
</script>

<style scoped>
.page {
  width: min(100%, 1320px);
  margin: 0 auto;
  padding: 0.25rem 0 2rem;
  display: grid;
  gap: 1.25rem;
}

.header-card,
.summary-card,
.panel,
.message {
  border-radius: 2rem;
  border: 1px solid rgba(191, 219, 254, 0.78);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(239, 246, 255, 0.92));
  box-shadow: 0 26px 60px rgba(148, 163, 184, 0.22);
}

.header-card {
  position: relative;
  overflow: hidden;
  display: grid;
  gap: 1rem;
  align-items: start;
  padding: 1.75rem 1.95rem;
  background:
    radial-gradient(circle at top right, rgba(96, 165, 250, 0.3), transparent 30%),
    radial-gradient(circle at bottom left, rgba(147, 197, 253, 0.24), transparent 26%),
    linear-gradient(135deg, #17306b 0%, #1f4eb6 100%);
  border-color: rgba(96, 165, 250, 0.38);
  box-shadow: 0 34px 70px rgba(37, 99, 235, 0.28);
}

.header-card::after {
  content: "";
  position: absolute;
  top: -3rem;
  right: -2rem;
  width: 15rem;
  height: 15rem;
  border-radius: 999px;
  background: rgba(147, 197, 253, 0.14);
}

.header-content {
  flex: 1;
  min-width: 0;
  position: relative;
  z-index: 1;
}

.title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  gap: 0.6rem;
}

.settings-wrap {
  position: relative;
}

.settings-trigger {
  width: 3rem;
  height: 3rem;
  padding: 0;
  display: grid;
  place-items: center;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
  border: 1px solid rgba(191, 219, 254, 0.48);
  backdrop-filter: blur(12px);
  transition: background-color 0.2s ease, border-color 0.2s ease, transform 0.2s ease;
}

.settings-trigger :deep(svg) {
  width: 18px;
  height: 18px;
  stroke-width: 2.1;
  stroke: currentColor;
}

.back-link,
.panel-header a {
  display: inline-flex;
  align-items: center;
  width: fit-content;
  border-radius: 999px;
  background: rgba(219, 234, 254, 0.2);
  border: 1px solid rgba(191, 219, 254, 0.18);
  color: #eff6ff;
  font-weight: 800;
  font-size: 0.79rem;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  text-decoration: none;
  backdrop-filter: blur(12px);
}

.eyebrow {
  margin-top: 1rem;
  color: var(--primary-blue);
  font-size: 0.8rem;
  font-weight: 700;
  text-transform: uppercase;
}

h1,
h2 {
  color: var(--text-primary);
}

h1 {
  margin: 0;
  margin-top: 0.6em;
  font-size: clamp(1rem, 3vw, 3rem);
  font-weight: 800;
  line-height: 1.1;
  letter-spacing: -0.04em;
  color: #fff;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-word;
  min-width: 0;
}


h2 {
  margin: 0;
  font-size: 1.45rem;
  font-weight: 800;
  letter-spacing: -0.03em;
}

.description,
.small-copy {
  color: var(--text-secondary);
}

.description {
  max-width: 50rem;
  margin: 0;
  font-size: clamp(0.68rem, 1vw, 1rem);
  line-height: 1.65;
  color: rgba(239, 246, 255, 0.86);
  display: -webkit-box;
  -webkit-line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-top: 1em;
  text-overflow: ellipsis;
  word-break: break-word;
}
.summary-grid,
.moderation-grid,
.content-grid {
  display: grid;
  gap: 1rem;
}

.summary-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.summary-card {
  padding: 1.1rem 1.2rem;
  display: grid;
  gap: 0.25rem;
  border-radius: 1.5rem;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(240, 247, 255, 0.92));
}

.summary-card strong {
  font-size: 1.55rem;
  font-weight: 800;
  color: #0f172a;
}

.label {
  color: #64748b;
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.action-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.action-row .secondary-link:only-child {
  grid-column: 1 / -1;
  justify-content: center;
  text-align: center;
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: #fff;
  border-color: transparent;
}

.action-row > * {
  width: 100%;
  justify-content: center;
  text-align: center;
}

.panel {
  padding: 1.5rem;
  display: grid;
  gap: 1rem;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.985), rgba(241, 247, 255, 0.94));
}

.moderation-panel {
  border-color: rgba(96, 165, 250, 0.28);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(235, 245, 255, 0.94));
}

.content-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.panel.wide {
  grid-column: 1 / -1;
}

.group-posts-panel {
  gap: 1rem;
}

.group-post-form {
  display: grid;
  gap: 0.85rem;
  padding: 1.1rem;
  border-radius: 1.5rem;
  border: 1px solid rgba(191, 219, 254, 0.68);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(239, 246, 255, 0.92));
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.88);
}

.group-post-form textarea {
  border: 1px solid rgba(203, 213, 225, 0.95);
  border-radius: 1.3rem;
  padding: 1rem 1.1rem;
  font: inherit;
  resize: vertical;
  min-height: 9rem;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: inset 0 1px 2px rgba(148, 163, 184, 0.08);
}

.group-post-actions {
  display: flex;
  gap: 0.6rem;
  align-items: center;
  flex-wrap: wrap;
}

.char-counter {
  font-size: 0.76rem;
  color: var(--text-secondary);
  text-align: right;
  margin-top: -0.4rem;
}

.char-counter.char-warn {
  color: #d97706;
}

.char-counter.char-limit {
  color: #dc2626;
  font-weight: 700;
}

.upload-trigger {
  cursor: pointer;
}

.hidden-file-input {
  display: none;
}


.post-menu-wrap {
  position: relative;
}

.post-menu-trigger {
  width: 2rem;
  height: 2rem;
  padding: 0;
  display: grid;
  place-items: center;
  border-radius: 999px;
}

.post-menu-trigger :deep(svg) {
  width: 1rem;
  height: 1rem;
}

.post-menu-popover {
  position: absolute;
  top: calc(100% + 0.35rem);
  right: 0;
  min-width: 9rem;
  display: grid;
  gap: 0.2rem;
  padding: 0.35rem;
  border: 1px solid var(--border-color);
  border-radius: 0.85rem;
  background: #fff;
  box-shadow: 0 12px 30px rgba(15, 23, 42, 0.16);
  z-index: 5;
}

.post-menu-item {
  width: 100%;
  display: flex;
  justify-content: flex-start;
  background: transparent;
  color: var(--text-primary);
  border: 1px solid transparent;
  border-radius: 0.65rem;
  padding: 0.55rem 0.7rem;
}

.post-menu-item:hover,
.post-menu-item:focus-visible {
  background: #f8fafc;
  border-color: #e2e8f0;
}

.danger-text {
  color: var(--status-error);
}


.comment-menu-wrap {
  position: relative;
}

.comment-menu-trigger {
  width: 1.55rem;
  height: 1.55rem;
  padding: 0;
  display: grid;
  place-items: center;
  border-radius: 999px;
  border: none;
  background: transparent;
}

.comment-menu-trigger :deep(svg) {
  width: 0.85rem;
  height: 0.85rem;
}

.comment-menu-popover {
  position: absolute;
  top: calc(100% + 0.25rem);
  right: 0;
  min-width: 7rem;
  display: grid;
  gap: 0.15rem;
  padding: 0.25rem;
  border: 1px solid var(--border-color);
  border-radius: 0.75rem;
  background: #fff;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.14);
  z-index: 5;
}

.comment-menu-item {
  width: 100%;
  display: flex;
  justify-content: flex-start;
  background: transparent;
  color: var(--text-primary);
  border: 1px solid transparent;
  border-radius: 0.55rem;
  padding: 0.42rem 0.55rem;
  font-size: 0.82rem;
}

.comment-menu-item:hover,
.comment-menu-item:focus-visible {
  background: #f8fafc;
  border-color: #e2e8f0;
}

.edit-post-textarea {
  border: 1px solid var(--border-color);
  border-radius: 1rem;
  padding: 0.8rem 0.95rem;
  font: inherit;
  resize: vertical;
  min-height: 8rem;
}


.compact-message {
  padding: 0.7rem 0.85rem;
}

.member-invite-overlay {
  position: fixed;
  inset: 0;
  z-index: 55;
  display: grid;
  place-items: center;
  padding: 1rem;
  background: rgba(15, 23, 42, 0.38);
}

.member-invite-popover {
  width: min(520px, 100%);
  max-height: min(70vh, 560px);
  overflow: auto;
  display: grid;
  gap: 1rem;
  padding: 1.15rem;
  border: 1px solid rgba(191, 219, 254, 0.72);
  border-radius: 1.5rem;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.985), rgba(239, 246, 255, 0.96));
  box-shadow: 0 30px 70px rgba(15, 23, 42, 0.24);
}

.member-invite-header {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  align-items: start;
}

.member-invite-header h3 {
  color: var(--text-primary);
}

.member-invite-header p {
  color: var(--text-secondary);
  font-size: 0.88rem;
}

.member-invite-toolbar {
  display: flex;
  gap: 0.65rem;
  align-items: center;
}

.member-invite-toolbar input {
  flex: 1;
  min-width: 0;
}

.member-invite-list {
  list-style: none;
  display: grid;
  gap: 0.55rem;
  padding: 0;
  margin: 0;
  max-height: 26rem;
  overflow-y: auto;
  border: 1px solid #e2e8f0;
  border-radius: 0.9rem;
  background: #fff;
  padding: 0.45rem;
}

.member-invite-item {
  border: 1px solid #e2e8f0;
  border-radius: 0.85rem;
  background: #f8fafc;
}

.member-invite-option {
  display: flex;
  gap: 0.75rem;
  align-items: flex-start;
  padding: 0.75rem 0.85rem;
  cursor: pointer;
}

.member-invite-option input {
  margin-top: 0.2rem;
  flex: 0 0 auto;
  min-width: auto;
}

.member-invite-copy {
  display: grid;
  gap: 0.15rem;
}

.member-invite-copy strong {
  color: var(--text-primary);
}

.member-invite-copy small {
  color: var(--text-secondary);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  align-items: center;
  flex-wrap: wrap;
}

.item-list {
  list-style: none;
  display: grid;
  gap: 0.75rem;
  padding: 0;
}

.item-list li {
  display: flex;
  justify-content: space-between;
  gap: 0.75rem;
  align-items: center;
  padding: 0.95rem 1rem;
  border-radius: 1.15rem;
  border: 1px solid rgba(191, 219, 254, 0.52);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(241, 247, 255, 0.92));
}

.request-list {
  list-style: none;
  display: grid;
  gap: 0.75rem;
  padding: 0;
}

.request-item {
  display: flex;
  justify-content: space-between;
  gap: 0.9rem;
  align-items: center;
  padding: 1rem 1.05rem;
  border-radius: 1.2rem;
  border: 1px solid rgba(96, 165, 250, 0.34);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.97), rgba(239, 246, 255, 0.95));
}

.request-meta {
  display: grid;
  gap: 0.2rem;
}

.request-counter {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.3rem 0.7rem;
  background: #dbeafe;
  color: #1d4ed8;
  font-size: 0.78rem;
  font-weight: 800;
}

.empty-copy {
  border-radius: 1.3rem;
  border: 1px dashed rgba(148, 163, 184, 0.5);
  padding: 1rem 1.1rem;
  color: #64748b;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(248, 250, 252, 0.96));
}

.inline-form,
.queue-actions {
  display: flex;
  gap: 0.6rem;
  flex-wrap: wrap;
}

.invite-search {
  display: grid;
  gap: 0.75rem;
}

.invite-suggestions {
  list-style: none;
  display: grid;
  gap: 0.5rem;
  padding: 0;
  max-height: 26rem;
  overflow-y: auto;
  border: 1px solid var(--border-color);
  border-radius: 0.9rem;
  background: #fff;
  padding: 0.45rem;
}

.invite-suggestion {
  width: 100%;
  display: grid;
  gap: 0.15rem;
  text-align: left;
  border: 1px solid var(--border-color);
  border-radius: 0.85rem;
  background: #fff;
  color: var(--text-primary);
  padding: 0.75rem 0.9rem;
}

.invite-suggestion strong {
  font-size: 0.95rem;
}

input {
  flex: 1;
  min-width: 10rem;
  border: 1px solid rgba(203, 213, 225, 0.95);
  border-radius: 1.1rem;
  padding: 0.92rem 1rem;
  font: inherit;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: inset 0 1px 2px rgba(148, 163, 184, 0.08);
}

button,
.secondary-link {
  border: 1px solid transparent;
  border-radius: 999px;
  padding: 0.82rem 1.12rem;
  font: inherit;
  font-weight: 700;
  text-decoration: none;
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease, background-color 0.2s ease, border-color 0.2s ease;
}

button {
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
  color: var(--text-white);
  box-shadow: 0 16px 32px rgba(37, 99, 235, 0.2);
}

button.secondary,
.secondary-link {
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(239, 246, 255, 0.94));
  color: #0f172a;
  border-color: rgba(191, 219, 254, 0.82);
  box-shadow: 0 14px 28px rgba(148, 163, 184, 0.16);
}

button.settings-trigger.secondary {
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
  box-shadow: none;
}

button.settings-trigger.secondary:hover,
button.settings-trigger.secondary:focus-visible {
  background: rgba(255, 255, 255, 0.26);
  color: #fff;
  border-color: rgba(255, 255, 255, 0.7);
  transform: translateY(-1px);
}

button.danger {
  background: rgba(239, 68, 68, 0.12);
  color: #dc2626;
  border-color: rgba(248, 113, 113, 0.28);
  box-shadow: 0 16px 32px rgba(248, 113, 113, 0.18);
}

button.compact {
  padding: 0.45rem 0.7rem;
}

.message {
  padding: 1rem 1.1rem;
  color: #475569;
}

.message.error {
  color: var(--status-error);
}

.message.success {
  color: var(--status-success);
}

.settings-overlay {
  position: fixed;
  inset: 0;
  background: rgba(17, 24, 39, 0.45);
  display: grid;
  place-items: center;
  z-index: 40;
  padding: 1rem;
}

.settings-popover {
  width: min(620px, 100%);
  max-height: calc(100vh - 2rem);
  overflow: auto;
  display: grid;
  gap: 0.9rem;
  padding: 1rem;
  border: 1px solid rgba(191, 219, 254, 0.72);
  border-radius: 1.5rem;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.985), rgba(239, 246, 255, 0.96));
  box-shadow: 0 30px 70px rgba(15, 23, 42, 0.24);
}

.settings-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.settings-header h3 {
  color: var(--text-primary);
}

.settings-close {
  width: 32px;
  height: 32px;
  padding: 0;
  display: grid;
  place-items: center;
}

.settings-close :deep(svg) {
  width: 16px;
  height: 16px;
}

.settings-nav {
  display: grid;
  gap: 0.45rem;
}

.accordion-item {
  display: grid;
  gap: 0.45rem;
}

.nav-action {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  border-radius: 0.75rem;
  padding: 0.55rem 0.72rem;
}

.nav-action :deep(svg) {
  width: 14px;
  height: 14px;
}

.delete-nav {
  color: var(--status-error);
}

.accordion-content {
  border: 1px solid rgba(191, 219, 254, 0.64);
  border-radius: 1.1rem;
  padding: 0.95rem;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(241, 247, 255, 0.92));
  display: grid;
  gap: 0.75rem;
}

.confirm-overlay {
  position: fixed;
  inset: 0;
  background: rgba(17, 24, 39, 0.45);
  display: grid;
  place-items: center;
  z-index: 50;
  padding: 1rem;
}

.confirm-dialog {
  width: min(420px, 100%);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.985), rgba(239, 246, 255, 0.96));
  border: 1px solid rgba(191, 219, 254, 0.72);
  border-radius: 1.5rem;
  padding: 1.2rem;
  display: grid;
  gap: 0.9rem;
  box-shadow: 0 30px 70px rgba(15, 23, 42, 0.24);
}

.confirm-dialog h3 {
  color: var(--text-primary);
}

.confirm-dialog p {
  color: var(--text-secondary);
}

.confirm-actions {
  display: flex;
  justify-content: end;
  gap: 0.65rem;
}

small {
  color: var(--text-secondary);
}

@media (max-width: 800px) {
  .summary-grid,
  .moderation-grid,
  .content-grid {
    grid-template-columns: 1fr;
  }

  .action-row {
    grid-template-columns: 1fr;
  }

  .header-card {
    flex-direction: column;
  }

  .request-item {
    flex-direction: column;
    align-items: flex-start;
  }
}

/* ── X-style group posts ── */
.gp-post-list {
  list-style: none;
  padding: 0;
  display: grid;
  gap: 1rem;
}

.gp-post {
  border: 1px solid rgba(191, 219, 254, 0.56);
  border-radius: 1.55rem;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.985), rgba(241, 247, 255, 0.94));
  padding: 1.2rem 1.15rem 0.95rem;
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.7rem;
  box-shadow: 0 20px 45px rgba(148, 163, 184, 0.16);
}

.gp-author-row {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
}

.gp-avatar-col {
  flex-shrink: 0;
}

.gp-avatar {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  object-fit: cover;
  display: block;
}

.gp-avatar-sm {
  width: 34px;
  height: 34px;
}

.gp-avatar-fallback {
  background: #2563eb;
  color: #fff;
  font-size: 0.8rem;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

.gp-author-meta {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem;
}

.gp-display-name {
  font-weight: 700;
  color: var(--text-primary);
  font-size: 0.95rem;
}

.gp-creator-badge {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.14rem 0.45rem;
  background: #dbeafe;
  color: #1d4ed8;
  font-size: 0.68rem;
  font-weight: 800;
  text-transform: uppercase;
  letter-spacing: 0.01em;
}

.gp-handle {
  color: var(--text-secondary);
  font-size: 0.85rem;
}

.gp-post-actions-wrap {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  flex-shrink: 0;
}

.gp-post-date {
  color: var(--text-secondary);
  font-size: 0.8rem;
  white-space: nowrap;
}

.gp-post-title {
  font-weight: 700;
  font-size: 1.05rem;
  color: var(--text-primary);
  overflow-wrap: break-word;
  margin: 0;
}

.gp-post-body {
  color: var(--text-primary);
  line-height: 1.65;
  white-space: pre-wrap;
  overflow-wrap: break-word;
}

.gp-post-image-wrap {
  border-radius: 1.2rem;
  overflow: hidden;
  border: 1px solid rgba(191, 219, 254, 0.62);
}

.gp-post-image {
  width: 100%;
  max-height: 400px;
  object-fit: cover;
  display: block;
}

.gp-divider {
  height: 1px;
  background: rgba(191, 219, 254, 0.72);
  margin: 0.2rem 0;
}

.gp-stats-row {
  display: flex;
  gap: 1.2rem;
  padding: 0.15rem 0;
}

.gp-stat {
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.gp-stat strong {
  color: var(--text-primary);
  font-weight: 700;
}

/* Composer */
.gp-composer {
  display: flex;
  gap: 0.75rem;
  align-items: flex-start;
}

.gp-composer-body {
  flex: 1;
  min-width: 0;
  display: grid;
  gap: 0.4rem;
}

.gp-composer-input {
  width: 100%;
  border: none;
  outline: none;
  resize: none;
  font: inherit;
  font-size: 1rem;
  color: var(--text-primary);
  background: transparent;
  padding: 0.3rem 0;
  border-bottom: 1px solid #e2e8f0;
  line-height: 1.5;
}

.gp-composer-input:focus {
  border-bottom-color: #2563eb;
}

.gp-composer-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding-top: 0.3rem;
}

.gp-attach-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  color: #2563eb;
  cursor: pointer;
}

.gp-file-name {
  font-size: 0.78rem;
  color: var(--text-secondary);
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.gp-reply-btn {
  padding: 0.3rem 1rem;
  border-radius: 999px;
  background: #2563eb;
  color: #fff;
  font-weight: 700;
  font-size: 0.88rem;
  border: none;
  cursor: pointer;
  transition: background 0.15s;
}

.gp-reply-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.gp-reply-btn:not(:disabled):hover {
  background: #1d4ed8;
}

.gp-form-error {
  color: var(--status-error, #dc2626);
  font-size: 0.83rem;
}

/* Comments */
.gp-loading,
.gp-empty-comments {
  color: var(--text-secondary);
  font-size: 0.9rem;
  padding: 0.4rem 0;
}

.gp-comment-row {
  display: flex;
  gap: 0.65rem;
  align-items: flex-start;
  padding: 0.8rem 0;
  border-top: 1px solid rgba(226, 232, 240, 0.88);
}

.gp-comment-left {
  flex-shrink: 0;
}

.gp-comment-content {
  flex: 1;
  min-width: 0;
  display: grid;
  gap: 0.2rem;
}

.gp-comment-header {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.3rem;
}

.gp-comment-name {
  font-weight: 700;
  color: var(--text-primary);
  font-size: 0.88rem;
}

.gp-comment-dot {
  color: var(--text-secondary);
  font-size: 0.85rem;
}

.gp-comment-time {
  color: var(--text-secondary);
  font-size: 0.82rem;
}

.gp-comment-text {
  color: var(--text-primary);
  font-size: 0.93rem;
  line-height: 1.5;
  white-space: pre-wrap;
  overflow-wrap: break-word;
  margin: 0;
}

.gp-comment-image {
  width: min(220px, 100%);
  border-radius: 0.95rem;
  border: 1px solid rgba(191, 219, 254, 0.62);
  margin-top: 0.3rem;
}

.gp-timestamp {
  color: #64748b;
  font-size: 0.82rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  padding: 0.1rem 0;
}

.gp-action-bar {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  flex-wrap: wrap;
  padding: 0.15rem 0;
}

.gp-action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background: rgba(239, 246, 255, 0.72);
  border: 1px solid rgba(191, 219, 254, 0.62);
  cursor: pointer;
  color: #475569;
  padding: 0.5rem 0.85rem;
  border-radius: 999px;
  box-shadow: none;
  transition: color 0.15s, background-color 0.15s, border-color 0.15s;
}

.gp-action-btn:hover {
  color: #0f172a;
  background: rgba(219, 234, 254, 0.9);
  border-color: rgba(96, 165, 250, 0.44);
}

.gp-action-icon {
  width: 20px;
  height: 20px;
}

.gp-action-count {
  font-size: 0.88rem;
}

.gp-like-btn.gp-active-like {
  color: #e0245e;
  background: rgba(251, 207, 232, 0.66);
  border-color: rgba(244, 114, 182, 0.45);
}

.gp-dislike-btn.gp-active-dislike {
  color: #2563eb;
  background: rgba(219, 234, 254, 0.92);
  border-color: rgba(96, 165, 250, 0.52);
}

.gp-like-error {
  font-size: 0.8rem;
  color: var(--status-error, #dc2626);
  margin-left: auto;
}

input:focus,
textarea:focus,
.group-post-form textarea:focus,
.member-invite-toolbar input:focus,
.invite-search input:focus,
.edit-post-textarea:focus {
  outline: none;
  border-color: rgba(59, 130, 246, 0.8);
  box-shadow: 0 0 0 4px rgba(191, 219, 254, 0.65);
}

button:hover,
.secondary-link:hover {
  transform: translateY(-1px);
}
</style>
