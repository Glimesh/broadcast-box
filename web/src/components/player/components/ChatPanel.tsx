import React, { FormEvent, memo, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { ChatBubbleLeftRightIcon, PencilSquareIcon, PaperAirplaneIcon } from '@heroicons/react/24/outline'
import ModalTextInput from '../../shared/ModalTextInput'
import { ChatAdapter, ChatStatus, Message, useChatSession } from '../../../hooks/useChatSession'

type ChatVariant = 'sidebar' | 'below'

interface ChatPanelProps {
  streamKey: string
  variant: ChatVariant
  isOpen: boolean
  adapter?: ChatAdapter
  fixedHeightPx?: number
}

const getNameColor = (displayName: string) => {
  let hash = 0
  for (let i = 0; i < displayName.length; i += 1) {
    hash = displayName.charCodeAt(i) + ((hash << 5) - hash)
  }

  return `hsl(${Math.abs(hash) % 360}, 70%, 60%)`
}

const ChatMessage = memo(function ChatMessage(props: { message: Message }) {
  const { message } = props
  const timestamp = new Date(message.ts).toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
  })

  return (
    <div className="bg-gray-900/40 p-2">
      <div className="flex items-center gap-2 text-xs">
        <span className="max-w-56 truncate font-bold" style={{ color: getNameColor(message.displayName) }}>
          {message.displayName}
        </span>
        <span className="text-gray-400">{timestamp}</span>
      </div>
      <p className="mt-1 break-words text-sm text-gray-100">{message.text}</p>
    </div>
  )
})

interface ChatComposerProps {
  status: ChatStatus
  isSending: boolean
  onNameRequested(): void
  onSend(text: string): Promise<boolean>
}

const ChatComposer = memo(function ChatComposer(props: ChatComposerProps) {
  const { status, isSending, onNameRequested, onSend } = props
  const [text, setText] = useState('')
  const canSend = text.trim().length > 0 && !isSending && status === 'connected'

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()

    if (!text.trim()) {
      return
    }

    const sent = await onSend(text)
    if (sent) {
      setText('')
    }
  }

  return (
    <form onSubmit={submit} className="border-t border-gray-700 bg-gray-900/70 p-3">
      <div className="flex items-center gap-2">
        <input
          type="text"
          value={text}
          maxLength={2000}
          onChange={(event) => setText(event.target.value)}
          placeholder="Write a message"
          className="h-9 flex-1 rounded-md border border-gray-700 bg-gray-800 px-3 text-sm text-gray-100 placeholder:text-gray-400 focus:outline-hidden"
        />

        <button
          type="button"
          onClick={onNameRequested}
          className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-gray-700 bg-gray-800 text-gray-100 hover:bg-gray-700"
          title="Change display name"
        >
          <PencilSquareIcon className="h-5 w-5" />
        </button>

        <button
          type="submit"
          disabled={!canSend}
          className="inline-flex h-9 w-9 items-center justify-center rounded-md bg-blue-600 text-white disabled:cursor-not-allowed disabled:bg-gray-700 disabled:text-gray-400"
          title="Send message"
        >
          <PaperAirplaneIcon className="h-5 w-5" />
        </button>
      </div>

      {text.length > 1800 && <div className="mt-1 text-right text-xs text-gray-400">{text.length}/2000</div>}
    </form>
  )
})

const statusColorClass = (status: ChatStatus) => {
  if (status === 'connected') {
    return 'bg-green-500'
  }

  if (status === 'connecting') {
    return 'animate-pulse bg-yellow-400'
  }

  return 'bg-red-500'
}

