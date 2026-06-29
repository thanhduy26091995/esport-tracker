const BASE = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

export interface SiteAccessQuestion {
  question: string
  enabled: boolean
}

export async function getQuestion(): Promise<SiteAccessQuestion> {
  const r = await fetch(`${BASE}/site-access/question`)
  if (!r.ok) throw new Error('failed to get question')
  return r.json()
}

export async function validateAnswer(answer: string): Promise<string> {
  const r = await fetch(`${BASE}/site-access/validate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ answer }),
  })
  if (!r.ok) throw new Error('incorrect')
  const data = await r.json()
  return data.token as string
}
