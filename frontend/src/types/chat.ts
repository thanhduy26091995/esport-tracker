export interface ChatMessage {
  id: string
  user_id: string
  user_name: string
  avatar_url: string
  message: string
  created_at: string // ISO 8601
}

export interface ChatSendFrame {
  type: 'chat_send'
  message: string
  mentions?: string[] // user UUIDs
}

export interface ChatMessageEvent {
  type: 'chat_message'
  id: string
  user_id: string
  user_name: string
  avatar_url: string
  message: string
  created_at: string
}

export interface ChatMentionEvent {
  type: 'chat_mention'
  message_id: string
  sender_id: string
  sender_name: string
  message: string
}

export interface ChatErrorFrame {
  type: 'error'
  message: string
}

export interface WcUserForMention {
  id: string
  name: string
  avatar_url: string | null
}
