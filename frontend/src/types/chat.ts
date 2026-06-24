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

export interface ChatErrorFrame {
  type: 'error'
  message: string
}
