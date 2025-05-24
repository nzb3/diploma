import { createContext, useContext, useState, useRef, ReactNode, useEffect, useCallback } from 'react';
import { streamAnswer, cancelStream } from '@/services/api';
import { Message, CompleteResult, Reference } from '@/types';

const decodeHtmlEntities = (html: string): string => {
  const textArea = document.createElement('textarea');
  textArea.innerHTML = html;
  return textArea.value;
};

const SSE_TIMEOUT = 20000;

interface ChatContextType {
  messages: Message[];
  isLoading: boolean;
  submitQuestion: (question: string, numReferences?: number, usePreviousMessages?: boolean) => Promise<void>;
  cancelGeneration: () => Promise<void>;
  decodeHtmlEntities: (text: string) => string;
  retryGeneration: (messageIndex: number) => Promise<void>;
  retryingMessageIndex: number | null;
  deleteMessage: (messageIndex: number) => void;
  clearChat: () => void;
}

const ChatContext = createContext<ChatContextType>({
  messages: [],
  isLoading: false,
  submitQuestion: async () => {},
  cancelGeneration: async () => {},
  decodeHtmlEntities: (text) => text,
  retryGeneration: async () => {},
  retryingMessageIndex: null,
  deleteMessage: () => {},
  clearChat: () => {},
});

export const useChat = () => useContext(ChatContext);

interface ChatProviderProps {
  children: ReactNode;
}

