<template>
  <div class="shade" :class="{ active: modal !== Modal.None }" @click="switchModal(Modal.None)"></div>
  <div class="header">
    <div class="start">
      <div class="menu material-icons" @click="switchModal(Modal.Sidebar)">menu</div>
      <router-link :to="{ name: 'home' }"><ConfaLogo></ConfaLogo></router-link>
    </div>
    <div class="end">
      <!-- eslint-disable-next-line vue/no-v-html -->
      <div v-if="allowedWrite" class="avatar" @click="switchModal(Modal.Profile)" v-html="profile.avatar"></div>
      <router-link v-if="!allowedWrite" class="btn-convex login" to="/login">Sign in</router-link>
    </div>

    <div v-if="modal === Modal.Sidebar" class="sidebar">
      <router-link v-if="allowedWrite" class="control-item" to="/">
        <span class="icon material-icons">hub</span>
        My content
      </router-link>
      <router-link class="control-item" to="/">
        <span class="icon material-icons">explore</span>
        Explore
      </router-link>
      <div v-if="allowedWrite" class="control-divider"></div>
      <router-link v-if="allowedWrite" class="control-item" to="/new">
        <span class="icon material-icons">add</span>
        Create confa
      </router-link>
      <div class="control-divider"></div>
      <div class="control-item" @click="toggleTheme">
        <span class="icon material-icons">{{ theme === Theme.Dark ? "light_mode" : "dark_mode" }}</span>
        Switch theme
      </div>
    </div>

    <div v-if="allowedWrite" class="control" :class="{ opened: modal === Modal.Profile }">
      <router-link class="control-item" to="/">
        <span class="icon material-icons">person</span>
        My profile
      </router-link>
      <div class="control-divider"></div>
      <div class="control-item">
        <span class="icon material-icons">logout</span>
        Sign out
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from "vue"
import { userStore } from "@/api"
import { genAvatar, genName } from "@/platform/gen"
import { Theme } from "@/platform/theme"
import ConfaLogo from "@/components/ConfaLogo.vue"

interface Profile {
  avatar?: string
  name?: string
}

enum Modal {
  None = "",
  Sidebar = "sidebar",
  Profile = "profile",
}

const emit = defineEmits<{
  (e: "theme", value: Theme): void
}>()

const theme = ref(Theme.Light)
const modal = ref(Modal.None)

const allowedWrite = computed<boolean>(() => {
  return userStore.getState().allowedWrite
})
const profile = computed<Profile>(() => {
  return {
    avatar: genAvatar(userStore.getState().id, 35),
    name: genName(userStore.getState().id),
  }
})

watch(theme, (val: Theme) => {
  emit("theme", val)
  localStorage.setItem("theme", val)
})

onMounted(() => {
  theme.value = (localStorage.getItem("theme") as Theme) || Theme.Light
})

function switchModal(val: Modal) {
  if (modal.value === val) {
    modal.value = Modal.None
    return
  }
  modal.value = val
}
function toggleTheme() {
  theme.value = theme.value === Theme.Light ? Theme.Dark : Theme.Light
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

$height: 100%

.header
  @include theme.shadow-m

  position: relative
  top: 0
  display: flex
  flex-direction: row
  align-items: center
  justify-content: flex-start
  background: var(--color-background)
  width: 100%
  height: $height
  z-index: 100

.menu
  @include theme.clickable

.avatar
  @include theme.clickable

  width: 34px
  height: 34px
  border-radius: 50%
  overflow: hidden
  margin: 10px

.start
  display: flex
  flex-direction: row
  align-items: center
  justify-content: flex-start
  z-index: 60
  padding: 0 30px
  height: 100%
  width: 300px
  background: var(--color-background)

.logo
  font-size: 1.5rem
  margin: 20px
  cursor: pointer

.end
  display: flex
  flex-direction: row
  align-items: center
  justify-content: flex-end
  margin-left: auto
  padding: 0 30px
  width: 300px
  height: 100%
  z-index: 60
  background: var(--color-background)

.sidebar
  @include theme.shadow-m

  display: inline-block
  position: absolute
  top: $height
  left: 0
  text-align: left
  height: 100vh
  width: 200px
  overflow: hidden
  background: var(--color-background)

  border-radius: 0 0 4px 4px
  z-index: 50

.control
  @include theme.shadow-m

  position: absolute
  top: $height
  text-align: left
  right: 20px
  height: 0
  width: 200px
  overflow: hidden
  background: var(--color-background)

  border-radius: 0 0 4px 4px
  z-index: 50
  &.opened
    height: auto

.control-item
  @include theme.clickable

  display: flex
  flex-direction: row
  align-items: center
  justify-content: flex-start

  text-decoration: none
  width: 100%
  height: 40px
  line-height: 40px
  color: var(--color-font)
  background: var(--color-background)
  &:hover
    color: var(--color-highlight-font)
    background: var(--color-highlight-background)

  .icon
    margin: 10px
    font-size: 20px

.control-divider
  height: 0
  width: 100%
  margin: 5px 0
  border-top: 1px solid var(--color-outline)

.name
  @include theme.clickable

.login
  margin: 10px

.theme-switch
  @include theme.clickable

.shade
  position: fixed
  left: 0
  top: 0
  height: 100vh
  width: 100vw
  background-color: var(--color-background)
  transition: opacity 1s ease-in-out
  opacity: 0.5
  display: none
  z-index: 10
  &.active
    display: block
</style>