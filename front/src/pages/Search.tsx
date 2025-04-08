import { useResourceModal } from '../hooks';
import { ChatMessages, ChatInput, ResourceModalView } from '../components';
import { useChat } from '../context';
import { Box } from '@mui/material';

export default function SearchPage() {
  const { messages, isLoading, submitQuestion, cancelGeneration } = useChat();
  const { 
    isResourceModalOpen, 
    selectedResource, 
    isLoadingResource, 
    openResourceModal, 
    closeResourceModal 
  } = useResourceModal();

  return (
    <Box 
      sx={{ 
        display: 'flex', 
        flexDirection: 'column',
        position: 'relative',
        height: '100%',
        overflow: 'hidden'
      }}
    >
      {/* Messages container - takes full available height and scrolls */}
      <Box sx={{ flex: 1, overflow: 'hidden', position: 'relative' }}>
        <ChatMessages 
          messages={messages} 
          openResourceModal={openResourceModal} 
        />
      </Box>
      
      {/* Input is positioned on top of the messages */}
      <Box sx={{ position: 'absolute', bottom: 0, left: 0, right: 0, zIndex: 100 }}>
        <ChatInput 
          onSubmit={submitQuestion} 
          onCancel={cancelGeneration} 
          isLoading={isLoading} 
        />
      </Box>
      
      <ResourceModalView 
        isOpen={isResourceModalOpen} 
        onClose={closeResourceModal} 
        resource={selectedResource} 
        isLoading={isLoadingResource} 
      />
    </Box>
  );
} 