export function ChatProvider({ children }: ChatProviderProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [currentProcessId, setCurrentProcessId] = useState<string | null>(null);
  const [retryingMessageIndex, setRetryingMessageIndex] = useState<number | null>(null);
  const hasReceivedData = useRef<boolean>(false);
  const currentAnswerRef = useRef<string>('');
  const currentReferencesRef = useRef<Reference[]>([]);
  const eventSourceRef = useRef<EventSource | null>(null);
  const lastActivityTimestampRef = useRef<number>(Date.now());
  const timeoutIdRef = useRef<number | null>(null);

  const cleanupConnection = useCallback(() => {
    if (timeoutIdRef.current !== null) {
      window.clearTimeout(timeoutIdRef.current);
      timeoutIdRef.current = null;
    }

    setIsLoading(false);
    setCurrentProcessId(null);
    setRetryingMessageIndex(null);

    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
  }, []);

  const resetActivityTimer = useCallback(() => {
    lastActivityTimestampRef.current = Date.now();

    if (timeoutIdRef.current !== null) {
      window.clearTimeout(timeoutIdRef.current);
    }

    timeoutIdRef.current = window.setTimeout(() => {
      const timeElapsed = Date.now() - lastActivityTimestampRef.current;

      if (timeElapsed >= SSE_TIMEOUT && isLoading && eventSourceRef.current) {
        console.warn('SSE connection inactive for too long, closing connection');
        cleanupConnection();
      }
    }, SSE_TIMEOUT);
  }, [isLoading, cleanupConnection]);

  const updateMessageContent = useCallback((
      messageIndex: number | null,
      content: string,
      references: Reference[],
      isComplete = false,
      isStopped = false
  ) => {
    setMessages(prev => {
      const updatedMessages = [...prev];

      if (messageIndex !== null && updatedMessages[messageIndex]) {
        updatedMessages[messageIndex] = {
          ...updatedMessages[messageIndex],
          content,
          references,
          complete: isComplete,
          stopped: isStopped
        };
        return updatedMessages;
      }

      const lastMessage = updatedMessages[updatedMessages.length - 1];
      if (lastMessage?.role === 'assistant') {
        lastMessage.content = content;
        lastMessage.references = references;
        if (isComplete) lastMessage.complete = true;
        return updatedMessages;
      } else {
        return [
          ...updatedMessages,
          { role: 'assistant', content, references, complete: isComplete, stopped: isStopped }
        ];
      }
    });
  }, []);

  const handleStreamError = useCallback((targetIndex: number | null) => {
    if (!hasReceivedData.current) {
      const errorMessage = 'Sorry, I was unable to generate a response. Please try again.';

      if (targetIndex !== null) {
        updateMessageContent(targetIndex, errorMessage, [], true);
      } else {
        setMessages(prev => {
          const newMessages = [...prev];
          const lastUserMessageIndex = newMessages.findIndex(m => m.role === 'user');

          if (lastUserMessageIndex !== -1) {
            return [
              ...newMessages.slice(0, lastUserMessageIndex + 1),
              { role: 'assistant', content: errorMessage, complete: true, references: [] }
            ];
          }
          return newMessages;
        });
      }
    }

    cleanupConnection();
  }, [cleanupConnection, updateMessageContent]);

  const setupStreamEventListeners = useCallback((
      eventSource: EventSource,
      targetMessageIndex: number | null = null
  ) => {
    eventSource.addEventListener('chunk', (event) => {
      resetActivityTimer();

      try {
        const data = JSON.parse(event.data);
        hasReceivedData.current = true;

        if (data.content) {
          if (data.process_id) {
            setCurrentProcessId(data.process_id);
          }

          const decodedContent = decodeHtmlEntities(data.content);
          currentAnswerRef.current += decodedContent;

          updateMessageContent(
              targetMessageIndex,
              currentAnswerRef.current,
              currentReferencesRef.current
          );
        }
      } catch (error) {
        console.error('Error parsing chunk message:', error);
      }
    });

    eventSource.addEventListener('complete', (event) => {
      resetActivityTimer();

      try {
        const data = JSON.parse(event.data);

        if (data.complete === true && data.result) {
          const result = data.result as CompleteResult;
          updateMessageContent(
              targetMessageIndex,
              result.answer,
              currentReferencesRef.current,
              true
          );
        }

        cleanupConnection();
      } catch (error) {
        console.error('Error parsing complete message:', error);
      }
    });

    eventSource.addEventListener('references', (event) => {
      resetActivityTimer();

      try {
        const data = JSON.parse(event.data);

        if (data.references && Array.isArray(data.references)) {
          currentReferencesRef.current = data.references;
          updateMessageContent(
              targetMessageIndex,
              currentAnswerRef.current,
              data.references
          );
        }
      } catch (error) {
        console.error('Error parsing references message:', error);
      }
    });

    eventSource.addEventListener('error', (event) => {
      console.error('Error event from SSE connection:', event);
      handleStreamError(targetMessageIndex);
    });
  }, [resetActivityTimer, updateMessageContent, handleStreamError, cleanupConnection]);

  const initiateStreaming = useCallback(async (
      prompt: string,
      numReferences: number = 5,
      targetMessageIndex: number | null = null
  ) => {
    try {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }

      const eventSource = await streamAnswer(prompt, numReferences);
      eventSourceRef.current = eventSource;

      setupStreamEventListeners(eventSource, targetMessageIndex);
      return true;
    } catch (error) {
      console.error('Error setting up SSE connection:', error);

      const errorMessage = 'Sorry, there was an error connecting to the service. Please try again later.';
      updateMessageContent(targetMessageIndex, errorMessage, [], true);

      cleanupConnection();
      return false;
    }
  }, [setupStreamEventListeners, cleanupConnection, updateMessageContent]);

  const retryGeneration = async (messageIndex: number) => {
    const previousAnswer = messages[messageIndex].content;

    let userMessageIndex = messageIndex - 1;
    while (userMessageIndex >= 0 && messages[userMessageIndex].role !== 'user') {
      userMessageIndex--;
    }

    if (userMessageIndex >= 0 && !isLoading) {
      const userMessage = messages[userMessageIndex];

      updateMessageContent(messageIndex, 'Regenerating answer...', []);

      hasReceivedData.current = false;
      currentAnswerRef.current = '';
      currentReferencesRef.current = [];
      setIsLoading(true);
      setRetryingMessageIndex(messageIndex);

      resetActivityTimer();

      const promptForRetry = `You gave me a bad answer to a question. Try it again, but paraphrase your answer, try to make it better then previous one.
** Your previous answer **
${previousAnswer}
** Question **
${userMessage.content}`;

      await initiateStreaming(promptForRetry, 5, messageIndex);
    }
  };

  const submitQuestion = async (question: string, numReferences: number = 5, usePreviousMessages: boolean = false) => {
    if (!question.trim() || isLoading) return;

    hasReceivedData.current = false;
    currentAnswerRef.current = '';
    currentReferencesRef.current = [];

    let fullPrompt = question;

    if (usePreviousMessages && messages.length > 0) {
      const conversationHistory = messages.map(msg =>
          `${msg.role === 'user' ? 'User' : 'Assistant'}: ${msg.content}`
      ).join('\n\n');

      fullPrompt = `I have the following conversation history:\n\n${conversationHistory}\n\nBased on this history, please answer my new question: ${question}`;
    }

    setMessages(prev => [...prev, { role: 'user', content: question }]);
    setIsLoading(true);
    resetActivityTimer();
    await initiateStreaming(fullPrompt, numReferences);
  };

  const cancelGeneration = async () => {
    if (currentProcessId) {
      try {
        await cancelStream(currentProcessId);
      } catch (error) {
        console.error('Error cancelling stream:', error);
      }
    }
    
    if (retryingMessageIndex !== null) {
      updateMessageContent(
        retryingMessageIndex, 
        currentAnswerRef.current, 
        currentReferencesRef.current, 
        false, 
        true
      );
    } else {
      updateMessageContent(
        null, 
        currentAnswerRef.current, 
        currentReferencesRef.current, 
        false, 
        true
      );
    }
    
    cleanupConnection();
  };

  const deleteMessage = (messageIndex: number) => {
    const messageToDelete = messages[messageIndex];

    if (messageToDelete.role === 'assistant') {
      let userMessageIndex = messageIndex - 1;
      while (userMessageIndex >= 0 && messages[userMessageIndex].role !== 'user') {
        userMessageIndex--;
      }

      setMessages(prev => {
        const newMessages = [...prev];
        if (userMessageIndex >= 0) {
          newMessages.splice(userMessageIndex, messageIndex - userMessageIndex + 1);
        } else {
          newMessages.splice(messageIndex, 1);
        }
        return newMessages;
      });
    } else if (messageToDelete.role === 'user') {
      let nextIndex = messageIndex + 1;
      let deleteCount = 1;

      while (nextIndex < messages.length && messages[nextIndex].role === 'assistant') {
        deleteCount++;
        nextIndex++;
      }

      setMessages(prev => {
        const newMessages = [...prev];
        newMessages.splice(messageIndex, deleteCount);
        return newMessages;
      });
    }
  };

  const clearChat = () => {
    if (isLoading) {
      cancelGeneration();
    }
    setMessages([]);
  };

  useEffect(() => {
    return () => {
      cleanupConnection();
    };
  }, [cleanupConnection]);

  return (
      <ChatContext.Provider value={{
        messages,
        isLoading,
        submitQuestion,
        cancelGeneration,
        decodeHtmlEntities,
        retryGeneration,
        retryingMessageIndex,
        deleteMessage,
        clearChat,
      }}>
        {children}
      </ChatContext.Provider>
  );
}