const ChatPanel = (props: ChatPanelProps) => {
  const { streamKey, variant, isOpen, adapter, fixedHeightPx } = props
  const { messages, status, error, sendMessage } = useChatSession(streamKey, adapter)

  const [displayName, setDisplayName] = useState('')
  const [isNameModalOpen, setIsNameModalOpen] = useState(false)
  const [isSending, setIsSending] = useState(false)
  const [sendError, setSendError] = useState<string | null>(null)

  const messageListRef = useRef<HTMLDivElement>(null)
  const shouldStickToBottomRef = useRef(true)
  const firstBatchRef = useRef(true)

  useEffect(() => {
    const storedName = localStorage.getItem('chatDisplayName')
    if (storedName) {
      setDisplayName(storedName)
    }
  }, [])

  useEffect(() => {
    if (!isOpen) {
      return
    }

    const node = messageListRef.current
    if (!node) {
      return
    }

    if (firstBatchRef.current || shouldStickToBottomRef.current) {
      node.scrollTop = node.scrollHeight
      firstBatchRef.current = false
    }
  }, [isOpen, messages])

  useEffect(() => {
    firstBatchRef.current = true
    shouldStickToBottomRef.current = true
  }, [streamKey])

  const onMessageListScroll = () => {
    const node = messageListRef.current
    if (!node) {
      return
    }

    const distanceToBottom = node.scrollHeight - node.scrollTop - node.clientHeight
    shouldStickToBottomRef.current = distanceToBottom <= 100
  }

  const saveDisplayName = useCallback((value: string) => {
    const nextValue = value.trim()
    if (!nextValue) {
      return
    }

    setDisplayName(nextValue)
    localStorage.setItem('chatDisplayName', nextValue)
    setIsNameModalOpen(false)
  }, [])

  const onSend = useCallback(
    async (text: string) => {
      if (!displayName.trim()) {
        setIsNameModalOpen(true)
        return false
      }

      setIsSending(true)
      setSendError(null)

      try {
        await sendMessage(text.trim(), displayName.trim())
        return true
      } catch (nextError) {
        const message = nextError instanceof Error ? nextError.message : 'Failed to send message'
        setSendError(message)
        return false
      } finally {
        setIsSending(false)
      }
    },
    [displayName, sendMessage],
  )

  const panelClassName = useMemo(() => {
    const base =
      'flex flex-col overflow-hidden rounded-md border border-gray-700 bg-slate-900 text-gray-100 transition-[max-height,max-width,opacity,transform,border-color] duration-200 ease-out'
    if (variant === 'sidebar') {
      return `${base} min-h-0 shrink-0 ${
        isOpen
          ? 'h-72 w-full max-h-96 translate-x-0 translate-y-0 opacity-100 2xl:max-h-none 2xl:max-w-sm'
          : 'h-0 w-full max-h-0 translate-y-1 opacity-0 pointer-events-none border-transparent 2xl:h-72 2xl:max-h-none 2xl:max-w-0 2xl:translate-x-2 2xl:translate-y-0'
      }`
    }

    return `${base} ${isOpen ? 'max-h-96 translate-y-0 opacity-100' : 'max-h-0 translate-y-1 border-transparent opacity-0 pointer-events-none'}`
  }, [isOpen, variant])

  return (
    <div className={panelClassName} style={variant === 'sidebar' && isOpen && fixedHeightPx ? { height: `${fixedHeightPx}px` } : undefined}>
      <div className="flex items-center justify-between border-b border-gray-700 bg-gray-900/70 px-3 py-2">
        <div className="flex items-center gap-2 text-sm font-semibold">
          <ChatBubbleLeftRightIcon className="h-4 w-4" />
          <span>Chat</span>
        </div>

        <div className="flex items-center gap-2 text-xs text-gray-300">
          <span className={`inline-flex h-2.5 w-2.5 rounded-full ${statusColorClass(status)}`} />
          <span className="capitalize">{status}</span>
        </div>
      </div>

      <div ref={messageListRef} onScroll={onMessageListScroll} className="flex-1 overflow-y-auto px-3 py-2">
        {error && <div className="mb-2 rounded-md border border-red-400 bg-red-950/40 px-2 py-1 text-xs text-red-200">{error}</div>}
        {sendError && (
          <div className="mb-2 rounded-md border border-red-400 bg-red-950/40 px-2 py-1 text-xs text-red-200">{sendError}</div>
        )}

        {!error && messages.length === 0 && status === 'connected' && (
          <div className="text-xs text-gray-400">No chat messages yet.</div>
        )}

        <div className="space-y-0">
          {messages.map((message) => (
            <ChatMessage key={message.id} message={message} />
          ))}
        </div>
      </div>

      <ChatComposer
        status={status}
        isSending={isSending}
        onNameRequested={() => setIsNameModalOpen(true)}
        onSend={onSend}
      />

      {isNameModalOpen && (
        <ModalTextInput<string>
          title="Display name"
          message="Set your display name for chat"
          placeholder="Enter display name"
          isOpen={isNameModalOpen}
          canCloseOnBackgroundClick
          onClose={() => setIsNameModalOpen(false)}
          onAccept={saveDisplayName}
        />
      )}
    </div>
  )
}

export default ChatPanel
