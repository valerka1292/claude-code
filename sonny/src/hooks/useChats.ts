import { useCallback, useEffect, useRef, useState } from 'react';
import type { ChatData, ChatSession, Message, StoredMessage } from '../types';

function generateChatId(): string {
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`;
}

function toStoredMessage(message: Message): StoredMessage {
  return {
    id: message.id,
    role: message.role,
    content: message.content,
    timestamp: message.timestamp.getTime(),
    thinking: message.thinking,
    toolCalls: message.toolCalls,
  };
}

function fromStoredMessage(message: StoredMessage): Message {
  return {
    id: message.id,
    role: message.role,
    content: message.content,
    timestamp: new Date(message.timestamp),
    thinking: message.thinking,
    toolCalls: message.toolCalls,
  };
}

function emptyChatData(id: string): ChatData {
  const now = Date.now();
  return {
    id,
    title: 'Untitled Chat',
    createdAt: now,
    updatedAt: now,
    messages: [],
    llmHistory: [],
    contextTokensUsed: 0,
  };
}

export function useChats() {
  const [chats, setChats] = useState<ChatSession[]>([]);
  const [activeChatId, setActiveChatId] = useState<string | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [llmHistory, setLlmHistory] = useState<{ role: string; content: string }[]>([]);
  const [contextTokensUsed, setContextTokensUsed] = useState(0);
  const [isTyping, setIsTyping] = useState(false);

  const activeChatIdRef = useRef<string | null>(null);
  const messagesRef = useRef<Message[]>([]);
  const llmHistoryRef = useRef<{ role: string; content: string }[]>([]);
  const contextTokensUsedRef = useRef(0);
  const savingRef = useRef(false);
  const switchingRef = useRef(false);

  useEffect(() => {
    activeChatIdRef.current = activeChatId;
  }, [activeChatId]);

  useEffect(() => {
    messagesRef.current = messages;
  }, [messages]);

  useEffect(() => {
    llmHistoryRef.current = llmHistory;
  }, [llmHistory]);

  useEffect(() => {
    contextTokensUsedRef.current = contextTokensUsed;
  }, [contextTokensUsed]);

  const loadChatList = useCallback(async (): Promise<ChatSession[]> => {
    const list = (await window.electron?.history?.list()) ?? [];
    setChats(list);
    return list;
  }, []);

  const loadChat = useCallback(async (chatId: string) => {
    const data = await window.electron?.history?.get(chatId);
    if (!data) {
      return;
    }

    setActiveChatId(data.id);
    setMessages(data.messages.map(fromStoredMessage));
    setLlmHistory(data.llmHistory ?? []);
    setContextTokensUsed(data.contextTokensUsed ?? 0);
    setIsTyping(false);
  }, []);

  const saveCurrentChat = useCallback(async () => {
    const chatId = activeChatIdRef.current;
    const bridge = window.electron?.history;

    if (!bridge || !chatId || savingRef.current) {
      return;
    }

    savingRef.current = true;
    try {
      const existing = await bridge.get(chatId);
      const now = Date.now();
      const data: ChatData = {
        id: chatId,
        title: existing?.title ?? 'Untitled Chat',
        createdAt: existing?.createdAt ?? now,
        updatedAt: now,
        messages: messagesRef.current.map(toStoredMessage),
        llmHistory: llmHistoryRef.current,
        contextTokensUsed: contextTokensUsedRef.current,
      };
      const updatedList = await bridge.save(chatId, data);
      setChats(updatedList);
    } finally {
      savingRef.current = false;
    }
  }, []);

  const switchChat = useCallback(
    async (chatId: string) => {
      if (switchingRef.current || activeChatIdRef.current === chatId) {
        return;
      }

      switchingRef.current = true;
      try {
        setIsTyping(false);
        await saveCurrentChat();
        await loadChat(chatId);
      } finally {
        switchingRef.current = false;
      }
    },
    [loadChat, saveCurrentChat],
  );

  const newChat = useCallback(async () => {
    if (switchingRef.current) {
      return;
    }

    switchingRef.current = true;
    try {
      setIsTyping(false);
      await saveCurrentChat();
      setActiveChatId(null);
      setMessages([]);
      setLlmHistory([]);
      setContextTokensUsed(0);
    } finally {
      switchingRef.current = false;
    }
  }, [saveCurrentChat]);

  const createChat = useCallback(async (): Promise<string> => {
    const bridge = window.electron?.history;
    if (!bridge) {
      throw new Error('History bridge is unavailable');
    }

    const id = generateChatId();
    await bridge.save(id, emptyChatData(id));
    await loadChatList();
    return id;
  }, [loadChatList]);

  const renameChat = useCallback(async (chatId: string, title: string) => {
    const bridge = window.electron?.history;
    if (!bridge) {
      return;
    }

    const data = await bridge.get(chatId);
    if (!data) {
      return;
    }

    const nextTitle = title.trim() || 'Untitled Chat';
    const updatedData: ChatData = {
      ...data,
      title: nextTitle,
      updatedAt: Date.now(),
    };
    const updatedList = await bridge.save(chatId, updatedData);
    setChats(updatedList);
  }, []);

  const deleteChat = useCallback(
    async (chatId: string) => {
      const bridge = window.electron?.history;
      if (!bridge) {
        return;
      }

      const list = await bridge.delete(chatId);
      setChats(list);

      if (activeChatIdRef.current !== chatId) {
        return;
      }

      const nextId = list[0]?.id ?? null;
      if (nextId) {
        await loadChat(nextId);
        return;
      }

      setActiveChatId(null);
      setMessages([]);
      setLlmHistory([]);
      setContextTokensUsed(0);
      setIsTyping(false);
    },
    [loadChat],
  );

  const persistCurrentChat = useCallback(async () => {
    await saveCurrentChat();
  }, [saveCurrentChat]);

  useEffect(() => {
    let mounted = true;

    async function bootstrap() {
      const list = await loadChatList();
      if (!mounted) {
        return;
      }

      if (list.length > 0) {
        await loadChat(list[0].id);
      }
    }

    void bootstrap();

    return () => {
      mounted = false;
    };
  }, [loadChat, loadChatList]);

  return {
    chats,
    activeChatId,
    activeChatIdRef,
    messages,
    setMessages,
    llmHistory,
    setLlmHistory,
    contextTokensUsed,
    setContextTokensUsed,
    isTyping,
    setIsTyping,
    loadChat,
    switchChat,
    newChat,
    createChat,
    renameChat,
    deleteChat,
    persistCurrentChat,
  };
}
