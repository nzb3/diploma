import { useState, useRef, useCallback } from 'react';
import { SaveDocumentRequest } from '../../types/api';

interface ResourceUploadFormProps {
  onUpload: (data: SaveDocumentRequest) => Promise<void>;
  isUploading: boolean;
}

export function ResourceUploadForm({ onUpload, isUploading }: ResourceUploadFormProps) {
  // State for form values
  const [name, setName] = useState('');
  const [type, setType] = useState('txt');
  const [content, setContent] = useState('');
  const [url, setUrl] = useState('');
  const [fileName, setFileName] = useState('');
  
  // State for upload mode (file or url)
  const [uploadMode, setUploadMode] = useState<'url' | 'file'>('url');
  
  // State for drag and drop
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const dropZoneRef = useRef<HTMLDivElement>(null);

  // Base64 encode function for file content
  const safeBase64Encode = (str: string): string => {
    const encoder = new TextEncoder();
    const utf8Bytes = encoder.encode(str);
    
    // Convert byte array to binary string
    const binaryString = String.fromCharCode(...utf8Bytes);
    
    return btoa(binaryString);
  };

  // Reset form
  const resetForm = () => {
    setName('');
    setContent('');
    setUrl('');
    setFileName('');
    setType('txt');
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  // Detect file type from extension
  const detectFileType = (fileName: string): string => {
    const extension = fileName.split('.').pop()?.toLowerCase();
    
    if (extension === 'pdf') {
      return 'pdf';
    } else if (extension === 'md' || extension === 'markdown') {
      return 'markdown';
    } else if (extension === 'txt') {
      return 'txt';
    } else {
      // Return the actual extension for unsupported types
      return extension || 'unknown';
    }
  };

  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (uploadMode === 'url') {
      if (!name || !url) {
        alert('Please enter both name and URL');
        return;
      }
      
      await onUpload({
        name,
        type: 'url',
        content: '',
        url: url
      });
    } else {
      if (!name || !content) {
        alert('Please enter a name and select a file');
        return;
      }
      
      await onUpload({
        name,
        type,
        content: type === 'pdf' ? content : safeBase64Encode(content),
        url: undefined
      });
    }

    // Reset form after submission
    resetForm();
  };

  // Process file
  const processFile = (file: File) => {
    // Set file name
    setFileName(file.name);
    
    // Detect file type
    const detectedType = detectFileType(file.name);
    
    // Check if file type is supported
    const supportedTypes = ['pdf', 'markdown', 'txt'];
    if (!supportedTypes.includes(detectedType)) {
      alert(`Unsupported file type: ${detectedType}. Only PDF, Markdown, and Text files are supported.`);
      // Clear the file input
      setFileName('');
      setContent('');
      return;
    }
    
    // Set type
    setType(detectedType);
    
    // Set name from filename if empty
    if (!name) {
      const namePart = file.name.split('.')[0];
      setName(namePart);
    }
    
    // Read file content based on type
    if (detectedType === 'pdf') {
      // Handle PDF as binary
      const reader = new FileReader();
      reader.onload = (event) => {
        if (event.target?.result) {
          // Extract the base64 part after the data URL prefix
          const base64Content = (event.target.result as string).split(',')[1];
          setContent(base64Content);
        }
      };
      // Add error handling for large files
      reader.onerror = () => {
        alert('Error processing file. Please try again.');
      };
      reader.readAsDataURL(file);
    } else {
      // Read as text (for markdown and text files)
      const reader = new FileReader();
      reader.onload = (event) => {
        if (event.target?.result) {
          setContent(event.target.result as string);
        }
      };
      // Add error handling for large files
      reader.onerror = () => {
        alert('Error processing file. Please try again.');
      };
      reader.readAsText(file);
    }
  };

  // Handle file drop
  const handleDrop = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    
    const files = Array.from(e.dataTransfer.files);
    if (files.length === 0) return;
    
    const file = files[0];
    processFile(file);
  }, [name]);

  // Handle drag events
  const handleDragOver = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  }, []);

  // Handle file input change
  const handleFileChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);
    if (files.length === 0) return;
    
    const file = files[0];
    processFile(file);
  }, [name]);

  // Switch to file mode and focus the drop zone
  const switchToFileMode = () => {
    setUploadMode('file');
    // Focus on the drop zone (for accessibility)
    setTimeout(() => {
      if (dropZoneRef.current) {
        dropZoneRef.current.focus();
      }
    }, 0);
  };

  return (
    <div className="bg-white shadow sm:rounded-lg">
      <div className="px-4 py-5 sm:p-6">
        <h3 className="text-lg font-medium leading-6 text-gray-900">Upload Resource</h3>
        
        {/* Upload Mode Switcher */}
        <div className="mt-4 flex space-x-1 rounded-md border border-gray-300 p-1 w-fit">
          <button
            type="button"
            onClick={() => setUploadMode('url')}
            className={`px-3 py-1.5 text-sm font-medium rounded-md ${
              uploadMode === 'url' 
                ? 'bg-blue-600 text-white' 
                : 'text-gray-700 hover:bg-gray-100'
            }`}
          >
            URL
          </button>
          <button
            type="button"
            onClick={switchToFileMode}
            className={`px-3 py-1.5 text-sm font-medium rounded-md ${
              uploadMode === 'file' 
                ? 'bg-blue-600 text-white' 
                : 'text-gray-700 hover:bg-gray-100'
            }`}
          >
            File
          </button>
        </div>
        
        <form onSubmit={handleSubmit} className="mt-5 space-y-4">
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-700">
              Name
            </label>
            <input
              type="text"
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
              required
            />
          </div>

          {uploadMode === 'url' ? (
            <div>
              <label htmlFor="url" className="block text-sm font-medium text-gray-700">
                URL
              </label>
              <input
                type="url"
                id="url"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                required={uploadMode === 'url'}
                placeholder="https://example.com/document.pdf"
              />
            </div>
          ) : (
            <>
              <div
                ref={dropZoneRef}
                onDragOver={handleDragOver}
                onDragLeave={handleDragLeave}
                onDrop={handleDrop}
                onClick={() => fileInputRef.current?.click()}
                tabIndex={0}
                className={`border-2 border-dashed rounded-md p-8 flex flex-col items-center justify-center cursor-pointer focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                  isDragging ? 'border-blue-500 bg-blue-50' : 'border-gray-300 hover:border-gray-400'
                }`}
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-12 w-12 text-gray-400"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"
                  />
                </svg>
                <p className="mt-2 text-base text-gray-600">
                  Drag and drop your file here
                </p>
                <p className="mt-1 text-sm text-gray-500">
                  or <span className="text-blue-600 font-medium">browse files</span>
                </p>
                <p className="mt-2 text-xs text-gray-500">
                  PDF, Markdown, or Text files (up to 100MB)
                </p>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".pdf,.md,.markdown,.txt,text/plain,application/pdf,text/markdown"
                  className="hidden"
                  onChange={handleFileChange}
                />
              </div>

              {fileName && (
                <div className={`mt-4 p-3 rounded-md flex items-center ${
                  type === 'pdf' 
                    ? 'bg-red-50 border border-red-200' 
                    : type === 'markdown' 
                      ? 'bg-green-50 border border-green-200'
                      : 'bg-blue-50 border border-blue-200'
                }`}>
                  <div className={`p-2 rounded-md mr-3 ${
                    type === 'pdf' 
                      ? 'bg-red-100 text-red-600' 
                      : type === 'markdown' 
                        ? 'bg-green-100 text-green-600'
                        : 'bg-blue-100 text-blue-600'
                  }`}>
                    {type === 'pdf' ? (
                      <svg className="h-6 w-6" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M4 4a2 2 0 012-2h4.586A2 2 0 0112 2.586L15.414 6A2 2 0 0116 7.414V16a2 2 0 01-2 2H6a2 2 0 01-2-2V4z" clipRule="evenodd" />
                      </svg>
                    ) : type === 'markdown' ? (
                      <svg className="h-6 w-6" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M4 4a2 2 0 012-2h8a2 2 0 012 2v12a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm3 1h6v4H7V5zm8 8v2h1v1H4v-1h1v-2H4v-1h16v1h-1zm-2 2v-2H7v2h6z" clipRule="evenodd" />
                      </svg>
                    ) : (
                      <svg className="h-6 w-6" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M4 4a2 2 0 012-2h8a2 2 0 012 2v12a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm3 1h6v4H7V5zm8 8v2h1v1H4v-1h1v-2H4v-1h16v1h-1zm-2 2v-2H7v2h6z" clipRule="evenodd" />
                      </svg>
                    )}
                  </div>
                  <div>
                    <p className="text-sm font-medium">{fileName}</p>
                    <p className="text-xs text-gray-500">
                      {type === 'pdf' 
                        ? 'PDF file ready to upload' 
                        : type === 'markdown'
                          ? 'Markdown file ready to upload'
                          : 'Text file ready to upload'}
                    </p>
                  </div>
                  <button 
                    type="button"
                    onClick={resetForm}
                    className="ml-auto text-gray-500 hover:text-gray-700"
                    title="Remove file"
                  >
                    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              )}

              {content && type !== 'pdf' && (
                <div className="mt-2">
                  <label htmlFor="content" className="block text-sm font-medium text-gray-700">
                    Content Preview
                  </label>
                  <textarea
                    id="content"
                    value={content}
                    onChange={(e) => setContent(e.target.value)}
                    rows={4}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500 sm:text-sm"
                  />
                </div>
              )}
            </>
          )}

          <button
            type="submit"
            disabled={isUploading || (uploadMode === 'url' ? !url : !content) || !name}
            className="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isUploading ? 'Uploading...' : 'Upload'}
          </button>
        </form>
      </div>
    </div>
  );
} 