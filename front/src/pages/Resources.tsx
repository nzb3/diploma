import { useEffect } from 'react';
import { Resource,  } from '@/types';
import {
  ResourceList,
  ResourceModal,
  ResourceUploadForm
} from '@components';
import { Box, useTheme, useMediaQuery } from '@mui/material';
import {useResourceModal, useResourceManagement} from "@/hooks";

export default function ResourcesPage() {
  const {
    resources,
    isLoading,
    isUploading,
    uploadErrors,
    loadResources,
    uploadResource,
    deleteResourceById,
  } = useResourceManagement();

  const {
    isResourceModalOpen,
    selectedResource,
    isLoadingResource,
    openResourceModal,
    closeResourceModal,
    isEditable,
  } = useResourceModal();

  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

  useEffect(() => {
    loadResources();

    const interval = setInterval(() => {
      loadResources();
    }, 60000);
    return () => {
      clearInterval(interval);
    }
  }, []);

  const handleResourceClick = async (resource: Resource) => {
    console.log('Resource clicked:', resource);
    if (resource.id) {
      await openResourceModal(resource.id)
    }
  };

  const handleCloseResourceModal = async () => {
    await loadResources();
    closeResourceModal();
  }

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
            isOpen={isResourceModalOpen}
            resource={selectedResource}
            isLoading={isLoadingResource}
            onClose={handleCloseResourceModal}
            isEditable={isEditable}
        />
      </Box>
  );
}
