import { useCallback, useEffect, useRef, useState } from 'react';
import type { ChatData, ChatSession, Message, StoredMessage } from '../types';
import { useChatStorage } from '../context/StorageContext';

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
  const chatStorage = useChatStorage();
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
  const isMountedRef = useRef(true);
  const loadVersionRef = useRef(0);

  useEffect(() => {
    return () => {
      isMountedRef.current = false;
      loadVersionRef.current += 1;
    };
  }, []);

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

  const logStorageError = useCallback((action: string, error: unknown) => {
    console.error(`[useChats] ${action} failed`, error);
  }, []);

  const loadChatList = useCallback(async (): Promise<ChatSession[]> => {
    try {
      const list = chatStorage ? await chatStorage.list() : [];
      if (isMountedRef.current) {
        setChats(list);
      }
      return list;
    } catch (error) {
      logStorageError('load chat list', error);
      return [];
    }
  }, [chatStorage, logStorageError]);

  const loadChat = useCallback(async (chatId: string) => {
    if (!chatStorage) {
      return;
    }

    const loadVersion = ++loadVersionRef.current;

    try {
      const data = await chatStorage.get(chatId);
      if (!data || !isMountedRef.current || loadVersion !== loadVersionRef.current) {
        return;
      }

      setActiveChatId(data.id);
      setMessages(data.messages.map(fromStoredMessage));
      setLlmHistory(data.llmHistory ?? []);
      setContextTokensUsed(data.contextTokensUsed ?? 0);
      setIsTyping(false);
    } catch (error) {
      logStorageError(`load chat ${chatId}`, error);
    }
  }, [chatStorage, logStorageError]);

  const saveCurrentChat = useCallback(async () => {
    const chatId = activeChatIdRef.current;

    if (!chatStorage || !chatId || savingRef.current) {
      return;
    }

    savingRef.current = true;
    try {
      const existing = await chatStorage.get(chatId);
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
      const updatedList = await chatStorage.save(chatId, data);
      if (isMountedRef.current) {
        setChats(updatedList);
      }
    } catch (error) {
      logStorageError(`save chat ${chatId}`, error);
      throw error;
    } finally {
      savingRef.current = false;
    }
  }, [chatStorage, logStorageError]);

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
      if (!isMountedRef.current) {
        return;
      }
      setActiveChatId(null);
      setMessages([]);
      setLlmHistory([]);
      setContextTokensUsed(0);
    } finally {
      switchingRef.current = false;
    }
  }, [saveCurrentChat]);

  const createChat = useCallback(async (): Promise<string> => {
    if (!chatStorage) {
      throw new Error('History bridge is unavailable');
    }

    const id = generateChatId();
    try {
      await chatStorage.save(id, emptyChatData(id));
      await loadChatList();
      return id;
    } catch (error) {
      logStorageError(`create chat ${id}`, error);
      throw error;
    }
  }, [chatStorage, loadChatList, logStorageError]);

  const renameChat = useCallback(async (chatId: string, title: string) => {
    if (!chatStorage) {
      return;
    }

    try {
      const data = await chatStorage.get(chatId);
      if (!data) {
        return;
      }

      const nextTitle = title.trim() || 'Untitled Chat';
      const updatedData: ChatData = {
        ...data,
        title: nextTitle,
        updatedAt: Date.now(),
      };
      const updatedList = await chatStorage.save(chatId, updatedData);
      if (isMountedRef.current) {
        setChats(updatedList);
      }
    } catch (error) {
      logStorageError(`rename chat ${chatId}`, error);
    }
  }, [chatStorage, logStorageError]);

  const deleteChat = useCallback(
    async (chatId: string) => {
      if (!chatStorage) {
        return;
      }

      try {
        const list = await chatStorage.delete(chatId);
        if (isMountedRef.current) {
          setChats(list);
        }

        if (activeChatIdRef.current !== chatId) {
          return;
        }

        const nextId = list[0]?.id ?? null;
        if (nextId) {
          await loadChat(nextId);
          return;
        }

        if (!isMountedRef.current) {
          return;
        }

        setActiveChatId(null);
        setMessages([]);
        setLlmHistory([]);
        setContextTokensUsed(0);
        setIsTyping(false);
      } catch (error) {
        logStorageError(`delete chat ${chatId}`, error);
      }
    },
    [chatStorage, loadChat, logStorageError],
  );

  const persistCurrentChat = useCallback(async () => {
    await saveCurrentChat();
  }, [saveCurrentChat]);

  useEffect(() => {
    async function bootstrap() {
      const list = await loadChatList();
      if (list.length > 0) {
        await loadChat(list[0].id);
      }
    }

    void bootstrap();
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
