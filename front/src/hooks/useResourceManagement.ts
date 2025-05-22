import { useState, useCallback } from 'react';
import {
    getResources,
    saveResource,
    deleteResource as apiDeleteResource,
    updateResource as apiUpdateResource,
} from '@/services/api';
import {Resource, ResourceEvent, SaveDocumentRequest, UpdateResourceRequest} from '@/types';

interface UseResourceManagementResult {
    resources: Resource[];
    isLoading: boolean;
    isUploading: boolean;
    uploadErrors: Record<string, string>;
    loadResources: () => Promise<void>;
    uploadResource: (data: SaveDocumentRequest) => Promise<void>;
    deleteResourceById: (id: string) => Promise<void>;
    updateResourceById: (data: UpdateResourceRequest) => Promise<Resource>;
    refreshResources: () => Promise<void>;
}

export function useResourceManagement(): UseResourceManagementResult {
    const [resources, setResources] = useState<Resource[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [isUploading, setIsUploading] = useState(false);
    const [uploadErrors, setUploadErrors] = useState<Record<string, string>>({});

    const loadResources = useCallback(async () => {
        if (isLoading) return;
        setIsLoading(true);
        try {
            const data = await getResources();
            setResources(Array.isArray(data) ? data : []);
        } catch (error) {
            console.error('Failed to load resources:', error);
            setResources([]);
        } finally {
            setIsLoading(false);
        }
    }, [isLoading]);

    const uploadResource = useCallback(async (data: SaveDocumentRequest) => {
        setIsUploading(true);
        setUploadErrors({});
        try {
            const eventSource = await saveResource(data);

            eventSource.onmessage = (event) => {
                try {
                    const msg = JSON.parse(event.data);
                    if (msg.status_update) {
                        const update = msg.status_update as ResourceEvent;
                        setResources(prev =>
                            Array.isArray(prev)
                                ? prev.map(resource =>
                                    resource.id === update.id
                                        ? { ...resource, status: update.status }
                                        : resource
                                )
                                : []
                        );
                        if (update.error) {
                            setUploadErrors(prev => ({
                                ...prev,
                                [update.id]: update.error || 'Unknown error',
                            }));
                            loadResources();
                        }
                    } else if (msg.completed) {
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
    }, [loadResources]);

    const deleteResourceById = useCallback(async (id: string) => {
        try {
            await apiDeleteResource(id);
            setResources(prev =>
                Array.isArray(prev) ? prev.filter(r => r.id !== id) : []
            );
        } catch (error) {
            console.error('Failed to delete resource:', error);
        }
    }, []);

    const updateResourceById = useCallback(async (data: UpdateResourceRequest)=> {
        try {
           const resource = await apiUpdateResource(data);
           setResources(prev => {
               if (Array.isArray(prev)) {
                   const resourceIndex = prev.findIndex(value => resource.id === value.id);
                   if (resourceIndex !== -1) {
                       const updatedResources = [...prev];
                       updatedResources[resourceIndex] = resource;
                       return updatedResources;
                   }
               }
               return prev;
           });
           return resource;
        } catch (error) {
            console.error('Failed to update resource:', error);
            throw error;
        }
    }, [])

    const refreshResources = loadResources;

    return {
        resources,
        isLoading,
        isUploading,
        uploadErrors,
        loadResources,
        uploadResource,
        deleteResourceById,
        refreshResources,
        updateResourceById,
    };
}
