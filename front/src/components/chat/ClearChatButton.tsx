import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, useMediaQuery, useTheme, ListItemIcon, ListItemText, Box } from "@mui/material";
import { useState } from "react";
import { useChat } from "@/context/ChatContext";
import DeleteSweepIcon from '@mui/icons-material/DeleteSweep';

interface ClearChatButtonProps {
    isMenuItem?: boolean;
    onClose?: () => void;
}

export const ClearChatButton = ({ isMenuItem = false, onClose }: ClearChatButtonProps) => {
    const theme = useTheme();
    const isMobile = useMediaQuery(theme.breakpoints.down("sm"));
    const { clearChat, messages } = useChat();
    const [open, setOpen] = useState(false);

    const handleClickOpen = () => {
        setOpen(true);
    };

    const handleClose = () => {
        setOpen(false);
    };

    const handleConfirm = () => {
        clearChat();
        setOpen(false);
        if (onClose) {
            onClose();
        }
    };
    
    if (messages.length === 0) {
        return null;
    }

    return (
        <>
            {isMenuItem ? (
                <Box 
                    onClick={handleClickOpen} 
                    sx={{ 
                        display: 'flex', 
                        alignItems: 'center', 
                        width: '100%',
                        cursor: 'pointer',
                        px: 1,
                        py: 0.5,
                        '&:hover': {
                            backgroundColor: 'rgba(211, 47, 47, 0.1)',
                        }
                    }}
                >
                    <ListItemIcon>
                        <DeleteSweepIcon fontSize="small" color="error" />
                    </ListItemIcon>
                    <ListItemText 
                        primary="Clear chat"
                    />
                </Box>
            ) : (
                <Button
                    variant="outlined"
                    color="error"
                    startIcon={<DeleteSweepIcon />}
                    size={isMobile ? "small" : "medium"}
                    onClick={handleClickOpen}
                    sx={{
                        borderRadius: theme.shape.borderRadius,
                        textTransform: 'none',
                        fontWeight: 500,
                        fontSize: isMobile ? '0.75rem' : '0.875rem',
                        py: isMobile ? 0.5 : 0.75,
                        px: isMobile ? 1.5 : 2,
                        boxShadow: 1,
                        borderColor: theme.palette.error.main,
                        '&:hover': {
                            backgroundColor: 'rgba(211, 47, 47, 0.1)',
                            borderColor: theme.palette.error.dark,
                        }
                    }}
                >
                    Clear Chat
                </Button>
            )}

            <Dialog
                open={open}
                onClose={handleClose}
                aria-labelledby="alert-dialog-title"
                aria-describedby="alert-dialog-description"
            >
                <DialogTitle id="alert-dialog-title">
                    {"Clear all chat messages?"}
                </DialogTitle>
                <DialogContent>
                    <DialogContentText id="alert-dialog-description">
                        This will delete all messages from the current chat. This action cannot be undone.
                    </DialogContentText>
                </DialogContent>
                <DialogActions>
                    <Button onClick={handleClose} color="primary">Cancel</Button>
                    <Button onClick={handleConfirm} color="error" autoFocus>
                        Clear
                    </Button>
                </DialogActions>
            </Dialog>
        </>
    );
} 