import { Resource } from '../../types/api';
import { ResourceListItem } from './ResourceListItem';

interface ResourceListProps {
  resources: Resource[];
  isLoading: boolean;
  uploadErrors: Record<string, string>;
  onResourceClick: (resource: Resource) => void;
  onDeleteResource: (resourceId: string) => void;
  onRefreshResources: () => void;
}

export function ResourceList({
  resources,
  isLoading,
  uploadErrors,
  onResourceClick,
  onDeleteResource,
  onRefreshResources,
}: ResourceListProps) {
  return (
    <div className="bg-white shadow sm:rounded-lg">
      <div className="px-4 py-5 sm:p-6">
        <div className="flex justify-between items-center">
          <h3 className="text-lg font-medium leading-6 text-gray-900">Resources</h3>
          <div className="flex items-center space-x-2">
            {isLoading && (
              <div className="animate-pulse flex space-x-1">
                <div className="h-2 w-2 bg-blue-500 rounded-full"></div>
                <div className="h-2 w-2 bg-blue-500 rounded-full"></div>
                <div className="h-2 w-2 bg-blue-500 rounded-full"></div>
              </div>
            )}
            <button
              onClick={onRefreshResources}
              disabled={isLoading}
              className="p-1 rounded-full hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
              title="Refresh resources"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
            </button>
          </div>
        </div>
        <div className="mt-5">
          {resources.length === 0 ? (
            <div className="text-center py-10 text-gray-500">
              No resources found. Upload a resource to get started.
            </div>
          ) : (
            <ul className="divide-y divide-gray-200">
              {resources.map((resource) => (
                <ResourceListItem
                  key={resource.id}
                  resource={resource}
                  onResourceClick={onResourceClick}
                  onDeleteResource={onDeleteResource}
                  uploadError={uploadErrors[resource.id]}
                />
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
} 