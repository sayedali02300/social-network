<template>
  <router-link to="/" class="nav-tab" :class="{'mobile-tab': isMobile}" active-class="active-link" @click="handleHomeClick">
    <HomeSolid v-if="route.path === '/'" class="nav-icon" />
    <HomeOutline v-else class="nav-icon" />
    <span class="nav-label">Home</span>
  </router-link>

  <router-link to="/search" class="nav-tab" :class="{'mobile-tab': isMobile}" active-class="active-link" @click="handleOtherClick">
    <SearchSolid v-if="route.path.startsWith('/search')" class="nav-icon" />
    <SearchOutline v-else class="nav-icon" />
    <span class="nav-label">Search</span>
  </router-link>

  <router-link to="/friends" class="nav-tab" :class="{'mobile-tab': isMobile}" active-class="active-link" @click="handleOtherClick">
    <UsersSolid v-if="route.path === '/friends'" class="nav-icon" />
    <UsersOutline v-else class="nav-icon" />
    <span class="nav-label">Friends</span>
  </router-link>

  <router-link to="/groups" class="nav-tab" :class="{'mobile-tab': isMobile}" active-class="active-link" @click="handleOtherClick">
    <GroupSolid v-if="route.path.startsWith('/groups')" class="nav-icon" />
    <GroupOutline v-else class="nav-icon" />
    <span class="nav-label">Groups</span>
  </router-link>

  <router-link to="/chats" class="nav-tab" :class="{'mobile-tab': isMobile}" active-class="active-link" @click="handleOtherClick">
    <ChatSolid v-if="route.path.startsWith('/chats')" class="nav-icon" />
    <ChatOutline v-else class="nav-icon" />
    <span class="nav-label">Chats</span>
  </router-link>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { 
  HomeIcon as HomeSolid, 
  MagnifyingGlassIcon as SearchSolid,
  UsersIcon as UsersSolid, 
  UserGroupIcon as GroupSolid,
  ChatBubbleLeftRightIcon as ChatSolid,
} from '@heroicons/vue/24/solid'
import { 
  HomeIcon as HomeOutline, 
  MagnifyingGlassIcon as SearchOutline,
  UsersIcon as UsersOutline, 
  UserGroupIcon as GroupOutline,
  ChatBubbleLeftRightIcon as ChatOutline,
} from '@heroicons/vue/24/outline'

const route = useRoute()

defineProps<{
  isMobile: boolean
}>()

const emit = defineEmits(['closeMenu', 'scrollToTop'])

const handleHomeClick = () => {
  emit('scrollToTop')
  emit('closeMenu')
}

const handleOtherClick = () => {
  emit('closeMenu')
}
</script>
