// Minimal type declarations for Google Identity Services (GSI)
interface Window {
  google?: {
    accounts: {
      id: {
        initialize: (config: {
          client_id: string
          callback: (response: { credential: string }) => void
          auto_select?: boolean
        }) => void
        renderButton: (
          element: HTMLElement,
          options: {
            theme?: string
            size?: string
            text?: string
            locale?: string
            width?: number
          }
        ) => void
        prompt: () => void
      }
    }
  }
}
