import { IconButton, useMediaQuery, useTheme } from "@mui/material";
import { useChat } from "@/context/ChatContext";
import DeleteIcon from '@mui/icons-material/Delete';

interface DeleteMessageButtonProps {
    messageIndex: number;
}

export const DeleteMessageButton = ({ messageIndex }: DeleteMessageButtonProps) => {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down("sm"));
    const { deleteMessage } = useChat();

    const handleButtonClick = () => {
        deleteMessage(messageIndex);
    };

    return (
        <IconButton
            onClick={handleButtonClick}
            size={isMobile ? "small" : "medium"}
            color="error"
            sx={{
                backgroundColor: 'rgba(211, 47, 47, 0.1)',
                borderRadius: '50%',
                p: isMobile ? 0.5 : 0.8,
                boxShadow: 1,
                '&:hover': {
                    backgroundColor: 'rgba(211, 47, 47, 0.2)',
                    boxShadow: 2,
                },
                mr: 1
            }}
            aria-label="Delete message"
        >
            <DeleteIcon fontSize={isMobile ? "small" : "medium"} />
        </IconButton>
    );
} 