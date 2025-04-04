import { Dialog } from '@headlessui/react';
import { Resource } from '../../types/api';

interface ResourceModalProps {
  isOpen: boolean;
  resource: Resource | null;
  isLoading: boolean;
  onClose: () => void;
}

export function ResourceModal({
  isOpen,
  resource,
  isLoading,
  onClose,
}: ResourceModalProps) {
  // Safely decode base64 content
  const decodeContent = (content: string | undefined): string => {
    if (!content) return '';
    try {
      return atob(content);
    } catch (e) {
      console.error('Failed to decode base64 content:', e);
      return 'Unable to decode content';
    }
  };

  return (
    <Dialog
      open={isOpen}
      onClose={onClose}
      className="relative z-10"
    >
      <div className="fixed inset-0 bg-black bg-opacity-25" />

      <div className="fixed inset-0 overflow-y-auto">
        <div className="flex min-h-full items-center justify-center p-4 text-center">
          <Dialog.Panel className="w-full max-w-2xl transform overflow-hidden rounded-2xl bg-white p-6 text-left align-middle shadow-xl transition-all">
            {resource && (
              <>
                <div className="text-lg font-medium leading-6 text-gray-900 flex justify-between items-center">
                  <div>
                    <span>{resource.name}</span>
                    <span className={`ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      resource.status === 'completed' || resource.status === 'processed' 
                        ? 'bg-green-100 text-green-800' 
                        : resource.status === 'failed'
                        ? 'bg-red-100 text-red-800'
                        : 'bg-yellow-100 text-yellow-800'
                    }`}>
                      {resource.status}
                    </span>
                  </div>
                  <button
                    type="button"
                    className="rounded-md bg-white text-gray-400 hover:text-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    onClick={onClose}
                  >
                    <span className="sr-only">Close</span>
                    <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
                
                <div className="mt-2">
                  <div className="flex space-x-2 text-sm text-gray-500 mb-2">
                    <span>Type: {resource.type}</span>
                    <span>•</span>
                    <span>Created: {new Date(resource.created_at).toLocaleString()}</span>
                  </div>
                  
                  {isLoading ? (
                    <div className="flex justify-center py-8">
                      <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-500"></div>
                    </div>
                  ) : (
                    <div className="border border-gray-200 rounded-md p-4 max-h-96 overflow-auto">
                      {resource.type === 'url' ? (
                        <div>
                          <a 
                            href={resource.extracted_content || ''} 
                            target="_blank" 
                            rel="noopener noreferrer"
                            className="text-blue-600 hover:underline break-all"
                          >
                            {resource.extracted_content || 'No URL available'}
                          </a>
                        </div>
                      ) : (
                        <pre className="whitespace-pre-wrap text-sm font-mono">
                          {resource.extracted_content || 
                          (resource.raw_content ? decodeContent(resource.raw_content) : 'No content available')}
                        </pre>
                      )}
                    </div>
                  )}
                </div>

                <div className="mt-4 flex justify-end">
                  <button
                    type="button"
                    className="inline-flex justify-center rounded-md border border-transparent bg-blue-100 px-4 py-2 text-sm font-medium text-blue-900 hover:bg-blue-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 focus-visible:ring-offset-2"
                    onClick={onClose}
                  >
                    Close
                  </button>
                </div>
              </>
            )}
          </Dialog.Panel>
        </div>
      </div>
    </Dialog>
  );
} 