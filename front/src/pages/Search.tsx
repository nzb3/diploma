import { useResourceModal } from '@/hooks';
import { ChatMessages, ChatInput } from '@/components';
import { ResourceModal } from "@components";
import { useChat } from '@/context';
import { Box, useTheme, useMediaQuery } from '@mui/material';
import { useEffect, useRef } from 'react';

export default function SearchPage() {
  const { messages, isLoading, submitQuestion, cancelGeneration } = useChat();
  const {
    isResourceModalOpen,
    selectedResource, 
    isLoadingResource, 
    openResourceModal, 
    closeResourceModal 
  } = useResourceModal();
  
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const isTablet = useMediaQuery(theme.breakpoints.down('md'));
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [messages]);

  return (
    <Box 
      sx={{ 
        display: 'flex', 
        flexDirection: 'column',
        position: 'relative',
        height: '100%',
        overflow: 'hidden',
        px: isMobile ? 0.5 : isTablet ? 1 : 2,
      }}
    >
      <Box 
        ref={containerRef}
        sx={{ 
          flex: 1, 
          overflow: 'auto', 
          position: 'relative',
          pb: isMobile ? '130px' : isTablet ? '120px' : '100px', 
          scrollBehavior: 'smooth',
        }}
      >
        <ChatMessages 
          messages={messages} 
          openResourceModal={openResourceModal} 
          isMobile={isMobile}
        />
      </Box>
      
      <Box sx={{ 
        position: 'absolute', 
        bottom: 0, 
        left: 0, 
        right: 0, 
        zIndex: 100,
        px: isMobile ? 0.5 : isTablet ? 1 : 2,
        pb: isMobile ? 0.5 : isTablet ? 1 : 2,
      }}>
        <ChatInput 
          onSubmit={submitQuestion} 
          onCancel={cancelGeneration} 
          isLoading={isLoading}
          isMobile={isMobile}
        />
      </Box>
      
      <ResourceModal
        isOpen={isResourceModalOpen}
        onClose={closeResourceModal} 
        resource={selectedResource} 
        isLoading={isLoadingResource} 
      />
    </Box>
  );
} 