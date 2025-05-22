import { useState } from 'react';
import { getResource } from '@/services/api';
import { Resource } from '@/types';

interface UseResourceModalResult {
  isResourceModalOpen: boolean;
  selectedResource: Resource | null;
  isLoadingResource: boolean;
  openResourceModal: (resourceId: string, edit?:boolean) => Promise<void>;
  closeResourceModal: () => void;
  isEditable: boolean;
}

export function useResourceModal(): UseResourceModalResult {
  const [isResourceModalOpen, setIsResourceModalOpen] = useState(false);
  const [selectedResource, setSelectedResource] = useState<Resource | null>(null);
  const [isLoadingResource, setIsLoadingResource] = useState(false);
  const [isEditable, setIsEditable] = useState(false);

  const openResourceModal = async (resourceId: string, edit?:boolean) => {
    setIsResourceModalOpen(true);
    if (edit !== undefined && edit) {
      setIsEditable(true);
    }
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
    closeResourceModal,
    isEditable,
  };
} 