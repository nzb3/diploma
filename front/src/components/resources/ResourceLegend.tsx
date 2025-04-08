import {
  Box,
  Paper,
  Typography,
  Chip,
  Stack
} from '@mui/material';
import PictureAsPdfIcon from '@mui/icons-material/PictureAsPdf';
import TextSnippetIcon from '@mui/icons-material/TextSnippet';
import MarkdownIcon from '@mui/icons-material/Code';
import LinkIcon from '@mui/icons-material/Link';

export function ResourceLegend() {
  return (
    <Paper elevation={1} sx={{ p: 2, mb: 2 }}>
      <Typography variant="subtitle2" gutterBottom sx={{ fontWeight: 'medium' }}>
        Resource Legend
      </Typography>
      
      <Stack direction={{ xs: 'column', md: 'row' }} spacing={3}>
        <Box>
          <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1 }}>
            Status Types:
          </Typography>
          <Box display="flex" gap={1} flexWrap="wrap">
            <Chip
              size="small"
              label="completed"
              sx={{
                borderRadius: 1,
                backgroundColor: 'rgba(76, 175, 80, 0.1)',
                color: '#2E7D32',
                fontWeight: 500,
                '& .MuiChip-label': { px: 1.5 }
              }}
            />
            <Chip
              size="small"
              label="processing"
              sx={{
                borderRadius: 1,
                backgroundColor: 'rgba(255, 152, 0, 0.1)',
                color: '#E65100',
                fontWeight: 500,
                '& .MuiChip-label': { px: 1.5 }
              }}
            />
            <Chip
              size="small"
              label="failed"
              sx={{
                borderRadius: 1,
                backgroundColor: 'rgba(244, 67, 54, 0.1)',
                color: '#C62828',
                fontWeight: 500,
                '& .MuiChip-label': { px: 1.5 }
              }}
            />
          </Box>
        </Box>
        
        <Box>
          <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 1 }}>
            Resource Types:
          </Typography>
          <Box display="flex" gap={2} flexWrap="wrap">
            <Box display="flex" alignItems="center" gap={0.5}>
              <PictureAsPdfIcon fontSize="small" sx={{ color: '#f44336' }} />
              <Typography variant="body2">PDF</Typography>
            </Box>
            <Box display="flex" alignItems="center" gap={0.5}>
              <MarkdownIcon fontSize="small" sx={{ color: '#4caf50' }} />
              <Typography variant="body2">Markdown</Typography>
            </Box>
            <Box display="flex" alignItems="center" gap={0.5}>
              <TextSnippetIcon fontSize="small" sx={{ color: '#2196f3' }} />
              <Typography variant="body2">Text</Typography>
            </Box>
            <Box display="flex" alignItems="center" gap={0.5}>
              <LinkIcon fontSize="small" sx={{ color: '#9c27b0' }} />
              <Typography variant="body2">URL</Typography>
            </Box>
          </Box>
        </Box>
      </Stack>
    </Paper>
  );
} 