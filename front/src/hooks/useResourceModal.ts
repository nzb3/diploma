import { useState } from 'react';
import { getResource } from '../services/api';
import { Resource } from '../types/api';

interface UseResourceModalResult {
  isResourceModalOpen: boolean;
  selectedResource: Resource | null;
  isLoadingResource: boolean;
  openResourceModal: (resourceId: string) => Promise<void>;
  closeResourceModal: () => void;
}

export function useResourceModal(): UseResourceModalResult {
  const [isResourceModalOpen, setIsResourceModalOpen] = useState(false);
  const [selectedResource, setSelectedResource] = useState<Resource | null>(null);
  const [isLoadingResource, setIsLoadingResource] = useState(false);

  const openResourceModal = async (resourceId: string) => {
    setIsResourceModalOpen(true);
    
    try {
      setIsLoadingResource(true);
      const resource = await getResource(resourceId);
      setSelectedResource(resource);
    } catch (error) {
      console.error('Failed to load resource details:', error);
    } finally {
      setIsLoadingResource(false);
    }
  };

  const closeResourceModal = () => {
    setIsResourceModalOpen(false);
    setSelectedResource(null);
  };

  return {
    isResourceModalOpen,
    selectedResource,
    isLoadingResource,
    openResourceModal,
    closeResourceModal
  };
} 