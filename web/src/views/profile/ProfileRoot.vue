<template>
  <PageLoader v-if="loading" />

  <div v-if="!loading && profile" class="content">
    <div class="title">{{ profile.givenName || genName(profile.ownerId) }} {{ profile.familyName }}</div>
    <div class="path">
      <router-link class="path-link" :to="route.profile(handle, tab)">{{ profile.handle }}</router-link>
    </div>
    <div class="header">
      <router-link :to="route.profile(handle, 'overview')" class="header-item" :class="{ active: tab === 'overview' }">
        <span class="material-icons icon">remove_red_eye</span>
        Overview
      </router-link>
      <router-link
        v-if="profile.ownerId === accessStore.state.id"
        :to="route.profile(handle, 'edit')"
        class="header-item"
        :class="{ active: tab === 'edit' }"
      >
        <span class="material-icons icon">edit</span>
        Edit
      </router-link>
      <router-link
        v-if="profile.ownerId === accessStore.state.id"
        :to="route.profile(handle, 'settings')"
        class="header-item"
        :class="{ active: tab === 'settings' }"
      >
        <span class="material-icons icon">settings</span>
        Settings
      </router-link>
    </div>
    <div class="header-divider"></div>
    <div class="tab">
      <ProfileOverview v-if="tab === 'overview'" :avatar="avatar" :profile="profile" />
      <ProfileEdit v-if="tab === 'edit'" :profile="profile" :avatar="avatar" @update="update" @avatar="updateAvatar" />
      <ProfileSettings v-if="tab === 'settings'" />
    </div>
  </div>

  <NotFound v-if="!loading && !profile" />
</template>

<script setup lang="ts">
import { ref, watch } from "vue"
import { useRouter } from "vue-router"
import { api, errorCode, Code } from "@/api"
import { ProfileClient } from "@/api/profile"
import { accessStore } from "@/api/models/access"
import { Profile, profileStore } from "@/api/models/profile"
import { route, ProfileTab, handleNew } from "@/router"
import { genAvatar, genName } from "@/platform/gen"
import PageLoader from "@/components/PageLoader.vue"
import NotFound from "@/views/NotFound.vue"
import ProfileEdit from "./ProfileEdit.vue"
import ProfileOverview from "./ProfileOverview.vue"
import ProfileSettings from "./ProfileSettings.vue"
import { notificationStore } from "@/api/models/notifications"

const props = defineProps<{
  tab: ProfileTab
  handle: string
}>()

const router = useRouter()

const profile = ref<Profile | null>()
const loading = ref(true)
const avatar = ref<string>("")

watch(
  [() => accessStore.state.id, () => props.handle],
  async () => {
    if (accessStore.state.id === "") {
      return
    }

    if (!accessStore.state.allowedWrite && (props.tab == "edit" || props.handle === handleNew)) {
      router.replace(route.login())
      return
    }

    if (profile.value && props.handle === profile.value.handle) {
      return
    }

    loading.value = true
    try {
      if (props.handle === handleNew) {
        profile.value = await new ProfileClient(api).update()
        router.replace(route.profile(profile.value.handle, props.tab))
      } else {
        profile.value = await new ProfileClient(api).fetchOne({
          handle: props.handle,
        })
      }

      if (profile.value.avatarUrl) {
        avatar.value = await new ProfileClient(api).fetchAvatar(profile.value.avatarUrl)
      }
      if (!avatar.value) {
        avatar.value = await genAvatar(profile.value.ownerId, 460)
      }
    } catch (e) {
      switch (errorCode(e)) {
        case Code.NotFound:
          break
        default:
          notificationStore.error("failed to load profile")
          break
      }
    } finally {
      loading.value = false
    }
  },
  { immediate: true },
)

function update(value: Profile) {
  profile.value = value
  profileStore.set(value)
  router.replace(route.profile(value.handle, props.tab))
}

function updateAvatar(full: string, thumbnail: string) {
  avatar.value = full
  profileStore.update("", "", thumbnail)
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.content
  width: 100%
  min-height: 100%
  display: flex
  flex-direction: column
  justify-content: flex-start
  align-items: center

.title
  cursor: default
  font-size: 1.5em
  margin-top: 40px
  width: 100%
  max-width: theme.$content-width
  text-align: left
  padding: 0 30px

.path
  width: 100%
  text-align: left
  max-width: theme.$content-width
  padding: 0 30px
  margin-bottom: 10px
  font-size: 12px

.path-link
  text-decoration: none
  color: var(--color-font-disabled)
  &:hover
    color: var(--color-font)
    text-decoration: underline

.header
  width: 100%
  max-width: theme.$content-width
  display: flex
  flex-direction: row
  margin-bottom: -1px
  padding: 0 20px

.header-item
  @include theme.clickable

  display: flex
  flex-direction: row
  align-items: center
  justify-content: center
  text-decoration: none
  color: var(--color-font)
  padding: 10px
  width: 150px
  border-bottom: 1px solid transparent
  transition: border 0.3s
  &.active
    border-bottom: 1px solid var(--color-highlight-background)
  &:hover:not(.active)
    border-bottom: 1px solid var(--color-font)

  .icon
    margin-right: 5px
    font-size: 15px

.header-divider
  width: 100%
  height: 0
  border-bottom: 1px solid var(--color-outline)

.tab
  width: 100%
  max-width: theme.$content-width
  flex: 1
</style>
