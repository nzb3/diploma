import { createContext, useContext, useState, useRef, ReactNode, useEffect, useCallback } from 'react';
import { streamAnswer, cancelStream } from '../services/api';
import { Message, CompleteResult, Reference } from '../types/api';

// For handling HTML preservation
const decodeHtmlEntities = (html: string): string => {
  const textArea = document.createElement('textarea');
  textArea.innerHTML = html;
  return textArea.value;
};

// Configuration
const SSE_TIMEOUT = 10000; // 10 seconds timeout for SSE connections without activity

interface ChatContextType {
  messages: Message[];
  isLoading: boolean;
  submitQuestion: (question: string) => Promise<void>;
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
  const eventSourceRef = useRef<EventSource | null>(null);
  const lastActivityTimestampRef = useRef<number>(Date.now());
  const timeoutIdRef = useRef<number | null>(null);

  // Helper function to clean up connection resources
  const cleanupConnection = useCallback(() => {
    // Clear any existing activity timeout
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

  // Resets the inactivity timer
  const resetActivityTimer = useCallback(() => {
    lastActivityTimestampRef.current = Date.now();
    
    // Clear any existing timeout
    if (timeoutIdRef.current !== null) {
      window.clearTimeout(timeoutIdRef.current);
    }
    
    // Set a new timeout to check for inactivity
    timeoutIdRef.current = window.setTimeout(() => {
      const timeElapsed = Date.now() - lastActivityTimestampRef.current;
      
      if (timeElapsed >= SSE_TIMEOUT) {
        console.warn('SSE connection inactive for too long, closing connection');
        
        // Only close if we're still loading and have an active EventSource
        if (isLoading && eventSourceRef.current) {
          console.log('Closing inactive SSE connection');
          cleanupConnection();
        }
      }
    }, SSE_TIMEOUT);
  }, [isLoading, cleanupConnection]);

  const submitQuestion = async (question: string) => {
    if (!question.trim() || isLoading) return;

    // Reset the data received flag and current answer
    hasReceivedData.current = false;
    currentAnswerRef.current = '';
    
    setMessages(prev => [...prev, { role: 'user', content: question, }]);
    setIsLoading(true);

    // Initialize the activity timer
    resetActivityTimer();

    try {
      console.log('Starting streaming with question:', question);

      // Close any existing EventSource
      if (eventSourceRef.current) {
        console.log('Closing existing EventSource');
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }

      // Create new EventSource - now awaiting the promise since streamAnswer is async
      const eventSource = await streamAnswer(question);
      eventSourceRef.current = eventSource;

      // Listen for chunk events
      eventSource.addEventListener('chunk', (event) => {
        console.log('Chunk event received:', event.data);
        // Reset activity timer with each incoming chunk
        resetActivityTimer();
        
        try {
          const data = JSON.parse(event.data);
          console.log('Parsed chunk data:', data);
          
          // Mark that we've received at least one message
          hasReceivedData.current = true;
          
          if (data.content) {
            console.log('Processing chunk content:', data.content);
            
            if (data.process_id) {
              console.log('Setting process ID:', data.process_id);
              setCurrentProcessId(data.process_id);
            }
            
            // Process HTML content properly
            const decodedContent = decodeHtmlEntities(data.content);
            
            // Update the current answer with new content
            currentAnswerRef.current += decodedContent;
            console.log('Current answer:', currentAnswerRef.current);
            
            // Update the assistant message
            setMessages(prev => {
              const newMessages = [...prev];
              const lastMessage = newMessages[newMessages.length - 1];
              if (lastMessage?.role === 'assistant') {
                lastMessage.content = currentAnswerRef.current;
                return newMessages;
              } else {
                return [...prev, { role: 'assistant', content: currentAnswerRef.current }];
              }
            });
          }
        } catch (error) {
          console.error('Error parsing chunk message:', error);
        }
      });

      // Listen for complete events
      eventSource.addEventListener('complete', (event) => {
        console.log('Complete event received:', event.data);
        // Reset activity timer when complete event is received
        resetActivityTimer();
        
        try {
          const data = JSON.parse(event.data);
          console.log('Parsed complete data:', data);
          
          // Handle successful completion
          if (data.complete === true && data.result) {
            console.log('Stream completed with result:', data.result);
            
            const result = data.result as CompleteResult;
            let formattedAnswer = result.answer;
            
            // Add references if they exist
            if (result.references && result.references.length > 0) {
              formattedAnswer += '\n\nReferences:';
              
              result.references.forEach((ref: Reference, index: number) => {
                if (ref.resource_id) {
                  formattedAnswer += `\n[${index + 1}] Resource: ${ref.resource_id}`;
                }
              });
            }
            
            // Decode HTML entities in the final answer
            formattedAnswer = decodeHtmlEntities(formattedAnswer);
            
            // Update the assistant message
            setMessages(prev => {
              const newMessages = [...prev];
              const lastMessage = newMessages[newMessages.length - 1];
              if (lastMessage?.role === 'assistant') {
                lastMessage.content = formattedAnswer;
                return newMessages;
              } else {
                return [...prev, { role: 'assistant', content: formattedAnswer }];
              }
            });
          }
          
          // Clean up the connection
          cleanupConnection();
        } catch (error) {
          console.error('Error parsing complete message:', error);
        }
      });

      // Reset activity timer for any event
      eventSource.addEventListener('message', () => {
        resetActivityTimer();
      });

      // Listen for resources events
      eventSource.addEventListener('resources', () => {
        resetActivityTimer();
      });

      // Listen for references events
      eventSource.addEventListener('references', (event) => {
        console.log('References event received:', event.data);
        resetActivityTimer();
        
        try {
          const data = JSON.parse(event.data);
          console.log('Parsed references data:', data);
          
          if (data.references && Array.isArray(data.references)) {
            // Update the assistant message with references
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

      // Listen for error events
      eventSource.addEventListener('error', (event) => {
        console.error('Error event from SSE connection:', event);
        
        // Check if we ever received any data
        if (!hasReceivedData.current) {
          // If no data was received, add an error message
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
        
        // Clean up the connection
        cleanupConnection();
      });
    } catch (error) {
      console.error('Error setting up SSE connection:', error);
      
      // Add error message for failed connection setup
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

  // Clean up resources on unmount
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