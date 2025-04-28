import { useState, useEffect } from 'react';
import { getResource } from '../services/api';
import { Resource,  } from '../types/api';
import {
  ResourceList,
  ResourceModal,
  ResourceUploadForm
} from '@components';
import { Box, useTheme, useMediaQuery } from '@mui/material';
import {useResourceManagement} from "@/hooks/useResourceManagement.ts";

export default function ResourcesPage() {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedResource, setSelectedResource] = useState<Resource | null>(null);
  const [isLoadingResource, setIsLoadingResource] = useState(false);

  const {
    resources,
    isLoading,
    isUploading,
    uploadErrors,
    loadResources,
    uploadResource,
    deleteResourceById,
  } = useResourceManagement();

  // Use theme breakpoints to detect mobile layout
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  // Load resources on component mount
  useEffect(() => {
    loadResources();
  }, []);

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
            onUpload={uploadResource}
            isUploading={isUploading}
        />

        <ResourceList
            resources={resources}
            isLoading={isLoading}
            uploadErrors={uploadErrors}
            onResourceClick={handleResourceClick}
            onDeleteResource={deleteResourceById}
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
