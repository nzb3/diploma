import React, { useState } from 'react';
import Button from './Button';

function URLForm({ onUpload }) {
    const [url, setUrl] = useState('');

    const handleSubmit = (e) => {
        e.preventDefault();
        if (url.trim()) {
            onUpload(url);
            setUrl('');
        }
    };

    const handleKeyPress = (e) => {
        if (e.key === 'Enter') {
            handleSubmit(e);
        }
    };

    return (
        <div className="upload-area url-upload">
            <h3>Upload Website</h3>
            <input
                type="url"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder="Enter website URL (e.g., https://example.com)"
            />
            <Button onClick={handleSubmit}>Upload</Button>
        </div>
    );
}

export default URLForm;
