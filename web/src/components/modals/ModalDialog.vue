<template>
  <div v-if="isVisible" class="background"></div>
  <div class="container">
    <Transition name="fade">
      <div v-if="isVisible" class="content">
        <div class="wrapper">
          <slot></slot>
        </div>
        <table v-if="buttons">
          <tr>
            <td
              v-for="btn in buttons"
              :key="btn.id"
              class="cell"
              :class="{ disabled: disabled }"
              @click="
                () => {
                  click(btn.id)
                  if (!disabled && btn.click) {
                    btn.click()
                  }
                }
              "
            >
              {{ btn.text }}
            </td>
          </tr>
        </table>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
export type Controller = {
  submit(id?: string): void
}

export type Button = {
  text: string
  id?: string
  click?: (() => void) | (() => Promise<void>)
}

const props = defineProps<{
  isVisible: boolean
  buttons: Button[]
  disabled?: boolean
  ctrl?: Controller
}>()

const emit = defineEmits<{
  (e: "click", id?: string): void
}>()

function click(id?: string) {
  if (props.disabled) {
    return
  }
  if (props.ctrl) {
    props.ctrl.submit(id)
  }
  emit("click", id)
}
</script>

<style scoped lang="sass">
@use '@/css/theme'

.fade-enter-active, .fade-leave-active
  transition: all 200ms linear

.fade-enter-from
  opacity: 0
  transform: scale(.95)

.fade-leave-to
  opacity: 0
  transform: scale(1.05)

.background
  position: fixed
  left: 0
  top: 0
  height: 100vh
  width: 100vw
  backdrop-filter: blur(3px)
  background-color: var(--color-background)
  opacity: 0.6
  z-index: 200

.container
  position: fixed
  z-index: 250
  top: 50%
  left: 50%
  transform: translate(-50%, -50%)

.content
  @include theme.shadow-l

  border-radius: 5px
  background-color: var(--color-background)
  text-align: center
  max-width: 500px

.aligner
  position: fixed
  top: 0
  left: 0
  height: 100vh
  width: 100vw

.wrapper
  padding: 1rem 3rem

table
  border-top: 1px solid var(--color-outline)
  width: 100%
  table-layout: fixed

.cell
  @include theme.clickable
  padding: 0.5rem 0
  font-weight: 500
  &.disabled
    cursor: default
    background-color: var(--color-fade-background)
  &:hover:not(.disabled)
    background-color: var(--color-highlight-background)

.cell + .cell
  border-left: 1px solid var(--color-outline)
</style>
