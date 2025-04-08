export const getStatusDescription = (status: string) => {
    switch (status) {
        case 'completed':
            return 'Saved and processed';
          case 'failed':
            return 'Processing failed';
          case 'processing':
            return 'Processing';
          default:
            return 'Status unknown';
    }
};