import React, { useState, useEffect } from 'react';
import { Resource, Reference } from '@/types/api';
import { 
    Typography, 
    Box, 
    List, 
    ListItem, 
    CircularProgress, 
    Paper, 
    Chip,
    Divider,
    useTheme,
    useMediaQuery,
    Tooltip,
    alpha,
    ListItemButton
} from '@mui/material';
import { getResource } from "@services/api.ts";
import { extractWords } from "@services/utils.ts";
import DescriptionIcon from '@mui/icons-material/Description';
import LaunchIcon from '@mui/icons-material/Launch';

interface ReferencesListProps {
    openResourceModal: (resourceId: string, textFragment?: string) => void;
    references: Reference[] | undefined;
}

export const ReferencesList: React.FC<ReferencesListProps> = ({ openResourceModal, references }) => {
    const [resourcesMap, setResourcesMap] = useState<Record<string, Resource>>({});
    const [loading, setLoading] = useState<boolean>(true);
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('sm'));

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

    const getPreviewText = (content: string, wordCount: number = 10): string => {
        return extractWords(content, wordCount) + '...';
    };

    if (loading) {
        return (
            <Box display="flex" justifyContent="center" alignItems="center" p={2} minHeight={100}>
                <CircularProgress size={28} color="primary" />
                <Typography variant="body2" color="text.secondary" ml={2}>
                    Loading references...
                </Typography>
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
        <Box mt={2} mb={3}>
            <Box 
                sx={{ 
                    display: 'flex', 
                    alignItems: 'center', 
                    mb: 1.5,
                    px: 1
                }}
            >
                <DescriptionIcon 
                    color="primary" 
                    fontSize="small" 
                    sx={{ mr: 1 }} 
                />
                <Typography 
                    variant='subtitle1' 
                    sx={{ 
                        fontWeight: 600, 
                        color: 'primary.main',
                        fontSize: isMobile ? '0.9rem' : '1rem'
                    }} 
                >
                    References
                </Typography>
                <Chip 
                    label={references.length} 
                    size="small" 
                    color="primary" 
                    sx={{ 
                        ml: 1, 
                        height: 20, 
                        '& .MuiChip-label': { 
                            px: 1, 
                            fontSize: '0.7rem' 
                        } 
                    }} 
                />
            </Box>
            
            <List sx={{ 
                padding: 0, 
                display: 'flex', 
                flexDirection: 'column', 
                gap: 1.5 
            }}>
                {Object.keys(groupedReferences).map((resourceId, resourceIndex) => {
                    const resource = resourcesMap[resourceId];
                    const refs = groupedReferences[resourceId];

                    if (!resource) return null;

                    return (
                        <Paper 
                            key={resourceId} 
                            elevation={1} 
                            sx={{ 
                                overflow: 'hidden',
                                borderRadius: 2,
                                border: `1px solid ${alpha(theme.palette.divider, 0.5)}`,
                                transition: 'all 0.2s ease-in-out',
                                '&:hover': {
                                    boxShadow: theme.shadows[2],
                                    borderColor: theme.palette.divider
                                }
                            }}
                        >
                            <ListItem 
                                sx={{ 
                                    py: 1.5,
                                    px: 2,
                                    bgcolor: alpha(theme.palette.primary.main, 0.05)
                                }}
                            >
                                <Box sx={{ 
                                    display: 'flex', 
                                    alignItems: 'center', 
                                    width: '100%', 
                                    justifyContent: 'space-between' 
                                }}>
                                    <Typography 
                                        variant="subtitle1" 
                                        sx={{ 
                                            fontWeight: 600,
                                            fontSize: isMobile ? '0.85rem' : '0.95rem',
                                            display: 'flex',
                                            alignItems: 'center'
                                        }}
                                    >
                                        <Box 
                                            component="span" 
                                            sx={{ 
                                                mr: 1.5, 
                                                bgcolor: 'primary.main', 
                                                color: 'white',
                                                borderRadius: '50%',
                                                minWidth: 22,
                                                minHeight: 22,
                                                width: 22,
                                                height: 22,
                                                display: 'flex',
                                                alignItems: 'center',
                                                justifyContent: 'center',
                                                fontSize: '0.75rem',
                                                fontWeight: 700,
                                                flexShrink: 0,
                                                aspectRatio: '1/1',
                                                textAlign: 'center',
                                                lineHeight: 1
                                            }}
                                        >
                                            {resourceIndex + 1}
                                        </Box>
                                        {resource.name}
                                    </Typography>
                                    
                                    <Tooltip title="Open resource">
                                        <Box 
                                            component="span" 
                                            onClick={() => openResourceModal(resourceId)}
                                            sx={{ 
                                                cursor: 'pointer',
                                                display: 'flex',
                                                p: 0.5,
                                                borderRadius: '50%',
                                                color: 'primary.main',
                                                transition: 'all 0.2s ease',
                                                '&:hover': { 
                                                    bgcolor: alpha(theme.palette.primary.main, 0.1)
                                                }
                                            }}
                                        >
                                            <LaunchIcon fontSize="small" />
                                        </Box>
                                    </Tooltip>
                                </Box>
                            </ListItem>
                            
                            <Divider />
                            
                            <List sx={{ py: 0.5 }}>
                                {refs.map((reference, refIndex) => (
                                    <ListItemButton
                                        key={`${resourceId}-${refIndex}`}
                                        onClick={() => openResourceModal(reference.resource_id, reference.content)}
                                        sx={{ 
                                            py: 1,
                                            px: 2,
                                            transition: 'background-color 0.2s ease',
                                            '&:hover': { 
                                                bgcolor: alpha(theme.palette.background.default, 0.5)
                                            }
                                        }}
                                    >
                                        <Box sx={{ display: 'flex', alignItems: 'flex-start' }}>
                                            <Typography 
                                                variant="body2" 
                                                color="text.secondary"
                                                sx={{ 
                                                    mr: 1.5,
                                                    minWidth: 24,
                                                    fontSize: '0.75rem'
                                                }}
                                            >
                                                {resourceIndex + 1}.{refIndex + 1}
                                            </Typography>
                                            <Typography 
                                                variant="body2"
                                                sx={{ 
                                                    color: 'text.primary',
                                                    fontSize: isMobile ? '0.8rem' : '0.875rem',
                                                    lineHeight: 1.4
                                                }}
                                            >
                                                {getPreviewText(reference.content)}
                                            </Typography>
                                        </Box>
                                    </ListItemButton>
                                ))}
                            </List>
                        </Paper>
                    );
                })}
            </List>
        </Box>
    );
};
