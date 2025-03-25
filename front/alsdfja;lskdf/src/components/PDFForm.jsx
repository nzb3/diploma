import React, { useRef } from 'react';
import Button from './Button';

function PDFForm({ onUpload }) {
    const fileInputRef = useRef(null);

    const handleSubmit = (e) => {
        e.preventDefault();
        const file = fileInputRef.current.files[0];
        if (file) {
            onUpload(file);
            fileInputRef.current.value = '';
        }
    };

    return (
        <div className="upload-area pdf-upload">
            <h3>Upload PDF Document</h3>
            <input
                type="file"
                ref={fileInputRef}
                accept=".pdf"
            />
            <Button onClick={handleSubmit}>Upload</Button>
        </div>
    );
}

export default PDFForm;
