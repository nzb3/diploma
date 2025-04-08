import { useRef, useEffect } from 'react';
import { Box, Paper } from '@mui/material';
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
    <Box 
      sx={{ 
        flex: 1,
        height: '100%',
        overflowY: 'auto', 
        p: 3,
        display: 'flex', 
        flexDirection: 'column',
        alignItems: 'center'
      }}
    >
      <Box
        sx={{
          width: '100%',
          maxWidth: '50%',
          display: 'flex',
          flexDirection: 'column',
          gap: 2,
          pb: 20
        }}
      >
        {messages.map((message, index) => (
          <Box
            key={index}
            sx={{
              display: 'flex',
              justifyContent: message.role === 'user' ? 'flex-end' : 'flex-start',
              width: '100%'
            }}
          >
            <Paper
              elevation={message.role === 'user' ? 0 : 1}
              sx={{
                maxWidth: '75%',
                px: 2,
                py: 1.5,
                borderRadius: 2,
                backgroundColor: message.role === 'user' 
                  ? 'primary.main' 
                  : 'background.paper',
                color: message.role === 'user' 
                  ? 'white' 
                  : 'text.primary',
                overflowWrap: 'break-word',
                wordBreak: 'break-word',
                '& a': {
                  color: message.role === 'user' ? 'white' : 'primary.main',
                  textDecoration: 'underline',
                  '&:hover': {
                    textDecoration: 'none'
                  }
                }
              }}
            >
              <FormatMessageContent content={message.content} openResourceModal={openResourceModal} />
            </Paper>
          </Box>
        ))}
        <div ref={messagesEndRef} />
      </Box>
    </Box>
  );
} 