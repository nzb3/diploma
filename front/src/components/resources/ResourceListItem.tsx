import { Resource } from '../../types/api';

interface ResourceListItemProps {
  resource: Resource;
  onResourceClick: (resource: Resource) => void;
  onDeleteResource: (resourceId: string) => void;
  uploadError?: string;
}

export function ResourceListItem({ 
  resource, 
  onResourceClick, 
  onDeleteResource, 
  uploadError 
}: ResourceListItemProps) {
  return (
    <li 
      className="py-4 hover:bg-gray-50 cursor-pointer"
      onClick={(e) => {
        // Make sure the click isn't on the delete button
        if (!(e.target as HTMLElement).closest('button')) {
          onResourceClick(resource);
        }
      }}
    >
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-900">{resource.name}</p>
          <p className="text-sm text-gray-500">{resource.type}</p>
        </div>
        <div className="flex items-center space-x-4">
          {resource.status && (
            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
              resource.status === 'completed' || resource.status === 'processed'
                ? 'bg-green-100 text-green-800' 
                : resource.status === 'failed'
                ? 'bg-red-100 text-red-800'
                : 'bg-yellow-100 text-yellow-800'
            }`}>
              {resource.status}
            </span>
          )}
          
          {uploadError && (
            <p className="text-sm text-red-600">{uploadError}</p>
          )}
          <button
            onClick={(e) => {
              e.stopPropagation();
              onDeleteResource(resource.id);
            }}
            className="text-red-600 hover:text-red-900"
          >
            Delete
          </button>
        </div>
      </div>
    </li>
  );
} 