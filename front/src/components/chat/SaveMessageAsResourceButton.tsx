import {Button, CircularProgress, Typography, useMediaQuery, useTheme} from "@mui/material";
import {Message} from "../../types/api";
import {useResourceManagement} from "@/hooks/useResourceManagement.ts";
import {extractWords, safeBase64Encode} from "@services/utils.ts";

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
        <Button
            onClick={handleButtonClick}
            size="small"
            variant="contained"
            sx={{
                minWidth: 20,
                borderRadius: 1.5,
                boxShadow: 2,
                flexShrink: 0,
                '&:hover': {
                    boxShadow: 3,
                }
            }}
            disabled={isUploading}
        >
            {isUploading ? (
                <>
                    <CircularProgress size={isMobile ? 16 : 20} sx={{ color: 'white' }} />
                    <Typography variant="button" sx={{
                        color: 'white',
                    }}>
                        Saving...
                    </Typography>
                </>
            ) : (
                <Typography variant='button' sx={{
                    color: 'white',
                }}>
                    Save answer
                </Typography>
            )}
        </Button>
    );
}