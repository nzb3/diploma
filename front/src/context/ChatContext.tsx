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
      
      if (timeElapsed >= SSE_TIMEOUT) {
        console.warn('SSE connection inactive for too long, closing connection');
        
        if (isLoading && eventSourceRef.current) {
          console.log('Closing inactive SSE connection');
          cleanupConnection();
        }
      }
    }, SSE_TIMEOUT);
  }, [isLoading, cleanupConnection]);

  const retryGeneration = async (messageIndex: number) => {
    const previousAnswer = messages[messageIndex].content;
    
    let userMessageIndex = messageIndex - 1;
    while (userMessageIndex >= 0 && messages[userMessageIndex].role !== 'user') {
      userMessageIndex--;
    }
    
    if (userMessageIndex >= 0 && !isLoading) {
      const userMessage = messages[userMessageIndex];
      
      const targetMessageIndex = messageIndex;
      
      setMessages(prev => {
        const updatedMessages = [...prev];
        if (updatedMessages[targetMessageIndex]) {
          updatedMessages[targetMessageIndex] = {
            ...updatedMessages[targetMessageIndex],
            content: 'Regenerating answer...',
            references: []
          };
        }
        return updatedMessages;
      });
      
      hasReceivedData.current = false;
      currentAnswerRef.current = '';
      currentReferencesRef.current = [];
      setIsLoading(true);
      setRetryingMessageIndex(messageIndex);
      
      resetActivityTimer();
      
      try {
        console.log('Retrying question:', userMessage.content);
        
        if (eventSourceRef.current) {
          console.log('Closing existing EventSource');
          eventSourceRef.current.close();
          eventSourceRef.current = null;
        }
        const promptForRetry = `You gave me a bad answer to a question. Try it again, but paraphrase your answer, try to make it better then previous one.
** Your previous answer **
${previousAnswer}
** Question **
${userMessage.content}`;
        
        const processStreamResponse = (content: string, references: Reference[], isComplete: boolean) => {
          setMessages(prev => {
            const updatedMessages = [...prev];
            if (updatedMessages[targetMessageIndex]) {
              updatedMessages[targetMessageIndex] = {
                ...updatedMessages[targetMessageIndex],
                content: content,
                references: references,
                complete: isComplete
              };
            }
            return updatedMessages;
          });
        };
        
        const setupRetryEventListeners = (eventSource: EventSource) => {
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
                
                processStreamResponse(currentAnswerRef.current, currentReferencesRef.current, false);
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
                const answer = result.answer;
                
                processStreamResponse(answer, currentReferencesRef.current, true);
              }
              
              cleanupConnection();
            } catch (error) {
              console.error('Error parsing complete message:', error);
            }
          });

          eventSource.addEventListener('message', () => {
            resetActivityTimer();
          });

          eventSource.addEventListener('resources', () => {
            resetActivityTimer();
          });

          eventSource.addEventListener('resources', (event) => {
            resetActivityTimer();
            
            try {
              const data = JSON.parse(event.data);
              
              if (data.references && Array.isArray(data.references)) {
                currentReferencesRef.current = data.references;
                processStreamResponse(currentAnswerRef.current, data.references, false);
              }
            } catch (error) {
              console.error('Error parsing references message:', error);
            }
          });

          eventSource.addEventListener('error', (event) => {
            console.error('Error event from SSE connection:', event);
            
            if (!hasReceivedData.current) {
              processStreamResponse('Sorry, I was unable to regenerate a response. Please try again.', [], true);
            }
            
            cleanupConnection();
          });
        };

        const eventSource = await streamAnswer(promptForRetry, 5);
        eventSourceRef.current = eventSource;
        
        setupRetryEventListeners(eventSource);
      } catch (error) {
        console.error('Error setting up SSE connection for retry:', error);
        
        setMessages(prev => {
          const updatedMessages = [...prev];
          if (updatedMessages[targetMessageIndex]) {
            updatedMessages[targetMessageIndex] = {
              ...updatedMessages[targetMessageIndex],
              content: 'Error regenerating answer. Please try again.',
              complete: true
            };
          }
          return updatedMessages;
        });
        
        cleanupConnection();
      }
    }
  };
  
  const setupEventListeners = (eventSource: EventSource) => {
    eventSource.addEventListener('chunk', (event) => {
      console.log('Chunk event received:', event.data);
      resetActivityTimer();
      
      try {
        const data = JSON.parse(event.data);
        console.log('Parsed chunk data:', data);
        
        hasReceivedData.current = true;
        
        if (data.content) {
          console.log('Processing chunk content:', data.content);
          
          if (data.process_id) {
            console.log('Setting process ID:', data.process_id);
            setCurrentProcessId(data.process_id);
          }
          
          const decodedContent = decodeHtmlEntities(data.content);
          
          currentAnswerRef.current += decodedContent;
          console.log('Current answer:', currentAnswerRef.current);
          
          setMessages(prev => {
            const newMessages = [...prev];
            const lastMessage = newMessages[newMessages.length - 1];
            if (lastMessage?.role === 'assistant') {
              lastMessage.content = currentAnswerRef.current;
              lastMessage.references = currentReferencesRef.current;
              return newMessages;
            } else {
              return [...prev, { role: 'assistant', content: currentAnswerRef.current, references: currentReferencesRef.current }];
            }
          });
        }
      } catch (error) {
        console.error('Error parsing chunk message:', error);
      }
    });

    eventSource.addEventListener('complete', (event) => {
      console.log('Complete event received:', event.data);
      resetActivityTimer();
      
      try {
        const data = JSON.parse(event.data);
        console.log('Parsed complete data:', data);
        
        if (data.complete === true && data.result) {
          console.log('Stream completed with result:', data.result);
          
          const result = data.result as CompleteResult;
          const answer = result.answer;
          
          setMessages(prev => {
            const newMessages = [...prev];
            const lastMessage = newMessages[newMessages.length - 1];
            if (lastMessage?.role === 'assistant') {
              lastMessage.content = answer;
              lastMessage.references = currentReferencesRef.current;
              lastMessage.complete = true;
              return newMessages;
            } else {
              return [...prev, { role: 'assistant', content: answer, references: currentReferencesRef.current }];
            }
          });
        }
        
        cleanupConnection();
      } catch (error) {
        console.error('Error parsing complete message:', error);
      }
    });

    eventSource.addEventListener('message', () => {
      resetActivityTimer();
    });

    eventSource.addEventListener('resources', () => {
      resetActivityTimer();
    });

    eventSource.addEventListener('resources', (event) => {
      console.log('References event received:', event.data);
      resetActivityTimer();
      
      try {
        const data = JSON.parse(event.data);
        console.log('Parsed references data:', data);
        
        if (data.references && Array.isArray(data.references)) {
          currentReferencesRef.current = data.references;

          setMessages(prev => {
            const newMessages = [...prev];
            const lastMessage = newMessages[newMessages.length - 1];
            if (lastMessage?.role === 'assistant') {
              lastMessage.references = data.references;
              return newMessages;
            } else {
              return [...prev, { 
                role: 'assistant', 
                content: currentAnswerRef.current,
                references: data.references 
              }];
            }
          });
        }
      } catch (error) {
        console.error('Error parsing references message:', error);
      }
    });

    eventSource.addEventListener('error', (event) => {
      console.error('Error event from SSE connection:', event);
      
      if (!hasReceivedData.current) {
        setMessages(prev => {
          const newMessages = [...prev];
          const lastUserMessageIndex = newMessages.findIndex(m => m.role === 'user');
          
          if (lastUserMessageIndex !== -1) {
            return [
              ...newMessages.slice(0, lastUserMessageIndex + 1),
              { role: 'assistant', content: 'Sorry, I was unable to generate a response. Please try again.' }
            ];
          }
          
          return newMessages;
        });
      }
      
      cleanupConnection();
    });
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

    try {
      console.log('Starting streaming with question:', fullPrompt);

      if (eventSourceRef.current) {
        console.log('Closing existing EventSource');
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }

      const eventSource = await streamAnswer(fullPrompt, numReferences);
      eventSourceRef.current = eventSource;

      setupEventListeners(eventSource);
    } catch (error) {
      console.error('Error setting up SSE connection:', error);
      
      setMessages(prev => [
        ...prev, 
        { role: 'assistant', content: 'Sorry, there was an error connecting to the service. Please try again later.' }
      ]);
      
      cleanupConnection();
    }
  };

  const cancelGeneration = async () => {
    if (currentProcessId) {
      try {
        await cancelStream(currentProcessId);
      } catch (error) {
        console.error('Error cancelling stream:', error);
      }
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
      
      if (userMessageIndex >= 0) {
        setMessages(prev => {
          const newMessages = [...prev];
          newMessages.splice(userMessageIndex, messageIndex - userMessageIndex + 1);
          return newMessages;
        });
      } else {
        setMessages(prev => {
          const newMessages = [...prev];
          newMessages.splice(messageIndex, 1);
          return newMessages;
        });
      }
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