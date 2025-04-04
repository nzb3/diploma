import { useResourceModal } from '../hooks';
import { ChatMessages, ChatInput, ResourceModalView } from '../components';
import { useChat } from '../context';

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
    <div className="flex flex-col h-[calc(100vh-8rem)]">
      <ChatMessages messages={messages} openResourceModal={openResourceModal} />
      <ChatInput 
        onSubmit={submitQuestion} 
        onCancel={cancelGeneration} 
        isLoading={isLoading} 
      />
      <ResourceModalView 
        isOpen={isResourceModalOpen} 
        onClose={closeResourceModal} 
        resource={selectedResource} 
        isLoading={isLoadingResource} 
      />
    </div>
  );
} 