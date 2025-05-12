import {CircularProgress, IconButton, useMediaQuery, useTheme} from "@mui/material";
import {Message} from "@/types/api";
import {useResourceManagement} from "@/hooks/useResourceManagement.ts";
import {extractWords, safeBase64Encode} from "@services/utils.ts";
import SaveIcon from '@mui/icons-material/Save';

interface saveMessageAsResourceButtonProps {
    message: Message,
}

export const SaveMessageAsResourceButton = ({message}: saveMessageAsResourceButtonProps) => {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down("sm"));
    const {
        isUploading,
        uploadResource,
    } = useResourceManagement();


    const handleButtonClick = async () => {
        const name = extractWords(message.content)
        await uploadResource({
            name: name,
            type: 'text',
            content: safeBase64Encode(message.content),
        });
    }

    return (
        <IconButton
            onClick={handleButtonClick}
            size={isMobile ? "small" : "medium"}
            color="primary"
            sx={{
                backgroundColor: 'rgba(25, 118, 210, 0.1)',
                borderRadius: '50%',
                p: isMobile ? 0.5 : 0.8,
                boxShadow: 1,
                '&:hover': {
                    backgroundColor: 'rgba(25, 118, 210, 0.2)',
                    boxShadow: 2,
                }
            }}
            disabled={isUploading}
            aria-label="Save as resource"
        >
            {isUploading ? (
                <CircularProgress size={isMobile ? 16 : 20} sx={{ color: 'primary.main' }} />
            ) : (
                <SaveIcon fontSize={isMobile ? "small" : "medium"} />
            )}
        </IconButton>
    );
}