import { IconButton, useMediaQuery, useTheme } from "@mui/material";
import { useChat } from "@/context/ChatContext";
import ReplayIcon from '@mui/icons-material/Replay';

interface RetryAskButtonProps {
    messageIndex: number;
}

export const RetryAskButton = ({ messageIndex }: RetryAskButtonProps) => {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down("sm"));
    const { retryGeneration } = useChat();

    const handleButtonClick = async () => {
        await retryGeneration(messageIndex);
    };

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
                },
                mr: 1
            }}
            aria-label="Retry generation"
        >
        <ReplayIcon fontSize={isMobile ? "small" : "medium"} />
        </IconButton>
    );
} 