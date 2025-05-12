import { useRef, useEffect } from 'react';
import { Box, Paper, useTheme, useMediaQuery, Chip, alpha } from '@mui/material';
import { FormatMessage } from '@/components';
import { Message } from '@/types/api';
import {SaveMessageAsResourceButton} from "@components/chat/SaveMessageAsResourceButton.tsx";
import {RetryAskButton} from "@components/chat/RetryAskButton.tsx";
import {DeleteMessageButton} from "@components/chat/DeleteMessageButton.tsx";
import {ClearChatButton} from "@components/chat/ClearChatButton.tsx";
import DoneAllIcon from '@mui/icons-material/DoneAll';

interface ChatMessagesProps {
  messages: Message[];
  openResourceModal: (resourceId: string) => void;
  isMobile?: boolean;
}

export function ChatMessages({ messages, openResourceModal, isMobile = false }: ChatMessagesProps) {
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const theme = useTheme();
  const isTablet = useMediaQuery(theme.breakpoints.down('md'));
  
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
        p: isMobile ? 1 : isTablet ? 2 : 3,
        display: 'flex', 
        flexDirection: 'column',
        alignItems: 'center'
      }}
    >
      {messages.length > 0 && (
        <Box 
          sx={{ 
            width: '100%',
            display: isMobile ? 'none' : 'flex', 
            justifyContent: 'flex-start',
            mb: 2,
            position: 'sticky',
            top: 8,
            zIndex: 10,
            pl: isMobile ? 1 : 2
          }}
        >
          <ClearChatButton />
        </Box>
      )}
      <Box
        sx={{
          width: '100%',
          maxWidth: isMobile ? '100%' : isTablet ? '85%' : '50%',
          display: 'flex',
          flexDirection: 'column',
          gap: isMobile ? 1.5 : 2,
          pb: isMobile ? 10 : isTablet ? 15 : 20
        }}
      >
        {messages.map((message, index) => (
          <Box
            key={index}
            sx={{
                width: '100%',
                display: 'flex',
                flexDirection: 'column',
                ...(message.role === 'user' && {
                  alignItems: 'flex-end',
                }),
            }}
          >
            {message.role === 'user' ? (
              // User message with bubble
              <Paper
                elevation={0}
                sx={{
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'end',
                  maxWidth: isMobile ? '85%' : '75%',
                  px: isMobile ? 1.5 : 2,
                  py: isMobile ? 1 : 1.5,
                  borderRadius: isMobile ? 1.5 : 2,
                  backgroundColor: 'primary.main',
                  color: 'white',
                  overflowWrap: 'break-word',
                  wordBreak: 'break-word',
                  fontSize: isMobile ? '0.9rem' : 'inherit',
                  boxShadow: 'none',
                  '& a': {
                    color: 'white',
                    textDecoration: 'underline',
                    '&:hover': {
                      textDecoration: 'none'
                    }
                  }
                }}
              >
                <FormatMessage message={message} openResourceModal={openResourceModal} />
              </Paper>
            ) : (
              // Assistant message without bubble, full width
              <Box 
                sx={{
                  width: '100%',
                  display: 'flex',
                  flexDirection: 'column',
                  pt: 1,
                  pb: 2,
                  mb: 0.5,
                  borderBottom: index < messages.length - 1 && messages[index + 1]?.role === 'user' ? 
                    `1px solid ${theme.palette.divider}` : 'none',
                  position: 'relative'
                }}
              >
                <FormatMessage message={message} openResourceModal={openResourceModal} />
                
                <Box sx={{ 
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  mt: 2,
                  width: '100%'
                }}>
                  {/* End of message indicator */}
                  {message.complete && (
                    <Chip
                      icon={<DoneAllIcon fontSize="small" />}
                      label="Response complete"
                      size="small"
                      sx={{
                        height: 24,
                        bgcolor: alpha(theme.palette.primary.main, 0.08),
                        color: 'text.secondary',
                        fontWeight: 500,
                        fontSize: '0.7rem',
                        borderRadius: '12px',
                        '& .MuiChip-icon': {
                          color: theme.palette.success.main,
                          fontSize: '0.9rem',
                        },
                        boxShadow: `0 0 0 1px ${alpha(theme.palette.divider, 0.5)}`,
                      }}
                    />
                  )}
                  
                  {message.role === 'assistant' && message.content && message.references && message.complete && (
                    <Box sx={{ 
                      alignSelf: 'flex-end',
                      display: 'flex',
                      gap: 1,
                    }}>
                      <DeleteMessageButton messageIndex={index} />
                      <RetryAskButton messageIndex={index} />
                      <SaveMessageAsResourceButton message={message}/>
                    </Box>
                  )}
                </Box>
              </Box>
            )}
          </Box>
        ))}
        <div ref={messagesEndRef} />
      </Box>
    </Box>
  );
} 