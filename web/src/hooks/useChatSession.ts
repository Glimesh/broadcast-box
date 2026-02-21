import { useCallback, useEffect, useMemo, useState } from 'react'

export type ChatStatus = 'connecting' | 'connected' | 'error' | 'disconnected'

export interface Message {
  id: string
  ts: number
  text: string
  displayName: string
}

export interface ChatAdapter {
  connect(streamKey: string): Promise<void>
  subscribe(
    onMessage: (message: Message) => void,
    onStatus: (status: ChatStatus) => void,
    onError: (error: string) => void,
  ): () => void
  send(text: string, displayName: string): Promise<void>
}

const MAX_MESSAGES = 1000

export const useChatSession = (streamKey: string, adapter?: ChatAdapter) => {
  const [messages, setMessages] = useState<Message[]>([])
  const [status, setStatus] = useState<ChatStatus>('disconnected')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setMessages([])
    setError(null)

    if (!adapter) {
      setStatus('disconnected')
      return
    }

    let unsubscribe = () => {}
    let stopped = false

    setStatus('connecting')

    adapter
      .connect(streamKey)
      .then(() => {
        if (stopped) {
          return
        }

        unsubscribe = adapter.subscribe(
          (message) => {
            if (stopped) {
              return
            }

            setMessages((current) => {
              if (current.some((existing) => existing.id === message.id)) {
                return current
              }

              const next = [...current, message]
              if (next.length <= MAX_MESSAGES) {
                return next
              }

              return next.slice(next.length - MAX_MESSAGES)
            })
          },
          (nextStatus) => {
            if (!stopped) {
              setStatus(nextStatus)
            }
          },
          (nextError) => {
            if (!stopped) {
              setError(nextError)
              setStatus('error')
            }
          },
        )
      })
      .catch((connectError) => {
        if (!stopped) {
          const message = connectError instanceof Error ? connectError.message : 'Failed to connect chat'
          setError(message)
          setStatus('error')
        }
      })

    return () => {
      stopped = true
      unsubscribe()
    }
  }, [adapter, streamKey])

  const sendMessage = useCallback(
    async (text: string, displayName: string) => {
      if (!adapter) {
        throw new Error('Chat is not connected')
      }

      setError(null)
      await adapter.send(text, displayName)
    },
    [adapter],
  )

  return useMemo(
    () => ({
      messages,
      status,
      error,
      sendMessage,
    }),
    [error, messages, sendMessage, status],
  )
}
