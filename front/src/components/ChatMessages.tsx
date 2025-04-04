import { useRef, useEffect } from 'react';
import { FormatMessageContent } from './FormatMessageContent';
import { Message } from '../types/api';

interface ChatMessagesProps {
  messages: Message[];
  openResourceModal: (resourceId: string) => void;
}

export function ChatMessages({ messages, openResourceModal }: ChatMessagesProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  return (
    <div className="flex-1 overflow-y-auto p-4 space-y-4">
      {messages.map((message, index) => (
        <div
          key={index}
          className={`flex ${
            message.role === 'user' ? 'justify-end' : 'justify-start'
          }`}
        >
          <div
            className={`max-w-3xl rounded-lg px-4 py-2 ${
              message.role === 'user'
                ? 'bg-blue-600 text-white'
                : 'bg-white text-gray-900 shadow'
            }`}
          >
            <FormatMessageContent content={message.content} openResourceModal={openResourceModal} />
          </div>
        </div>
      ))}
      <div ref={messagesEndRef} />
    </div>
  );
} 