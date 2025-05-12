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
  submitQuestion: (question: string, numReferences?: number) => Promise<void>;
  cancelGeneration: () => Promise<void>;
  decodeHtmlEntities: (text: string) => string;
}

const ChatContext = createContext<ChatContextType>({
  messages: [],
  isLoading: false,
  submitQuestion: async () => {},
  cancelGeneration: async () => {},
  decodeHtmlEntities: (text) => text,
});

export const useChat = () => useContext(ChatContext);

interface ChatProviderProps {
  children: ReactNode;
}

export function ChatProvider({ children }: ChatProviderProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [currentProcessId, setCurrentProcessId] = useState<string | null>(null);
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

  const submitQuestion = async (question: string, numReferences: number = 5) => {
    if (!question.trim() || isLoading) return;

    hasReceivedData.current = false;
    currentAnswerRef.current = '';
    
    setMessages(prev => [...prev, { role: 'user', content: question, }]);
    setIsLoading(true);

    resetActivityTimer();

    try {
      console.log('Starting streaming with question:', question);

      if (eventSourceRef.current) {
        console.log('Closing existing EventSource');
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }

      const eventSource = await streamAnswer(question, numReferences);
      eventSourceRef.current = eventSource;

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
    }}>
      {children}
    </ChatContext.Provider>
  );
} 