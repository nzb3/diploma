import axios from 'axios';
import {AskRequest, AskResponse, Resource, SaveDocumentRequest} from '../types/api';
import authService from './authService';

const API_BASE_URL = '/api/v1';

const api = axios.create({
  baseURL: API_BASE_URL,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  }
});

// Add interceptor to include auth token in requests
api.interceptors.request.use(
  async (config) => {
    if (authService.isAuthenticated()) {
      const token = authService.getToken();
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Add interceptor to handle 401/403 responses
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    
    // If the error is unauthorized and we haven't tried to refresh the token yet
    if ((error.response?.status === 401 || error.response?.status === 403) && 
        !originalRequest._retry && 
        authService.isAuthenticated()) {
      
      originalRequest._retry = true;
      
      try {
        // Try to refresh the token
        const refreshed = await authService.updateToken(10);
        
        if (refreshed) {
          // If token refresh was successful, retry the original request
          const token = authService.getToken();
          if (token) {
            originalRequest.headers.Authorization = `Bearer ${token}`;
          }
          return axios(originalRequest);
        }
      } catch (refreshError) {
        console.error('Failed to refresh token', refreshError);
        // If refresh fails, redirect to login
        authService.login();
      }
    }
    
    return Promise.reject(error);
  }
);

export const getResources = async (): Promise<Resource[]> => {
  const response = await api.get('/resources');
  return response.data;
};

export const getResource = async (id: string): Promise<Resource> => {
  const response = await api.get(`/resources/${id}`);
  return response.data;
};

export const deleteResource = async (id: string): Promise<void> => {
  await api.delete(`/resources/${id}`);
};

export const saveResource = async (data: SaveDocumentRequest): Promise<EventSource> => {
  await api.post('/resources', data);
  
  const eventSourceUrl = `${window.location.origin}${API_BASE_URL}/resources`;
  
  // Add authentication to EventSource if available
  const eventSourceOptions: EventSourceInit = {
    withCredentials: true,
  };
  
  const eventSource = new EventSource(eventSourceUrl, eventSourceOptions);
  
  return eventSource;
};

export const askQuestion = async (data: AskRequest): Promise<AskResponse> => {
  const response = await api.post('/ask', data);
  return response.data;
};

export const streamAnswer = (question: string): EventSource => {
  const url = new URL(`${window.location.origin}${API_BASE_URL}/ask/stream`);

  url.searchParams.append('question', question);

  // Create EventSource with proper authorization if available
  const eventSourceOptions: EventSourceInit = {
    withCredentials: true,
  };
  
  const eventSource = new EventSource(url.toString(), eventSourceOptions);
  
  return eventSource;
};

export const cancelStream = async (processId: string): Promise<void> => {
  await api.delete(`/ask/stream/cancel/${processId}`);
}; 