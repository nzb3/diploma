import { useState, useEffect } from 'react';
import { getResources, saveResource, deleteResource, getResource } from '../services/api';
import { Resource, ResourceEvent, SaveDocumentRequest } from '../types/api';
import { 
  ResourceList, 
  ResourceModal, 
  ResourceUploadForm 
} from '../components/resources';
import { Box } from '@mui/material';

export default function ResourcesPage() {
  // Resource state
  const [resources, setResources] = useState<Resource[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  
  // Upload errors tracking
  const [uploadErrors, setUploadErrors] = useState<Record<string, string>>({});
  
  // Modal state
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedResource, setSelectedResource] = useState<Resource | null>(null);
  const [isLoadingResource, setIsLoadingResource] = useState(false);

  // Load resources on component mount
  useEffect(() => {
    loadResources();
  }, []);

  // Load resources from the API
  const loadResources = async () => {
    if (isLoading) return;
    
    setIsLoading(true);
    try {
      const data = await getResources();
      setResources(data);
    } catch (error) {
      console.error('Failed to load resources:', error);
    } finally {
      setIsLoading(false);
    }
  };

  // Handle resource upload submission
  const handleUploadSubmit = async (data: SaveDocumentRequest) => {
    console.log("Uploading resource with data:", data);
    setIsUploading(true);
    setUploadErrors({});

    try {
      const eventSource = await saveResource(data);

      eventSource.onmessage = (event) => {
        const data = JSON.parse(event.data);
        
        if (data.status_update) {
          const update = data.status_update as ResourceEvent;

          // Update the resource status instead of progress
          setResources(prevResources => {
            return prevResources.map(resource => {
              if (resource.id === update.id) {
                return { ...resource, status: update.status };
              }
              return resource;
            });
          });

          if (update.error) {
            setUploadErrors(prev => ({
              ...prev,
              [update.id]: update.error || 'Unknown error',
            }));
            loadResources();
          }
        } else if (data.completed) {
          setIsUploading(false);
          loadResources();
        }
      };

      eventSource.onerror = () => {
        setIsUploading(false);
        loadResources();
      };
    } catch (error) {
      setIsUploading(false);
      console.error('Failed to upload resource:', error);
    }
  };

  // Handle resource deletion
  const handleDeleteResource = async (id: string) => {
    try {
      await deleteResource(id);
      setResources(resources.filter(r => r.id !== id));
    } catch (error) {
      console.error('Failed to delete resource:', error);
    }
  };

  // Handle resource click to open modal
  const handleResourceClick = async (resource: Resource) => {
    console.log('Resource clicked:', resource);
    setSelectedResource(resource);
    setIsModalOpen(true);
    console.log('Modal state set to open');
    
    // If we don't have the full resource content, fetch it
    if (!resource.extracted_content && !resource.raw_content) {
      setIsLoadingResource(true);
      try {
        console.log('Fetching resource details...');
        const fullResource = await getResource(resource.id);
        console.log('Resource details fetched:', fullResource);
        setSelectedResource(fullResource);
      } catch (error) {
        console.error('Failed to load resource details:', error);
      } finally {
        setIsLoadingResource(false);
      }
    }
  };

  // Handle modal close
  const handleCloseModal = () => {
    setIsModalOpen(false);
    setSelectedResource(null);
  };

  return (
    <Box sx={{ display: 'flex', height: '100%' }}>
      {/* Sidebar with ResourceUploadForm */}
     
      <ResourceUploadForm 
        onUpload={handleUploadSubmit}
        isUploading={isUploading}
      />
      
      
      {/* Main content area with ResourceList */}
      <Box 
        sx={{ 
          flexGrow: 1, 
          p: 3, 
          overflowY: 'auto', 
          display: 'flex', 
          justifyContent: 'center'
        }}
      >
        <Box sx={{ width: '100%', maxWidth: 'lg' }}>
          <ResourceList 
            resources={resources}
            isLoading={isLoading}
            uploadErrors={uploadErrors}
            onResourceClick={handleResourceClick}
            onDeleteResource={handleDeleteResource}
            onRefreshResources={loadResources}
          />
        </Box>
      </Box>
      
      <ResourceModal 
        isOpen={isModalOpen}
        resource={selectedResource}
        isLoading={isLoadingResource}
        onClose={handleCloseModal}
      />
    </Box>
  );
} 