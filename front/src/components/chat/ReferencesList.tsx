import React, { useState, useEffect } from 'react';
import { Resource, Reference } from '@/types/api';
import { Typography, Box, Link, List, ListItem, CircularProgress } from '@mui/material';
import {getResource} from "@services/api.ts";
import {extractWords} from "@services/utils.ts";

interface ReferencesListProps {
    openResourceModal: (resourceId: string, textFragment?: string) => void;
    references: Reference[] | undefined;
}

export const ReferencesList: React.FC<ReferencesListProps> = ({ openResourceModal, references }) => {
    const [resourcesMap, setResourcesMap] = useState<Record<string, Resource>>({});
    const [loading, setLoading] = useState<boolean>(true);

    useEffect(() => {
        const fetchResources = async () => {
            if (!references || references.length === 0) {
                setLoading(false);
                return;
            }

            try {
                const resourceIds = new Set(references.map(r => r.resource_id));

                const resources = await getResources(resourceIds);

                const map: Record<string, Resource> = {};
                resources.forEach(resource => {
                    map[resource.id] = resource;
                });

                setResourcesMap(map);
            } finally {
                setLoading(false);
            }
        };

        fetchResources();
    }, [references]);

    const getResources = async (idSet: Set<string>): Promise<Resource[]> => {
        const resources: Resource[] = [];
        for (const id of idSet) {
            try {
                const doc = await getResource(id);
                resources.push(doc);
            } catch (error) {
                console.error(`Error fetching resource ${id}:`, error);
            }
        }
        return resources;
    };

    const getPreviewText = (content: string, wordCount: number = 5): string => {
        return extractWords(content, wordCount)+'...';
    };

    if (loading) {
        return (
            <Box display="flex" justifyContent="center" p={2}>
                <CircularProgress size={24} />
            </Box>
        );
    }

    if (!references || references.length === 0) {
        return null;
    }

    const groupedReferences: Record<string, Reference[]> = {};
    references.forEach(ref => {
        if (!groupedReferences[ref.resource_id]) {
            groupedReferences[ref.resource_id] = [];
        }
        groupedReferences[ref.resource_id].push(ref);
    });

    return (
        <Box>
            <Typography variant='h6' sx={{fontWeight: 'bold'}} gutterBottom>References:</Typography>
            <List sx={{ padding: 0 }}>
                {Object.keys(groupedReferences).map((resourceId, resourceIndex) => {
                    const resource = resourcesMap[resourceId];
                    const refs = groupedReferences[resourceId];

                    if (!resource) return null;

                    return (
                        <Box key={resourceId} mb={2}>
                            <ListItem sx={{ fontWeight: 'bold', py: 0 }}>
                                <Typography variant="subtitle1">
                                    <Link
                                    component="button"
                                    onClick={() => openResourceModal(resourceId)}
                                    sx={{
                                        cursor: 'pointer',
                                        fontWeight: 'bold',
                                        textDecoration: 'none',
                                        '&:hover': { textDecoration: 'underline' }
                                    }}
                                >
                                        {resourceIndex + 1}.
                                        {resource.name}
                                </Link>
                                </Typography>
                            </ListItem>
                            <List sx={{ pl: 4, mt: 0}}>
                                {refs.map((reference, refIndex) => (
                                    <ListItem
                                        key={`${resourceId}-${refIndex}`}
                                        sx={{ py: 0 }}>
                                        <Typography variant="body2">
                                            <Link
                                            component="button"
                                            onClick={() => openResourceModal(reference.resource_id, reference.content)}
                                            sx={{
                                                cursor: 'pointer',
                                                textDecoration: 'none',
                                                '&:hover': { textDecoration: 'underline' }
                                            }}
                                        >
                                                {refIndex + 1}.
                                                {getPreviewText(reference.content)}
                                        </Link>
                                        </Typography>
                                    </ListItem>
                                ))}
                            </List>
                        </Box>
                    );
                })}
            </List>
        </Box>
    );
};
