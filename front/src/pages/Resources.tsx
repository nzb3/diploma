import { useState, useEffect } from 'react';
import { getResources, saveResource, deleteResource, getResource } from '../services/api';
import { Resource, ResourceEvent, SaveDocumentRequest } from '../types/api';
import {
  ResourceList,
  ResourceModal,
  ResourceUploadForm
} from '@components/resources';
import { Box, useTheme, useMediaQuery } from '@mui/material';

export default function ResourcesPage() {
  const [resources, setResources] = useState<Resource[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadErrors, setUploadErrors] = useState<Record<string, string>>({});
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedResource, setSelectedResource] = useState<Resource | null>(null);
  const [isLoadingResource, setIsLoadingResource] = useState(false);

  // Use theme breakpoints to detect mobile layout
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

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
      // Ensure data is an array
      setResources(Array.isArray(data) ? data : []);
    } catch (error) {
      console.error('Failed to load resources:', error);
      // Set empty array on error
      setResources([]);
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
        try {
          const data = JSON.parse(event.data);

          if (data.status_update) {
            const update = data.status_update as ResourceEvent;

            // Update the resource status instead of progress
            setResources(prevResources => {
              if (!Array.isArray(prevResources)) return [];

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
        } catch (error) {
          console.error('Error processing SSE message:', error);
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
      setResources(prevResources =>
          Array.isArray(prevResources)
              ? prevResources.filter(r => r.id !== id)
              : []
      );
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
      <Box sx={{
        display: 'flex',
        flexDirection: isMobile ? 'column' : 'row',
        height: '100vh',
        overflow: 'hidden'
      }}>
        <ResourceUploadForm
            onUpload={handleUploadSubmit}
            isUploading={isUploading}
        />

        <ResourceList
            resources={resources}
            isLoading={isLoading}
            uploadErrors={uploadErrors}
            onResourceClick={handleResourceClick}
            onDeleteResource={handleDeleteResource}
            onRefreshResources={loadResources}
        />

        <ResourceModal
            isOpen={isModalOpen}
            resource={selectedResource}
            isLoading={isLoadingResource}
            onClose={handleCloseModal}
        />
      </Box>
  );
}
