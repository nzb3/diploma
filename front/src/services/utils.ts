export const extractWords = (content: string, numOfWords: number = 6): string => {
    return content.trim().split(" ").slice(0, numOfWords).join(" ")
}

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

export const safeBase64Encode = (str: string): string => {
    const encoder = new TextEncoder();
    const utf8Bytes = encoder.encode(str);
    const binaryString = String.fromCharCode(...utf8Bytes);
    return btoa(binaryString);
};