import { useState, useRef } from 'react';
import { streamAnswer, cancelStream } from '../services/api';
import { Message, CompleteResult, Reference } from '../types/api';

interface UseStreamingAnswerResult {
  isLoading: boolean;
  handleSubmitQuestion: (question: string) => Promise<void>;
  handleCancel: () => Promise<void>;
  messages: Message[];
}

export function useStreamingAnswer(): UseStreamingAnswerResult {
  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [currentProcessId, setCurrentProcessId] = useState<string | null>(null);
  const hasReceivedData = useRef<boolean>(false);
  const currentAnswerRef = useRef<string>('');

  const handleSubmitQuestion = async (question: string) => {
    if (!question.trim() || isLoading) return;

    // Reset the data received flag and current answer
    hasReceivedData.current = false;
    currentAnswerRef.current = '';
    
    setMessages(prev => [...prev, { role: 'user', content: question }]);
    setIsLoading(true);
    
    // Create a reference to the event source that we can close later
    let eventSource: EventSource | null = null;

    try {
      console.log('Starting streaming with question:', question);
      
      // Add an initial empty assistant message that we'll update
      setMessages(prev => [...prev, { role: 'assistant', content: '' }]);
      
      eventSource = await streamAnswer(question);

      // Listen for chunk events
      eventSource.addEventListener('chunk', (event) => {
        console.log('Chunk event received:', event.data);
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
            
            // Update the current answer with new content
            currentAnswerRef.current += data.content;
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
          
          setIsLoading(false);
          setCurrentProcessId(null);
          if (eventSource) {
            eventSource.close();
            eventSource = null;
          }
        } catch (error) {
          console.error('Error parsing complete message:', error);
        }
      });

      // Listen for error events
      eventSource.addEventListener('error', (error) => {
        console.error('Error event received:', error);
        
        setIsLoading(false);
        setCurrentProcessId(null);
        if (eventSource) {
          eventSource.close();
          eventSource = null;
        }
        
        // Only show the error message if we haven't received any data yet
        if (!hasReceivedData.current) {
          // Update the last message with the error
          setMessages(prev => {
            const newMessages = [...prev];
            const lastMessage = newMessages[newMessages.length - 1];
            if (lastMessage?.role === 'assistant') {
              lastMessage.content = 'Sorry, we have some issues. Please try again later';
              return newMessages;
            } else {
              return [...prev, { role: 'assistant', content: 'Sorry, we have some issues. Please try again later' }];
            }
          });
        }
      });

    } catch (error) {
      console.error('Failed to initialize streaming:', error);
      setIsLoading(false);
      setCurrentProcessId(null);
      if (eventSource) {
        eventSource.close();
      }
      
      // Create or update the assistant message
      setMessages(prev => {
        const newMessages = [...prev];
        const lastMessage = newMessages[newMessages.length - 1];
        if (lastMessage?.role === 'assistant') {
          lastMessage.content = 'Sorry, we have some issues. Please try again later';
          return newMessages;
        } else {
          return [...prev, { role: 'assistant', content: 'Sorry, we have some issues. Please try again later' }];
        }
      });
    }
  };

  const handleCancel = async () => {
    if (currentProcessId) {
      try {
        console.log('Attempting to cancel stream with process ID:', currentProcessId);
        await cancelStream(currentProcessId);
        console.log('Stream cancellation request sent');
      } catch (error) {
        console.error('Error canceling stream:', error);
      } finally {
        setIsLoading(false);
        setCurrentProcessId(null);
      }
    } else {
      console.log('No process ID available to cancel');
      setIsLoading(false);
    }
  };

  return {
    isLoading,
    handleSubmitQuestion,
    handleCancel,
    messages
  };
} 