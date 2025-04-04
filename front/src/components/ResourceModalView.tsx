import { Dialog } from '@headlessui/react';
import { Resource } from '../types/api';

interface ResourceModalViewProps {
  isOpen: boolean;
  onClose: () => void;
  resource: Resource | null;
  isLoading: boolean;
}

export function ResourceModalView({ isOpen, onClose, resource, isLoading }: ResourceModalViewProps) {
  return (
    <Dialog open={isOpen} onClose={onClose} className="relative z-50">
      <div className="fixed inset-0 bg-black/30" aria-hidden="true" />
      
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <Dialog.Panel className="mx-auto max-w-3xl rounded bg-white p-6 shadow-xl">
          <Dialog.Title className="text-lg font-medium leading-6 text-gray-900">
            Resource Details
          </Dialog.Title>
          
          {isLoading ? (
            <div className="mt-4">
              <p>Loading resource...</p>
            </div>
          ) : resource ? (
            <div className="mt-4">
              <h3 className="font-medium">{resource.name}</h3>
              <p className="text-sm text-gray-500">Type: {resource.type}</p>
              <p className="text-sm text-gray-500">Created: {new Date(resource.created_at).toLocaleString()}</p>
              
              <div className="mt-4 border p-4 rounded-md bg-gray-50 max-h-96 overflow-auto">
                {resource.extracted_content ? (
                  <pre className="whitespace-pre-wrap text-sm">{resource.extracted_content}</pre>
                ) : resource.raw_content ? (
                  <pre className="whitespace-pre-wrap text-sm">{resource.raw_content}</pre>
                ) : (
                  <p className="text-sm text-gray-500">No content available</p>
                )}
              </div>
            </div>
          ) : (
            <div className="mt-4">
              <p className="text-red-500">Resource not found or failed to load</p>
            </div>
          )}
          
          <div className="mt-6 flex justify-end">
            <button
              type="button"
              className="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
              onClick={onClose}
            >
              Close
            </button>
          </div>
        </Dialog.Panel>
      </div>
    </Dialog>
  );
} 