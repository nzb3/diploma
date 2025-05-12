import { Resource } from '@/types/api';
import {
    ListItem,
    ListItemText,
    Box,
    Button,
    Typography,
    Tooltip,
    Paper,
    useMediaQuery,
    useTheme
} from '@mui/material';
import PictureAsPdfIcon from '@mui/icons-material/PictureAsPdf';
import TextSnippetIcon from '@mui/icons-material/TextSnippet';
import MarkdownIcon from '@mui/icons-material/Code';
import LinkIcon from '@mui/icons-material/Link';
import DeleteIcon from '@mui/icons-material/Delete';
import { getStatusDescription } from '@services/utils';
import { ReactNode } from "react";

interface ResourceListItemProps {
    resource: Resource;
    onResourceClick: (resource: Resource) => void;
    onDeleteResource: (resourceId: string) => void;
    uploadError?: string;
    renderStatusChip?: (status: string) => ReactNode;
}

export function ResourceListItem({
                                     resource,
                                     onResourceClick,
                                     onDeleteResource,
                                     uploadError,
                                     renderStatusChip
                                 }: ResourceListItemProps) {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down('lg'));

    const getTypeIcon = () => {
        switch (resource.type) {
            case 'pdf':
                return <PictureAsPdfIcon fontSize="small" sx={{ color: '#f44336' }} />;
            case 'markdown':
                return <MarkdownIcon fontSize="small" sx={{ color: '#4caf50' }} />;
            case 'txt':
                return <TextSnippetIcon fontSize="small" sx={{ color: '#2196f3' }} />;
            case 'url':
                return <LinkIcon fontSize="small" sx={{ color: '#9c27b0' }} />;
            default:
                return <TextSnippetIcon fontSize="small" sx={{ color: '#757575' }} />;
        }
    };

    return (
        <Paper
            elevation={1}
            sx={{
                marginTop: '24px',
                mb: 1,
                border: '1px solid ' + theme.palette.divider,
                borderRadius: 2,
                overflow: 'hidden',
                transition: 'transform 0.2s ease-in-out, box-shadow 0.2s ease-in-out',
                '&:hover': {
                    border: '1px solid ' + theme.palette.divider,
                    transform: 'translateY(-4px)',
                    boxShadow: '0 6px 24px rgba(0, 0, 0, 0.15)',
                },
                width: '100%',
                maxWidth: '100%'
            }}
        >
            <ListItem
                sx={{
                    py: 2,
                    px: { xs: 1.5, sm: 2 },
                    cursor: 'pointer',
                    position: 'relative',
                    transition: 'background-color 0.2s',
                    flexDirection: isMobile ? 'column' : 'row',
                    alignItems: isMobile ? 'flex-start' : 'center',
                    gap: isMobile ? 2 : 0
                }}
            >
                <Box
                    onClick={() => onResourceClick(resource)}
                    sx={{
                        display: 'flex',
                        alignItems: 'center',
                        width: '100%',
                        gap: 2,
                        flexShrink: 1,
                        overflow: 'hidden'
                    }}
                >
                    <Box sx={{ minWidth: 24, display: 'flex' }}>
                        {getTypeIcon()}
                    </Box>
                    <ListItemText
                        primary={
                            <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                <Typography noWrap sx={{ maxWidth: { xs: '180px', sm: '250px', md: '100%' } }}>
                                    {resource.name}
                                </Typography>
                            </Box>
                        }
                        secondary={
                            <Tooltip title={`Resource type: ${resource.type}`}>
                                <Typography variant="body2" component="span">
                                    {resource.type}
                                </Typography>
                            </Tooltip>
                        }
                    />
                </Box>

                <Box
                    display="flex"
                    alignItems="center"
                    gap={1.5}
                    sx={{
                        width: isMobile ? '100%' : 'auto',
                        justifyContent: isMobile ? 'space-between' : 'flex-end',
                        mt: isMobile ? 1 : 0,
                        flexWrap: { xs: 'wrap', sm: 'nowrap' }
                    }}
                >
                    {resource.status && renderStatusChip && (
                        <Tooltip title={getStatusDescription(resource.status)}>
                            <Box sx={{ flexShrink: 0 }}>
                                {renderStatusChip(resource.status)}
                            </Box>
                        </Tooltip>
                    )}

                    {uploadError && (
                        <Typography
                            variant="caption"
                            color="error"
                            sx={{
                                display: 'block',
                                width: isMobile ? '100%' : 'auto',
                                mb: isMobile ? 1 : 0
                            }}
                        >
                            {uploadError}
                        </Typography>
                    )}
                    <Button
                        size="small"
                        variant="contained"
                        color="error"
                        onClick={(e) => {
                            e.stopPropagation();
                            onDeleteResource(resource.id);
                        }}
                        startIcon={<DeleteIcon fontSize="small" />}
                        sx={{
                            minWidth: 90,
                            px: 2,
                            py: 0.75,
                            borderRadius: 1.5,
                            boxShadow: 2,
                            flexShrink: 0,
                            '&:hover': {
                                bgcolor: 'error.dark',
                                boxShadow: 3,
                            }
                        }}
                    >
                        Delete
                    </Button>
                </Box>
            </ListItem>
        </Paper>
    );
}
