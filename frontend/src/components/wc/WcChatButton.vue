<template>
  <button class="chat-fab" :class="{ 'is-open': chatStore.isPanelOpen }" @click="toggle" :aria-label="t('wc.chat.title')">
    <!-- Chat icon -->
    <svg v-if="!chatStore.isPanelOpen" class="fab-icon" width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
      <path d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2z"/>
      <circle cx="8" cy="11" r="1" fill="rgba(255,255,255,0.7)"/>
      <circle cx="12" cy="11" r="1" fill="rgba(255,255,255,0.7)"/>
      <circle cx="16" cy="11" r="1" fill="rgba(255,255,255,0.7)"/>
    </svg>
    <!-- Close icon when open -->
    <svg v-else class="fab-icon" width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
      <path d="M19 6.41L17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/>
    </svg>

    <span v-if="chatStore.unreadCount > 0 && !chatStore.isPanelOpen" class="chat-badge">
      {{ chatStore.unreadCount > 99 ? '99+' : chatStore.unreadCount }}
    </span>

    <!-- Pulse ring when there are unread messages -->
    <span v-if="chatStore.unreadCount > 0 && !chatStore.isPanelOpen" class="pulse-ring" />
  </button>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useChatStore } from '@/stores/chatStore'

const { t } = useI18n()
const chatStore = useChatStore()

function toggle() {
  if (chatStore.isPanelOpen) {
    chatStore.closePanel()
  } else {
    chatStore.openPanel()
  }
}
</script>

<style scoped>
.chat-fab {
  position: fixed;
  bottom: 24px;
  right: 24px;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: linear-gradient(135deg, #6366f1 0%, #3b82f6 100%);
  color: #fff;
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow:
    0 4px 14px rgba(99, 102, 241, 0.5),
    0 2px 6px rgba(0, 0, 0, 0.15);
  z-index: 2000;
  transition: transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1),
              box-shadow 0.2s ease,
              background 0.2s ease;
}

.chat-fab:hover {
  transform: scale(1.12) translateY(-2px);
  box-shadow:
    0 8px 20px rgba(99, 102, 241, 0.6),
    0 4px 10px rgba(0, 0, 0, 0.2);
}

.chat-fab:active {
  transform: scale(0.96);
}

.chat-fab.is-open {
  background: linear-gradient(135deg, #6b7280 0%, #4b5563 100%);
  box-shadow:
    0 4px 14px rgba(75, 85, 99, 0.4),
    0 2px 6px rgba(0, 0, 0, 0.15);
}

.fab-icon {
  transition: transform 0.2s ease;
  filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.2));
}

.chat-fab:hover .fab-icon {
  transform: scale(1.05);
}

.chat-badge {
  position: absolute;
  top: -3px;
  right: -3px;
  background: #ef4444;
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  min-width: 20px;
  height: 20px;
  border-radius: 10px;
  padding: 0 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 2px solid #fff;
  box-shadow: 0 2px 6px rgba(239, 68, 68, 0.5);
}

/* Pulse animation on unread */
.pulse-ring {
  position: absolute;
  inset: -4px;
  border-radius: 50%;
  border: 2px solid rgba(99, 102, 241, 0.6);
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  pointer-events: none;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50%       { opacity: 0; transform: scale(1.25); }
}
</style>
